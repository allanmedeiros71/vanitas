package cli

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/allanmedeiros71/vanitas/internal/app"
	"github.com/allanmedeiros71/vanitas/internal/config"
	"github.com/allanmedeiros71/vanitas/internal/keymap"
	"github.com/allanmedeiros71/vanitas/internal/ui/theme"
)

const defaultTheme = "vanitas"

// Startup resolves the configuration paths, sets up logging and loads settings, keymap and theme.
// The returned function releases the log file. A settings or keymap syntax error is returned as is
// and nothing is written to the configuration directory.
func Startup() (app.Options, func(), error) {
	paths, err := config.ResolvePaths()
	if err != nil {
		return app.Options{}, nil, err
	}
	return startupIn(paths, os.Getenv("VANITAS_LOG"), theme.NoColorFromEnv())
}

func startupIn(paths config.Paths, logLevel string, noColor bool) (app.Options, func(), error) {
	closeLog, err := setupLogging(paths, logLevel)
	if err != nil {
		return app.Options{}, nil, err
	}
	opts, err := loadOptions(paths, noColor)
	if err != nil {
		closeLog()
		return app.Options{}, nil, err
	}
	for _, w := range opts.Warnings {
		slog.Warn("startup warning", "warning", w)
	}
	return opts, closeLog, nil
}

func loadOptions(paths config.Paths, noColor bool) (app.Options, error) {
	settings, warnings, err := config.LoadSettings(paths.Settings)
	if err != nil {
		return app.Options{}, err
	}
	overrides, kw, err := keymap.LoadOverrides(paths.Keymap)
	if err != nil {
		return app.Options{}, err
	}
	reg := keymap.Defaults()
	warnings = append(warnings, kw...)
	warnings = append(warnings, reg.Apply(overrides)...)

	th, tw, err := theme.Load(settings.Theme, settings.ThemeOverrides)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("unknown theme %q, using %q", settings.Theme, defaultTheme))
		if th, tw, err = theme.Load(defaultTheme, settings.ThemeOverrides); err != nil {
			return app.Options{}, err
		}
	}
	warnings = append(warnings, tw...)
	return app.Options{
		Paths: paths, Settings: settings, Registry: reg, Theme: th, Warnings: warnings, NoColor: noColor,
	}, nil
}

// setupLogging appends debug logs to the log file when level is "debug"; otherwise logging is
// discarded and no file is created.
func setupLogging(paths config.Paths, level string) (func(), error) {
	if level != "debug" {
		slog.SetDefault(slog.New(slog.DiscardHandler))
		return func() {}, nil
	}
	if err := os.MkdirAll(paths.Dir, 0o700); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}
	f, err := os.OpenFile(filepath.Clean(paths.Log), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(f, &slog.HandlerOptions{Level: slog.LevelDebug})))
	return func() { closeQuietly(f) }, nil
}

func closeQuietly(c io.Closer) { _ = c.Close() }
