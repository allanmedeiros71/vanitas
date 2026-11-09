// Package config resolves Vanitas' configuration files and loads its settings.
package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// Paths holds the location of every file Vanitas keeps in its configuration directory.
type Paths struct {
	Dir, Settings, Keymap, Connections, State, Log string
}

// PathsIn returns the file paths inside dir.
func PathsIn(dir string) Paths {
	return Paths{
		Dir:         dir,
		Settings:    filepath.Join(dir, "settings.toml"),
		Keymap:      filepath.Join(dir, "keymap.toml"),
		Connections: filepath.Join(dir, "connections.toml"),
		State:       filepath.Join(dir, "state.db"),
		Log:         filepath.Join(dir, "vanitas.log"),
	}
}

// ResolvePaths locates the configuration directory for the running OS. It creates nothing.
func ResolvePaths() (Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		home = ""
	}
	dir, err := resolveDir(runtime.GOOS, os.Getenv, home)
	if err != nil {
		return Paths{}, err
	}
	return PathsIn(dir), nil
}

// resolveDir is pure so every OS branch can be tested on any platform.
func resolveDir(goos string, getenv func(string) string, home string) (string, error) {
	if d := getenv("VANITAS_CONFIG_DIR"); d != "" {
		return filepath.Clean(d), nil
	}
	if x := getenv("XDG_CONFIG_HOME"); x != "" && filepath.IsAbs(x) {
		return filepath.Join(x, "vanitas"), nil
	}
	if goos == "windows" {
		appData := getenv("APPDATA")
		if appData == "" {
			return "", errors.New("cannot locate configuration directory: APPDATA is empty")
		}
		return filepath.Join(appData, "vanitas"), nil
	}
	if home == "" {
		return "", errors.New("cannot locate configuration directory: home directory unknown")
	}
	if goos == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "vanitas"), nil
	}
	return filepath.Join(home, ".config", "vanitas"), nil
}
