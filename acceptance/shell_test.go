package acceptance

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"
	"github.com/cucumber/godog"

	"github.com/allanmedeiros71/vanitas/internal/app"
	"github.com/allanmedeiros71/vanitas/internal/cli"
	"github.com/allanmedeiros71/vanitas/internal/config"
)

const goldenDir = "../testdata/golden"

// shellWorld drives app.Model directly (messages in, view out) so the screen is exact and
// independent of the host terminal; "vanitas exits" replays the keys in a teatest program.
type shellWorld struct {
	t         *testing.T
	dir       string
	width     int
	height    int
	model     app.Model
	opts      app.Options
	started   bool
	keys      []string
	quit      bool
	saved     map[string]*string
	stderr    bytes.Buffer
	exitCode  int
	closeLogs func()
}

func (w *shellWorld) setEnv(name, value string) {
	if w.saved == nil {
		w.saved = map[string]*string{}
	}
	if _, seen := w.saved[name]; !seen {
		if old, ok := os.LookupEnv(name); ok {
			w.saved[name] = &old
		} else {
			w.saved[name] = nil
		}
	}
	_ = os.Setenv(name, value)
}

func (w *shellWorld) restoreEnv() {
	for name, old := range w.saved {
		if old == nil {
			_ = os.Unsetenv(name)
		} else {
			_ = os.Setenv(name, *old)
		}
	}
	w.saved = nil
}

func (w *shellWorld) emptyDir() error {
	dir, err := os.MkdirTemp("", "vanitas-shell-")
	if err != nil {
		return err
	}
	w.dir = dir
	w.setEnv("VANITAS_CONFIG_DIR", dir)
	w.setEnv("VANITAS_LOG", "")
	w.setEnv("NO_COLOR", "")
	_ = os.Unsetenv("NO_COLOR")
	return nil
}

func (w *shellWorld) terminal(width, height int) { w.width, w.height = width, height }

func (w *shellWorld) writeConfig(name, content string) error {
	return os.WriteFile(filepath.Join(w.dir, name), []byte(content), 0o600)
}

func (w *shellWorld) settingsContains(content string) error {
	return w.writeConfig("settings.toml", content+"\n")
}

func (w *shellWorld) keymapContains(table, line string) error {
	return w.writeConfig("keymap.toml", table+"\n"+line+"\n")
}

func (w *shellWorld) envIs(name, value string) { w.setEnv(name, value) }

func (w *shellWorld) start() error {
	opts, closeLogs, err := cli.Startup()
	if err != nil {
		return err
	}
	w.opts, w.closeLogs = opts, closeLogs
	w.model = app.New(opts)
	w.started = true
	return w.send(tea.WindowSizeMsg{Width: w.width, Height: w.height})
}

func (w *shellWorld) startSized(width, height int) error {
	w.terminal(width, height)
	return w.start()
}

// send feeds msg to the model and runs the commands it returns, as a program would.
func (w *shellWorld) send(msg tea.Msg) error {
	next, cmd := w.model.Update(msg)
	m, ok := next.(app.Model)
	if !ok {
		return fmt.Errorf("model type %T", next)
	}
	w.model = m
	return w.run(cmd)
}

func (w *shellWorld) run(cmd tea.Cmd) error {
	if cmd == nil {
		return nil
	}
	switch msg := cmd().(type) {
	case nil:
		return nil
	case tea.QuitMsg:
		w.quit = true
		return nil
	case tea.BatchMsg:
		for _, c := range msg {
			if err := w.run(c); err != nil {
				return err
			}
		}
		return nil
	default:
		return w.send(msg)
	}
}

// parseKey turns "ctrl+q", "alt+3", "shift+tab", "space", "?" into a key press.
func parseKey(s string) (tea.KeyPressMsg, error) {
	parts := strings.Split(s, "+")
	base := parts[len(parts)-1]
	if s == "+" {
		base, parts = "+", nil
	}
	var k tea.Key
	for _, mod := range parts[:max(0, len(parts)-1)] {
		switch mod {
		case "ctrl":
			k.Mod |= tea.ModCtrl
		case "alt":
			k.Mod |= tea.ModAlt
		case "shift":
			k.Mod |= tea.ModShift
		default:
			return tea.KeyPressMsg{}, fmt.Errorf("unknown modifier in %q", s)
		}
	}
	named := map[string]rune{"tab": tea.KeyTab, "esc": tea.KeyEscape, "enter": tea.KeyEnter, "f1": tea.KeyF1}
	switch code, isNamed := named[base]; {
	case base == "space":
		k.Code, k.Text = tea.KeySpace, " "
	case isNamed:
		k.Code = code
	case len([]rune(base)) == 1:
		k.Code = []rune(base)[0]
		if k.Mod == 0 {
			k.Text = base
		}
	default:
		return tea.KeyPressMsg{}, fmt.Errorf("unknown key %q", s)
	}
	return tea.KeyPressMsg(k), nil
}

func (w *shellWorld) press(s string) error {
	k, err := parseKey(s)
	if err != nil {
		return err
	}
	w.keys = append(w.keys, s)
	return w.send(k)
}

func (w *shellWorld) resize(width, height int) error {
	w.terminal(width, height)
	return w.send(tea.WindowSizeMsg{Width: width, Height: height})
}

func (w *shellWorld) screen() string { return ansi.Strip(w.model.View().Content) }

func unescape(s string) string { return strings.ReplaceAll(s, `\"`, `"`) }

func (w *shellWorld) shows(text string) error {
	if s := w.screen(); !strings.Contains(s, unescape(text)) {
		return fmt.Errorf("screen does not show %q:\n%s", unescape(text), s)
	}
	return nil
}

func (w *shellWorld) notShows(text string) error {
	if s := w.screen(); strings.Contains(s, unescape(text)) {
		return fmt.Errorf("screen shows %q:\n%s", unescape(text), s)
	}
	return nil
}

func (w *shellWorld) showsPanels(a, b, c string) error {
	for _, p := range []string{a, b, c} {
		if err := w.shows(p); err != nil {
			return err
		}
	}
	return nil
}

func (w *shellWorld) focused(name string) error {
	if got := w.model.Focused(); got != name {
		return fmt.Errorf("focused panel %q, want %q", got, name)
	}
	return nil
}

// row returns the n-th line counted from the bottom of the screen (0 = last).
func (w *shellWorld) row(fromBottom int) string {
	lines := strings.Split(w.screen(), "\n")
	return lines[len(lines)-1-fromBottom]
}

func (w *shellWorld) footerContains(text string) error {
	if got := w.row(0); !strings.Contains(got, unescape(text)) {
		return fmt.Errorf("footer %q does not contain %q", got, unescape(text))
	}
	return nil
}

func (w *shellWorld) statusShows(text string) error {
	if got := w.row(1); !strings.Contains(got, unescape(text)) {
		return fmt.Errorf("status bar %q does not contain %q", got, unescape(text))
	}
	return nil
}

func (w *shellWorld) goldenMatches(name string) error {
	path := filepath.Join(goldenDir, name+".golden")
	got := w.screen()
	if f := flag.Lookup("update"); f != nil && f.Value.String() == "true" {
		if err := os.MkdirAll(goldenDir, 0o750); err != nil {
			return err
		}
		return os.WriteFile(path, []byte(got), 0o600)
	}
	want, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	if exp := strings.ReplaceAll(string(want), "\r\n", "\n"); exp != got {
		return fmt.Errorf("screen differs from %s:\n%s", path, got)
	}
	return nil
}

func (w *shellWorld) exits() error {
	if !w.quit {
		return errors.New("the model never asked to quit")
	}
	tm := teatest.NewTestModel(w.t, app.New(w.opts),
		teatest.WithInitialTermSize(w.width, w.height),
		teatest.WithProgramOptions(tea.WithColorProfile(colorprofile.Ascii)))
	for _, s := range w.keys {
		k, err := parseKey(s)
		if err != nil {
			return err
		}
		tm.Send(k)
	}
	timedOut := false
	tm.WaitFinished(w.t, teatest.WithFinalTimeout(2*time.Second), teatest.WithTimeoutFn(func(testing.TB) { timedOut = true }))
	if timedOut {
		return errors.New("the program did not finish within 2s")
	}
	return nil
}

func (w *shellWorld) settingsHas(key, value string) error {
	s, _, err := config.LoadSettings(filepath.Join(w.dir, "settings.toml"))
	if err != nil {
		return err
	}
	if key != "theme" || s.Theme != value {
		return fmt.Errorf("settings %s = %q, want %q", key, s.Theme, value)
	}
	return nil
}

func (w *shellWorld) settingsStillContains(content string) error {
	got, err := os.ReadFile(filepath.Join(w.dir, "settings.toml"))
	if err != nil {
		return err
	}
	if !strings.Contains(string(got), content) {
		return fmt.Errorf("settings file is now %q", got)
	}
	return nil
}

func (w *shellWorld) runCommand(string) {
	cmd := cli.NewRootCommand(cli.BuildInfo{})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&w.stderr)
	cmd.SetArgs([]string{})
	w.exitCode = 0
	if err := cmd.Execute(); err != nil {
		w.exitCode = 1
	}
}

func (w *shellWorld) exitCodeIs(want int) error {
	if w.exitCode != want {
		return fmt.Errorf("exit code %d, want %d", w.exitCode, want)
	}
	return nil
}

func (w *shellWorld) errorContains(want string) error {
	if !strings.Contains(w.stderr.String(), want) {
		return fmt.Errorf("error output %q lacks %q", w.stderr.String(), want)
	}
	return nil
}

func (w *shellWorld) noFile(name string) error {
	if _, err := os.Stat(filepath.Join(w.dir, name)); err == nil {
		return fmt.Errorf("%s exists", name)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (w *shellWorld) cleanup() {
	if w.closeLogs != nil {
		w.closeLogs()
	}
	w.restoreEnv()
	if w.dir != "" {
		_ = os.RemoveAll(w.dir)
	}
}

func TestShellFeature(t *testing.T) {
	runFeature(t, "shell", func(sc *godog.ScenarioContext) {
		w := &shellWorld{t: t}
		sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
			*w = shellWorld{t: t}
			return ctx, nil
		})
		sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
			w.cleanup()
			return ctx, nil
		})
		text := `"((?:[^"\\]|\\.)*)"`
		sc.Step(`^an empty configuration directory$`, w.emptyDir)
		sc.Step(`^a terminal of (\d+)x(\d+)$`, w.terminal)
		sc.Step(`^vanitas starts$`, w.start)
		sc.Step(`^vanitas starts in a terminal of (\d+)x(\d+)$`, w.startSized)
		sc.Step(`^I press `+text+`$`, w.press)
		sc.Step(`^the terminal is resized to (\d+)x(\d+)$`, w.resize)
		sc.Step(`^the screen shows `+text+`$`, w.shows)
		sc.Step(`^the screen does not show `+text+`$`, w.notShows)
		sc.Step(`^the screen shows the panels `+text+`, `+text+` and `+text+`$`, w.showsPanels)
		sc.Step(`^the focused panel is `+text+`$`, w.focused)
		sc.Step(`^the footer contains `+text+`$`, w.footerContains)
		sc.Step(`^the status bar shows `+text+`$`, w.statusShows)
		sc.Step(`^the screen matches the golden file `+text+`$`, w.goldenMatches)
		sc.Step(`^vanitas exits$`, w.exits)
		sc.Step(`^the settings file has `+text+` set to `+text+`$`, w.settingsHas)
		sc.Step(`^the settings file contains '([^']*)'$`, w.settingsContains)
		sc.Step(`^the settings file contains "([^"]*)"$`, w.settingsContains)
		sc.Step(`^the settings file still contains "([^"]*)"$`, w.settingsStillContains)
		sc.Step(`^the keymap file contains "([^"]*)" and '([^']*)'$`, w.keymapContains)
		sc.Step(`^the environment variable "([^"]*)" is "([^"]*)"$`, w.envIs)
		sc.Step(`^I run "vanitas"$`, func() { w.runCommand("vanitas") })
		sc.Step(`^the exit code is (\d+)$`, w.exitCodeIs)
		sc.Step(`^the error output contains `+text+`$`, func(s string) error { return w.errorContains(unescape(s)) })
		sc.Step(`^the configuration directory has no `+text+`$`, w.noFile)
	})
}
