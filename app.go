package main

import (
	"context"

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

// startup is called when the app starts. The context is saved
// so we can call the runtime methods later.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Load persisted config (creates defaults on first run)
	if _, err := config.Load(); err != nil {
		// Non-fatal: app continues with in-memory defaults
		// Error will surface when user tries to save settings
		_ = err
	}
}

// --- Config API (exposed to frontend via Wails bindings) ---

// GetConfig returns the current application configuration.
func (a *App) GetConfig() *config.Config {
	return config.Get()
}

// SetOutputDir updates and persists the output directory.
func (a *App) SetOutputDir(dir string) error {
	return config.SetOutputDir(dir)
}

// SetFilenamePattern updates and persists the filename format pattern.
// Supported placeholders: {title}, {artist}, {album}
func (a *App) SetFilenamePattern(pattern string) error {
	return config.SetFilenamePattern(pattern)
}
