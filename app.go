package main

import "context"

// App is the main application struct.
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
}
