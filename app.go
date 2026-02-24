package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"PureNCM/internal/config"
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
// Returns a list of selected file paths.
func (a *App) OpenFileDialog() ([]string, error) {
	return runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 NCM 文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "NCM 加密音乐 (*.ncm)", Pattern: "*.ncm"},
		},
	})
}

// OpenDirectoryDialog opens a native folder picker.
// Returns the selected directory path (empty string if cancelled).
func (a *App) OpenDirectoryDialog() (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "选择输出目录",
		DefaultDirectory: defaultOutputDir(),
	})
}

// defaultOutputDir returns the user's Music folder as the default pick location.
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
