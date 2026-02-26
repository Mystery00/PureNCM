package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"

	"github.com/gen2brain/beeep"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"PureNCM/internal/config"
	"PureNCM/internal/ncm"
)

// App is the main application struct bound to the Wails frontend.
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct.
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if _, err := config.Load(); err != nil {
		_ = err // non-fatal, continue with defaults
	}
}

// --- Config API ---

func (a *App) GetConfig() *config.Config         { return config.Get() }
func (a *App) SetOutputDir(dir string) error     { return config.SetOutputDir(dir) }
func (a *App) SetFilenamePattern(p string) error { return config.SetFilenamePattern(p) }
func (a *App) SetCopyLrc(enabled bool) error     { return config.SetCopyLrc(enabled) }

// --- Dialog API ---

func (a *App) OpenFileDialog() ([]string, error) {
	return wailsRuntime.OpenMultipleFilesDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title:   "选择 NCM 文件",
		Filters: []wailsRuntime.FileFilter{{DisplayName: "NCM 加密音乐 (*.ncm)", Pattern: "*.ncm"}},
	})
}

func (a *App) OpenDirectoryDialog() (string, error) {
	return wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title:            "选择输出目录",
		DefaultDirectory: defaultOutputDir(),
	})
}

// --- Conversion API ---

// ConvertProgress is the event payload emitted for each file during conversion.
type ConvertProgress struct {
	Path       string  `json:"path"`
	Status     string  `json:"status"`   // "converting" | "done" | "error"
	Size       int64   `json:"size"`     // source file size in bytes
	Progress   float64 `json:"progress"` // 0.0 – 1.0 write progress
	OutputPath string  `json:"outputPath"`
	Error      string  `json:"error"`
}

const EventConvertProgress = "ncm:progress"

// maxWorkers limits concurrent goroutines to the number of logical CPUs.
var maxWorkers = runtime.NumCPU()

// ConvertFiles converts a list of NCM files concurrently.
// Each file is processed in its own goroutine (pool size = NumCPU).
// Progress is streamed via "ncm:progress" Wails events.
// A system notification is shown when all files are done.
func (a *App) ConvertFiles(paths []string, outputDir string, pattern string) {
	total := len(paths)
	if total == 0 {
		return
	}

	var (
		wg     sync.WaitGroup
		sem    = make(chan struct{}, maxWorkers) // semaphore
		doneN  atomic.Int32                      // count successfully converted
		errorN atomic.Int32
	)

	for _, p := range paths {
		p := p // capture loop variable
		wg.Add(1)
		sem <- struct{}{} // acquire slot

		go func() {
			defer wg.Done()
			defer func() { <-sem }() // release slot

			a.convertOne(p, outputDir, pattern, &doneN, &errorN)
		}()
	}

	// Wait for all goroutines, then send a system notification
	go func() {
		wg.Wait()
		d := int(doneN.Load())
		e := int(errorN.Load())
		sendNotification(d, e, total)
	}()
}

// convertOne processes a single NCM file and emits progress events.
func (a *App) convertOne(p, outputDir, pattern string, doneN, errorN *atomic.Int32) {
	emit := func(ev ConvertProgress) {
		wailsRuntime.EventsEmit(a.ctx, EventConvertProgress, ev)
	}

	// Read file size
	var fileSize int64
	if fi, err := os.Stat(p); err == nil {
		fileSize = fi.Size()
	}

	// Emit "converting" immediately so frontend shows the row as active
	emit(ConvertProgress{Path: p, Status: "converting", Size: fileSize, Progress: 0})

	// Determine output dir
	outDir := outputDir
	if outDir == "" {
		outDir = filepath.Dir(p)
	}
	if err := os.MkdirAll(outDir, 0755); err != nil {
		emit(ConvertProgress{Path: p, Status: "error", Error: "无法创建输出目录: " + err.Error()})
		errorN.Add(1)
		return
	}

	// Decrypt
	result, err := ncm.DecryptFile(p)
	if err != nil {
		emit(ConvertProgress{Path: p, Status: "error", Error: err.Error()})
		errorN.Add(1)
		return
	}

	// Write with progress callback — throttle to avoid flooding the event bus
	// (emit at most once per ~1% change using a simple threshold)
	var lastPct float64
	progressFn := func(pct float64) {
		if pct-lastPct >= 0.01 || pct >= 1.0 {
			lastPct = pct
			emit(ConvertProgress{
				Path:     p,
				Status:   "converting",
				Size:     fileSize,
				Progress: pct,
			})
		}
	}

	outPath, err := ncm.WriteToFileWithProgress(result, outDir, pattern, progressFn)
	if err != nil {
		emit(ConvertProgress{Path: p, Status: "error", Error: err.Error()})
		errorN.Add(1)
		return
	}

	emit(ConvertProgress{Path: p, Status: "done", Size: fileSize, Progress: 1.0, OutputPath: outPath})
	tryLrcCopy(p, outputDir) // copy .lrc sidecar if feature is enabled
	doneN.Add(1)
}

// sendNotification shows a system toast/notification when conversion finishes.
func sendNotification(done, errors, total int) {
	title := "PureNCM — 转换完成"
	var msg string
	switch {
	case errors == 0:
		msg = fmt.Sprintf("成功转换 %d 个文件", done)
	case done == 0:
		msg = fmt.Sprintf("全部 %d 个文件转换失败", total)
	default:
		msg = fmt.Sprintf("完成 %d / %d，失败 %d 个", done, total, errors)
	}
	_ = beeep.Notify(title, msg, "")
}

// tryLrcCopy copies a .lrc sidecar next to the source .ncm into the output directory.
// It is a no-op when: config.CopyLrc is false, outputDir is empty,
// or there is no matching .lrc file.
func tryLrcCopy(srcNCM, outputDir string) {
	if outputDir == "" {
		return // no explicit output dir — lrc would stay next to source anyway
	}
	if !config.Get().CopyLrc {
		return
	}
	// Look for a .lrc file with the same base name as the ncm
	lrcSrc := srcNCM[:len(srcNCM)-len(".ncm")] + ".lrc"
	if _, err := os.Stat(lrcSrc); err != nil {
		return // no .lrc found — skip silently
	}
	lrcDst := filepath.Join(outputDir, filepath.Base(lrcSrc))
	src, err := os.Open(lrcSrc)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.Create(lrcDst)
	if err != nil {
		return
	}
	defer dst.Close()
	_, _ = io.Copy(dst, src)
}

// --- Helpers ---

func defaultOutputDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	music := filepath.Join(home, "Music")
	if _, err := os.Stat(music); err == nil {
		return music
	}
	return home
}
