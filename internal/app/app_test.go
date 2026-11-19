package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/teatest/v2"

	"github.com/allanmedeiros71/vanitas/internal/config"
	"github.com/allanmedeiros71/vanitas/internal/keymap"
	"github.com/allanmedeiros71/vanitas/internal/ui/theme"
)

func testOptions(t *testing.T, themeName string) Options {
	t.Helper()
	th, _, err := theme.Load(themeName, nil)
	if err != nil {
		t.Fatal(err)
	}
	s := config.DefaultSettings()
	s.Theme = themeName
	return Options{
		Paths:    config.PathsIn(t.TempDir()),
		Settings: s,
		Registry: keymap.Defaults(),
		Theme:    th,
	}
}

// send feeds msgs to m and runs the commands they return, as a program would.
func send(m Model, msgs ...tea.Msg) Model {
	for _, msg := range msgs {
		next, cmd := m.Update(msg)
		m = next.(Model)
		m = runCmd(m, cmd)
	}
	return m
}

func runCmd(m Model, cmd tea.Cmd) Model {
	if cmd == nil {
		return m
	}
	switch msg := cmd().(type) {
	case nil, tea.QuitMsg:
	case tea.BatchMsg:
		for _, c := range msg {
			m = runCmd(m, c)
		}
	default:
		m = send(m, msg)
	}
	return m
}

var namedKeys = map[string]tea.KeyPressMsg{
	"space":     {Code: tea.KeySpace, Text: " "},
	"tab":       {Code: tea.KeyTab},
	"shift+tab": {Code: tea.KeyTab, Mod: tea.ModShift},
	"esc":       {Code: tea.KeyEscape},
	"ctrl+q":    {Code: 'q', Mod: tea.ModCtrl},
	"alt+1":     {Code: '1', Mod: tea.ModAlt},
	"alt+2":     {Code: '2', Mod: tea.ModAlt},
	"alt+3":     {Code: '3', Mod: tea.ModAlt},
}

func press(s string) tea.KeyPressMsg {
	if k, ok := namedKeys[s]; ok {
		return k
	}
	return tea.KeyPressMsg{Code: []rune(s)[0], Text: s}
}

func presses(m Model, keys ...string) Model {
	for _, k := range keys {
		m = send(m, press(k))
	}
	return m
}

func sized(m Model, w, h int) Model { return send(m, tea.WindowSizeMsg{Width: w, Height: h}) }

func screen(m Model) string { return ansi.Strip(m.View().Content) }

func TestLayoutFillsWindow(t *testing.T) {
	t.Parallel()
	for _, sz := range [][2]int{{80, 24}, {120, 40}, {200, 60}} {
		for _, hidden := range []bool{false, true} {
			m := sized(New(testOptions(t, "vanitas")), sz[0], sz[1])
			if hidden {
				m = presses(m, "space", "e")
			}
			v := m.View().Content
			if w, h := lipgloss.Width(v), lipgloss.Height(v); w != sz[0] || h != sz[1] {
				t.Errorf("%dx%d hidden=%v: view is %dx%d", sz[0], sz[1], hidden, w, h)
			}
		}
	}
}

func TestLayoutGeometry(t *testing.T) {
	t.Parallel()
	g := layout(120, 40, true)
	if g.explorerW != 30 || g.mainW != 90 || g.bodyH != 37 || g.editorH+g.resultsH != 36 {
		t.Fatalf("geometry %+v", g)
	}
	if g = layout(80, 24, true); g.explorerW != 20 {
		t.Fatalf("explorer min width: %+v", g)
	}
	if g = layout(100, 30, false); g.explorerW != 0 || g.mainW != 100 {
		t.Fatalf("hidden explorer: %+v", g)
	}
}

func TestStartupScreen(t *testing.T) {
	t.Parallel()
	m := sized(New(testOptions(t, "vanitas")), 120, 40)
	s := screen(m)
	for _, want := range []string{
		"vanitas ─ no connection", "Explorer", "Editor", "Results", "No tabs open",
		"No connection", "No query tab open", "No results", "? help", "␣ menu",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("screen lacks %q", want)
		}
	}
	if m.Focused() != "Explorer" {
		t.Fatalf("focused %q", m.Focused())
	}
}

func TestFocusKeys(t *testing.T) {
	t.Parallel()
	m := sized(New(testOptions(t, "vanitas")), 120, 40)
	for key, want := range map[string]string{"alt+3": "Results", "alt+2": "Editor", "alt+1": "Explorer"} {
		if got := presses(m, key).Focused(); got != want {
			t.Errorf("%s: focused %q, want %q", key, got, want)
		}
	}
	if got := presses(m, "tab", "tab", "tab").Focused(); got != "Explorer" {
		t.Errorf("tab wraps to %q", got)
	}
	if got := presses(m, "shift+tab").Focused(); got != "Results" {
		t.Errorf("shift+tab from explorer gives %q", got)
	}
}

func TestCycleSkipsHiddenExplorer(t *testing.T) {
	t.Parallel()
	m := presses(sized(New(testOptions(t, "vanitas")), 120, 40), "space", "e")
	if m.Focused() != "Editor" {
		t.Fatalf("hiding moves focus to Editor, got %q", m.Focused())
	}
	if got := presses(m, "tab", "tab").Focused(); got != "Editor" {
		t.Errorf("tab cycle with hidden explorer reached %q", got)
	}
	if got := presses(m, "shift+tab").Focused(); got != "Results" {
		t.Errorf("shift+tab reached %q", got)
	}
	if s := screen(m); strings.Contains(s, "Explorer") {
		t.Error("explorer still drawn")
	}
	m = presses(m, "alt+1")
	if m.Focused() != "Explorer" || !strings.Contains(screen(m), "Explorer") {
		t.Error("alt+1 does not show the explorer")
	}
}

func TestHideExplorerKeepsOtherFocus(t *testing.T) {
	t.Parallel()
	m := presses(sized(New(testOptions(t, "vanitas")), 120, 40), "alt+3", "space", "e")
	if m.Focused() != "Results" {
		t.Fatalf("focused %q", m.Focused())
	}
}

func TestHelpModal(t *testing.T) {
	t.Parallel()
	m := presses(sized(New(testOptions(t, "vanitas")), 120, 40), "alt+3", "?")
	s := screen(m)
	if !strings.Contains(s, "Help — Results") || !strings.Contains(s, "Global") {
		t.Fatalf("help not shown:\n%s", s)
	}
	m = presses(m, "esc")
	if strings.Contains(screen(m), "Help — Results") || m.Focused() != "Results" {
		t.Fatal("esc did not close help or moved focus")
	}
}

func TestLeaderPopupCapturesKeys(t *testing.T) {
	t.Parallel()
	m := presses(sized(New(testOptions(t, "vanitas")), 120, 40), "space")
	if s := screen(m); !strings.Contains(s, "Menu") || !strings.Contains(s, "toggle explorer") {
		t.Fatalf("leader missing:\n%s", s)
	}
	m = presses(m, "?")
	if strings.Contains(screen(m), "Help —") {
		t.Fatal("? opened help over the leader popup")
	}
	if strings.Contains(screen(m), "toggle explorer") {
		t.Fatal("unknown key should close the popup")
	}
	m = presses(m, "space", "esc")
	if strings.Contains(screen(m), "toggle explorer") {
		t.Fatal("esc did not close the popup")
	}
}

func TestQuitKeys(t *testing.T) {
	t.Parallel()
	m := sized(New(testOptions(t, "vanitas")), 120, 40)
	_, cmd := m.Update(press("ctrl+q"))
	if cmd == nil || cmd() != tea.Msg(tea.QuitMsg{}) {
		t.Fatal("ctrl+q does not quit")
	}
	m = presses(m, "space")
	_, cmd = m.Update(press("q"))
	if cmd == nil {
		t.Fatal("no command from leader q")
	}
	sel := cmd()
	next, cmd := m.Update(sel)
	_ = next
	if cmd == nil || cmd() != tea.Msg(tea.QuitMsg{}) {
		t.Fatal("space q does not quit")
	}
}

func TestSwitchThemeWraps(t *testing.T) {
	t.Parallel()
	opts := testOptions(t, "dracula")
	m := presses(sized(New(opts), 120, 40), "space", "t")
	if !strings.Contains(screen(m), "theme: vanitas") {
		t.Fatalf("status lacks theme: vanitas")
	}
	s, _, err := config.LoadSettings(opts.Paths.Settings)
	if err != nil || s.Theme != "vanitas" {
		t.Fatalf("saved theme %q (%v)", s.Theme, err)
	}
}

func TestSwitchThemeSaveFailure(t *testing.T) {
	t.Parallel()
	opts := testOptions(t, "vanitas")
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	opts.Paths.Settings = filepath.Join(blocker, "settings.toml")
	m := presses(sized(New(opts), 120, 40), "space", "t")
	if s := screen(m); !strings.Contains(s, "could not save settings:") {
		t.Fatalf("status lacks save error:\n%s", s)
	}
	if m.statusKind != statusError {
		t.Fatal("save error is not styled as error")
	}
}

func TestWarningsInStatusBar(t *testing.T) {
	t.Parallel()
	opts := testOptions(t, "vanitas")
	opts.Warnings = []string{"first", "second"}
	if s := screen(sized(New(opts), 120, 40)); !strings.Contains(s, "first · second") {
		t.Fatal("warnings not joined")
	}
	opts.Warnings = []string{strings.Repeat("w", 200)}
	lines := strings.Split(screen(sized(New(opts), 120, 40)), "\n")
	if st := lines[len(lines)-2]; !strings.Contains(st, "…") || lipgloss.Width(st) != 120 {
		t.Fatalf("status not truncated to the width: %q", st)
	}
}

func TestTooSmallAndRestore(t *testing.T) {
	t.Parallel()
	m := presses(sized(New(testOptions(t, "vanitas")), 120, 40), "alt+3", "?")
	m = sized(m, 70, 20)
	s := screen(m)
	if !strings.Contains(s, "Terminal too small: 70x20 (minimum 80x24)") || strings.Contains(s, "Explorer") {
		t.Fatalf("small screen wrong:\n%s", s)
	}
	if m = presses(m, "esc", "tab"); m.Focused() != "Results" {
		t.Fatal("keys other than quit must be ignored when too small")
	}
	_, cmd := m.Update(press("ctrl+q"))
	if cmd == nil || cmd() != tea.Msg(tea.QuitMsg{}) {
		t.Fatal("ctrl+q must quit when too small")
	}
	m = sized(m, 120, 40)
	if s := screen(m); !strings.Contains(s, "Help — Results") {
		t.Fatalf("help modal not restored:\n%s", s)
	}
	if h := sized(New(testOptions(t, "vanitas")), 80, 23); !strings.Contains(screen(h), "80x23") {
		t.Fatal("height below 24 must warn")
	}
}

func TestZeroModelDoesNotPanic(t *testing.T) {
	t.Parallel()
	var m Model
	if v := m.View().Content; v != "" {
		t.Fatalf("zero model renders %q", v)
	}
	if m.Init() != nil {
		t.Fatal("unexpected init command")
	}
	m = presses(m, "?", "esc", "space", "esc", "tab", "alt+3")
	m = sized(m, 120, 40)
	m = presses(m, "?")
	if !strings.Contains(screen(m), "Help —") {
		t.Fatal("zero model falls back to the default keymap")
	}
	if m.Focused() == "" {
		t.Fatal("no focus name")
	}
}

func TestUserKeymapOverride(t *testing.T) {
	t.Parallel()
	opts := testOptions(t, "vanitas")
	opts.Registry.Apply(keymap.Overrides{keymap.ScopeGlobal: {"help": {"f1"}}})
	m := sized(New(opts), 120, 40)
	if !strings.Contains(screen(m), "f1 help") {
		t.Fatal("footer lacks f1 help")
	}
	m = send(m, tea.KeyPressMsg{Code: tea.KeyF1})
	if !strings.Contains(screen(m), "Help — Explorer") {
		t.Fatal("f1 does not open help")
	}
}

func TestProgramQuitsWithCtrlQ(t *testing.T) {
	t.Parallel()
	tm := teatest.NewTestModel(t, New(testOptions(t, "vanitas")),
		teatest.WithInitialTermSize(120, 40),
		teatest.WithProgramOptions(tea.WithColorProfile(colorprofile.Ascii)))
	tm.Send(press("space"))
	tm.Send(press("q"))
	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))
	if f := tm.FinalModel(t).(Model).Focused(); f != "Explorer" {
		t.Fatalf("final focus %q", f)
	}
}
