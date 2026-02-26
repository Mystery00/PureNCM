package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

const (
	// DefaultFilenamePattern is the default output filename format.
	// Supported placeholders: {title}, {artist}, {album}
	DefaultFilenamePattern = "{title}"
)

// Config holds all persisted application settings.
type Config struct {
	OutputDir       string `json:"outputDir"`
	FilenamePattern string `json:"filenamePattern"`
	CopyLrc         bool   `json:"copyLrc"` // copy .lrc sidecar to output dir after conversion
}

var (
	mu       sync.RWMutex
	instance *Config
	cfgPath  string
)

// Load reads the config file from disk (or returns defaults if missing).
// Must be called once at startup.
func Load() (*Config, error) {
	mu.Lock()
	defer mu.Unlock()

	path, err := configFilePath()
	if err != nil {
		return nil, err
	}
	cfgPath = path

	cfg := defaultConfig()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// First run — persist defaults immediately
		instance = cfg
		return cfg, save(cfg)
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	// Fill in fields that may be missing in older config files
	if cfg.FilenamePattern == "" {
		cfg.FilenamePattern = DefaultFilenamePattern
	}

	instance = cfg
	return cfg, nil
}

// Get returns the current in-memory config (must call Load first).
func Get() *Config {
	mu.RLock()
	defer mu.RUnlock()
	if instance == nil {
		return defaultConfig()
	}
	return instance
}

// SetOutputDir updates the output directory and persists the change.
func SetOutputDir(dir string) error {
	mu.Lock()
	defer mu.Unlock()
	if instance == nil {
		instance = defaultConfig()
	}
	instance.OutputDir = dir
	return save(instance)
}

// SetFilenamePattern updates the filename pattern and persists the change.
func SetFilenamePattern(pattern string) error {
	mu.Lock()
	defer mu.Unlock()
	if instance == nil {
		instance = defaultConfig()
	}
	if pattern == "" {
		pattern = DefaultFilenamePattern
	}
	instance.FilenamePattern = pattern
	return save(instance)
}

// SetCopyLrc sets whether to copy .lrc sidecar files after conversion.
func SetCopyLrc(enabled bool) error {
	mu.Lock()
	defer mu.Unlock()
	if instance == nil {
		instance = defaultConfig()
	}
	instance.CopyLrc = enabled
	return save(instance)
}

// save writes the config to disk. Caller must hold mu.
func save(cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cfgPath, data, 0644)
}

// configFilePath returns the platform-appropriate config file path.
// On Windows this is %AppData%\PureNCM\config.json.
func configFilePath() (string, error) {
	appData, err := os.UserConfigDir() // returns %AppData% on Windows
	if err != nil {
		return "", err
	}
	return filepath.Join(appData, "PureNCM", "config.json"), nil
}

func defaultConfig() *Config {
	return &Config{
		OutputDir:       "",
		FilenamePattern: DefaultFilenamePattern,
	}
}
