package acceptance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/allanmedeiros71/vanitas/internal/config"
)

type configWorld struct {
	restore  []func()
	paths    config.Paths
	dir      string
	settings config.Settings
	warnings []string
	loadErr  error
	loaded   bool
}

func (w *configWorld) reset() {
	for _, f := range w.restore {
		f()
	}
	*w = configWorld{}
}

func (w *configWorld) setEnv(name, value string) {
	old, had := os.LookupEnv(name)
	w.restore = append(w.restore, func() {
		if had {
			_ = os.Setenv(name, old)
		} else {
			_ = os.Unsetenv(name)
		}
	})
	_ = os.Setenv(name, value)
}

func (w *configWorld) resolve() error {
	p, err := config.ResolvePaths()
	w.paths = p
	return err
}

func eq(what, got, want string) error {
	if filepath.ToSlash(got) != want {
		return fmt.Errorf("%s is %q, want %q", what, got, want)
	}
	return nil
}

func (w *configWorld) emptyDir() error {
	d, err := os.MkdirTemp("", "vanitas-config-")
	if err != nil {
		return err
	}
	w.restore = append(w.restore, func() { _ = os.RemoveAll(d) })
	w.dir = d
	return nil
}

func (w *configWorld) settingsFile(doc *godog.DocString) error {
	if err := w.emptyDir(); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(w.dir, "settings.toml"), []byte(doc.Content+"\n"), 0o600)
}

func (w *configWorld) load() {
	w.settings, w.warnings, w.loadErr = config.LoadSettings(filepath.Join(w.dir, "settings.toml"))
	w.loaded = true
}

func (w *configWorld) loadOK() error {
	if w.loadErr != nil {
		return w.loadErr
	}
	if len(w.warnings) != 0 {
		return fmt.Errorf("unexpected warnings %v", w.warnings)
	}
	return nil
}

func (w *configWorld) loadWarning(want string) error {
	if w.loadErr != nil {
		return w.loadErr
	}
	for _, m := range w.warnings {
		if m == strings.ReplaceAll(want, `\"`, `"`) {
			return nil
		}
	}
	return fmt.Errorf("warnings %v do not include %q", w.warnings, want)
}

func (w *configWorld) loadFails(want string) error {
	if w.loadErr == nil {
		return fmt.Errorf("loading succeeded, want error containing %q", want)
	}
	if !strings.Contains(w.loadErr.Error(), want) {
		return fmt.Errorf("error %q does not contain %q", w.loadErr.Error(), want)
	}
	return nil
}

func (w *configWorld) setting(name, want string) error {
	var got string
	switch name {
	case "theme":
		got = w.settings.Theme
	case "page_size":
		got = fmt.Sprint(w.settings.PageSize)
	case "query_timeout":
		got = w.settings.QueryTimeout.String()
	case "vim_mode":
		got = fmt.Sprint(w.settings.VimMode)
	case "zebra":
		got = fmt.Sprint(w.settings.Zebra)
	case "fk_display":
		got = fmt.Sprint(w.settings.FKDisplay)
	default:
		return fmt.Errorf("unknown setting %q", name)
	}
	if got != want {
		return fmt.Errorf("setting %s is %q, want %q", name, got, want)
	}
	return nil
}

func (w *configWorld) dirNames() ([]string, error) {
	es, err := os.ReadDir(w.dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range es {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names, nil
}

func (w *configWorld) nothingCreated() error {
	names, err := w.dirNames()
	if err != nil {
		return err
	}
	if len(names) != 0 {
		return fmt.Errorf("directory contains %v", names)
	}
	return nil
}

func (w *configWorld) save(name, value string) error {
	s := config.DefaultSettings()
	if name != "theme" {
		return fmt.Errorf("unsupported setting %q", name)
	}
	s.Theme = value
	return config.SaveSettings(filepath.Join(w.dir, "settings.toml"), s)
}

func (w *configWorld) containsOnly(name string) error {
	names, err := w.dirNames()
	if err != nil {
		return err
	}
	if len(names) != 1 || names[0] != name {
		return fmt.Errorf("directory contains %v, want only %q", names, name)
	}
	return nil
}

func (w *configWorld) perms(name, want string) error {
	fi, err := os.Stat(filepath.Join(w.dir, name))
	if err != nil {
		return err
	}
	if got := fmt.Sprintf("%04o", fi.Mode().Perm()); got != want {
		return fmt.Errorf("permissions %s, want %s", got, want)
	}
	return nil
}

func TestConfigFeature(t *testing.T) {
	runFeature(t, "config", func(sc *godog.ScenarioContext) {
		w := &configWorld{}
		sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
			w.reset()
			return ctx, nil
		})
		sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
			w.reset()
			return ctx, nil
		})
		sc.Step(`^the environment variable "([^"]*)" is "([^"]*)"$`, w.setEnv)
		sc.Step(`^the configuration paths are resolved$`, w.resolve)
		sc.Step(`^the configuration directory is "([^"]*)"$`, func(s string) error { return eq("directory", w.paths.Dir, s) })
		sc.Step(`^the settings file is "([^"]*)"$`, func(s string) error { return eq("settings file", w.paths.Settings, s) })
		sc.Step(`^the keymap file is "([^"]*)"$`, func(s string) error { return eq("keymap file", w.paths.Keymap, s) })
		sc.Step(`^an empty configuration directory$`, w.emptyDir)
		sc.Step(`^a settings file containing:$`, w.settingsFile)
		sc.Step(`^the settings are loaded$`, w.load)
		sc.Step(`^loading succeeds with no warnings$`, w.loadOK)
		sc.Step(`^loading succeeds with the warning "((?:[^"\\]|\\.)*)"$`, w.loadWarning)
		sc.Step(`^loading fails with an error containing "([^"]*)"$`, w.loadFails)
		sc.Step(`^the error contains "([^"]*)"$`, w.loadFails)
		sc.Step(`^the setting "([^"]*)" is "([^"]*)"$`, w.setting)
		sc.Step(`^no file was created in the configuration directory$`, w.nothingCreated)
		sc.Step(`^the settings are saved with "([^"]*)" set to "([^"]*)"$`, w.save)
		sc.Step(`^the configuration directory contains only "([^"]*)"$`, w.containsOnly)
		sc.Step(`^the file "([^"]*)" has permissions "([^"]*)"$`, w.perms)
	})
}
