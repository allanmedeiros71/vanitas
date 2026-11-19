package acceptance

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/allanmedeiros71/vanitas/internal/keymap"
)

type keymapWorld struct {
	dir      string
	registry *keymap.Registry
	warnings []string
	loadErr  error
}

func (w *keymapWorld) reset() {
	if w.dir != "" {
		_ = os.RemoveAll(w.dir)
	}
	*w = keymapWorld{}
}

func (w *keymapWorld) noFile() error {
	d, err := os.MkdirTemp("", "vanitas-keymap-")
	if err != nil {
		return err
	}
	w.dir = d
	return nil
}

func (w *keymapWorld) file(doc *godog.DocString) error {
	if err := w.noFile(); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(w.dir, "keymap.toml"), []byte(doc.Content+"\n"), 0o600)
}

// load mirrors what the app does at startup: defaults, then the user's overrides.
func (w *keymapWorld) load() {
	o, warnings, err := keymap.LoadOverrides(filepath.Join(w.dir, "keymap.toml"))
	w.loadErr = err
	if err != nil {
		return
	}
	w.registry = keymap.Defaults()
	w.warnings = append(warnings, w.registry.Apply(o)...)
}

func (w *keymapWorld) noWarnings() error {
	if w.loadErr != nil {
		return w.loadErr
	}
	if len(w.warnings) != 0 {
		return fmt.Errorf("unexpected warnings %q", w.warnings)
	}
	return nil
}

func (w *keymapWorld) warning(want string) error {
	if w.loadErr != nil {
		return w.loadErr
	}
	want = strings.ReplaceAll(want, `\"`, `"`)
	for _, m := range w.warnings {
		if m == want {
			return nil
		}
	}
	return fmt.Errorf("warnings %q do not include %q", w.warnings, want)
}

func (w *keymapWorld) runs(scope, key, action string) error {
	if w.loadErr != nil {
		return w.loadErr
	}
	got, ok := w.registry.Match(keymap.Scope(scope), key)
	if !ok || got != action {
		return fmt.Errorf("[%s] %q runs %q (bound: %v), want %q", scope, key, got, ok, action)
	}
	return nil
}

func (w *keymapWorld) runsNothing(scope, key string) error {
	if w.loadErr != nil {
		return w.loadErr
	}
	if got, ok := w.registry.Match(keymap.Scope(scope), key); ok {
		return fmt.Errorf("[%s] %q runs %q, want nothing", scope, key, got)
	}
	return nil
}

func (w *keymapWorld) notListed(scope, action string) error {
	if w.loadErr != nil {
		return w.loadErr
	}
	for _, b := range w.registry.Bindings(keymap.Scope(scope)) {
		if b.Action == action {
			return fmt.Errorf("[%s] lists %q with keys %v", scope, action, b.Keys())
		}
	}
	return nil
}

func (w *keymapWorld) fails(want string) error {
	if w.loadErr == nil {
		return fmt.Errorf("loading succeeded, want error containing %q", want)
	}
	if !strings.Contains(w.loadErr.Error(), want) {
		return fmt.Errorf("error %q does not contain %q", w.loadErr.Error(), want)
	}
	return nil
}

func TestKeymapFeature(t *testing.T) {
	runFeature(t, "keymap", func(sc *godog.ScenarioContext) {
		w := &keymapWorld{}
		sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
			w.reset()
			return ctx, nil
		})
		sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
			w.reset()
			return ctx, nil
		})
		sc.Step(`^no keymap file$`, w.noFile)
		sc.Step(`^a keymap file containing:$`, w.file)
		sc.Step(`^the keymap is loaded$`, w.load)
		sc.Step(`^there are no keymap warnings$`, w.noWarnings)
		sc.Step(`^there is the keymap warning "((?:[^"\\]|\\.)*)"$`, w.warning)
		sc.Step(`^in scope "([^"]*)" the key "([^"]*)" runs "([^"]*)"$`, w.runs)
		sc.Step(`^in scope "([^"]*)" the key "([^"]*)" runs nothing$`, w.runsNothing)
		sc.Step(`^scope "([^"]*)" does not list the action "([^"]*)"$`, w.notListed)
		sc.Step(`^loading the keymap fails with an error containing "([^"]*)"$`, w.fails)
	})
}
