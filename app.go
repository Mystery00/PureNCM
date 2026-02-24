package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"

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

// GetConfig returns the current application configuration.
func (a *App) GetConfig() *config.Config {
	return config.Get()
}

// SetOutputDir updates and persists the output directory.
func (a *App) SetOutputDir(dir string) error {
	return config.SetOutputDir(dir)
}

// SetFilenamePattern updates and persists the filename format pattern.
func (a *App) SetFilenamePattern(pattern string) error {
	return config.SetFilenamePattern(pattern)
}

// --- Dialog API ---

// OpenFileDialog opens a native file picker filtered to .ncm files.
func (a *App) OpenFileDialog() ([]string, error) {
	return runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 NCM 文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "NCM 加密音乐 (*.ncm)", Pattern: "*.ncm"},
		},
	})
}

// OpenDirectoryDialog opens a native folder picker.
func (a *App) OpenDirectoryDialog() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "选择输出目录",
		DefaultDirectory: defaultOutputDir(),
	})
}

// --- Conversion API ---

// ConvertProgress is the payload emitted as a Wails event for each file.
type ConvertProgress struct {
	Path       string `json:"path"`       // original input path (used as key by frontend)
	Status     string `json:"status"`     // "converting" | "done" | "error"
	OutputPath string `json:"outputPath"` // set when status == "done"
	Error      string `json:"error"`      // set when status == "error"
}

const EventConvertProgress = "ncm:progress"

// ConvertFiles decrypts a list of NCM files sequentially.
// Progress is reported via the "ncm:progress" Wails event for each file.
// outputDir: destination directory; empty string = same dir as source.
// pattern:   filename pattern, e.g. "{title} - {artist}".
func (a *App) ConvertFiles(paths []string, outputDir string, pattern string) {
	for _, p := range paths {
		// Emit "converting" status
		runtime.EventsEmit(a.ctx, EventConvertProgress, ConvertProgress{
			Path:   p,
			Status: "converting",
		})

		// Determine output directory
		outDir := outputDir
		if outDir == "" {
			outDir = filepath.Dir(p)
		}
		if err := os.MkdirAll(outDir, 0755); err != nil {
			runtime.EventsEmit(a.ctx, EventConvertProgress, ConvertProgress{
				Path:   p,
				Status: "error",
				Error:  "无法创建输出目录: " + err.Error(),
			})
			continue
		}

		// Decrypt
		result, err := ncm.DecryptFile(p)
		if err != nil {
			runtime.EventsEmit(a.ctx, EventConvertProgress, ConvertProgress{
				Path:   p,
				Status: "error",
				Error:  err.Error(),
			})
			continue
		}

		// Write tagged file
		outPath, err := ncm.WriteToFile(result, outDir, pattern)
		if err != nil {
			runtime.EventsEmit(a.ctx, EventConvertProgress, ConvertProgress{
				Path:   p,
				Status: "error",
				Error:  err.Error(),
			})
			continue
		}

		runtime.EventsEmit(a.ctx, EventConvertProgress, ConvertProgress{
			Path:       p,
			Status:     "done",
			OutputPath: outPath,
		})
	}
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
