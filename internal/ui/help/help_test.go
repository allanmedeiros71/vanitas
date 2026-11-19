package help

import (
	"slices"
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/allanmedeiros71/vanitas/internal/keymap"
	"github.com/allanmedeiros71/vanitas/internal/ui/theme"
)

func styles(t *testing.T) theme.Styles {
	t.Helper()
	th, _, err := theme.Load(theme.Names()[0], nil)
	if err != nil {
		t.Fatal(err)
	}
	return theme.NewStyles(th, false)
}

func bind(k, desc string) key.Binding {
	return key.NewBinding(key.WithKeys(k), key.WithHelp(keymap.DisplayKey(k), desc))
}

func modalKeys() ModalKeys {
	return ModalKeys{Close: bind("esc", "close"), Up: bind("up", "up"), Down: bind("down", "down")}
}

func typeText(m Modal, s string) Modal {
	for _, r := range s {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

func TestFooter(t *testing.T) {
	t.Parallel()
	st := styles(t)
	bs := []key.Binding{bind("alt+1", "explorer"), bind("alt+2", "editor"), bind("?", "help"), bind("space", "menu")}
	tests := []struct {
		name  string
		bs    []key.Binding
		width int
		want  string
	}{
		{"all", bs, 120, "alt+1 explorer  alt+2 editor  ? help  ␣ menu"},
		{"narrow", bs, 30, "alt+1 explorer  alt+2 editor"},
		{"exact", bs, 28, "alt+1 explorer  alt+2 editor"},
		{"one short", bs, 27, "alt+1 explorer"},
		{"none", nil, 80, ""},
		{"wide desc", []key.Binding{bind("a", "日本語"), bind("b", "x")}, 10, "a 日本語"},
		{"drops following", []key.Binding{bind("a", "aaaaaaaa"), bind("b", "x")}, 5, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ansi.Strip(RenderFooter(tc.bs, tc.width, st))
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
			if lipgloss.Width(got) > tc.width {
				t.Fatalf("width %d > %d", lipgloss.Width(got), tc.width)
			}
		})
	}
}

func TestFooterMaxEight(t *testing.T) {
	t.Parallel()
	var bs []key.Binding
	for _, k := range []string{"f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10"} {
		bs = append(bs, bind(k, "d"))
	}
	got := ansi.Strip(RenderFooter(bs, 200, styles(t)))
	if n := len(strings.Split(got, "  ")); n != 8 || strings.Contains(got, "f9") {
		t.Fatalf("got %q", got)
	}
}

func TestModalFilterAndClose(t *testing.T) {
	t.Parallel()
	g := keymap.Defaults().Bindings(keymap.ScopeGlobal)
	bs := make([]key.Binding, len(g))
	for i := range g {
		bs[i] = g[i].Binding
	}
	m := NewModal("global", []Section{{"global", bs}}, modalKeys(), styles(t))
	if l := m.Lines(); l[0] != "Help — global" || !strings.HasPrefix(l[1], "Filter:") {
		t.Fatalf("lines %q", l)
	}
	m = typeText(m, "QUIT")
	l := m.Lines()
	if !slices.Contains(l, "ctrl+q  quit") || slices.Contains(l, "?  help") || l[1] != "Filter: QUIT" {
		t.Fatalf("lines %q", l)
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	if m.Lines()[1] != "Filter: QUI" {
		t.Fatalf("backspace: %q", m.Lines()[1])
	}
	m = typeText(m, "zzz")
	if l := m.Lines(); l[len(l)-1] != "No matching shortcuts" {
		t.Fatalf("lines %q", l)
	}
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("no close cmd")
	}
	if _, ok := cmd().(ClosedMsg); !ok {
		t.Fatal("not ClosedMsg")
	}
	if _, cmd := m.Update(tea.WindowSizeMsg{}); cmd != nil {
		t.Fatal("unexpected cmd")
	}
}

func TestModalBackspaceEmptyNoop(t *testing.T) {
	t.Parallel()
	m := NewModal("x", []Section{{"s", []key.Binding{bind("a", "b")}}}, modalKeys(), styles(t))
	before := m.Lines()
	m, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyBackspace})
	if cmd != nil || !slices.Equal(before, m.Lines()) {
		t.Fatal("backspace changed state")
	}
}

func TestModalScrollBounds(t *testing.T) {
	t.Parallel()
	var bs []key.Binding
	for _, k := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"} {
		bs = append(bs, bind(k, "desc "+k))
	}
	m := NewModal("x", []Section{{"s", bs}}, modalKeys(), styles(t))
	n := len(m.Lines())
	for range n + 5 {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	if m.offset != n-pinned-1 {
		t.Fatalf("offset %d, want %d", m.offset, n-pinned-1)
	}
	for range n + 5 {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	}
	if m.offset != 0 {
		t.Fatalf("offset %d", m.offset)
	}
	// title and filter stay pinned while binding rows scroll in a small box
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	v := ansi.Strip(m.View(40, 8))
	if !strings.Contains(v, "Help — x") || !strings.Contains(v, "Filter:") {
		t.Fatalf("title/filter not pinned after scroll:\n%s", v)
	}
	if strings.Contains(v, "│s ") {
		t.Fatalf("section heading still visible after scroll:\n%s", v)
	}
}

func TestModalViewFits(t *testing.T) {
	t.Parallel()
	reg := keymap.Defaults()
	var secs []Section
	for _, s := range reg.Scopes() {
		var bs []key.Binding
		for _, b := range reg.Bindings(s) {
			bs = append(bs, b.Binding)
		}
		secs = append(secs, Section{string(s), bs})
	}
	m := NewModal("all", secs, modalKeys(), styles(t))
	for _, sz := range [][2]int{{120, 40}, {80, 24}, {20, 6}} {
		v := m.View(sz[0], sz[1])
		if lipgloss.Width(v) > sz[0] || lipgloss.Height(v) > sz[1] {
			t.Fatalf("%v: got %dx%d", sz, lipgloss.Width(v), lipgloss.Height(v))
		}
		if !strings.Contains(ansi.Strip(v), "Help — all") {
			t.Fatalf("%v: no title", sz)
		}
	}
	if m.View(2, 2) != "" {
		t.Fatal("tiny view should be empty")
	}
}

func TestModalViewFilterStates(t *testing.T) {
	t.Parallel()
	m := NewModal("x", []Section{{"s", []key.Binding{bind("a", "b")}}}, modalKeys(), styles(t))
	if v := ansi.Strip(m.View(80, 24)); !strings.Contains(v, "type to filter") {
		t.Fatalf("no placeholder:\n%s", v)
	}
	m = typeText(m, "zz")
	if v := ansi.Strip(m.View(80, 24)); !strings.Contains(v, "No matching shortcuts") || !strings.Contains(v, "Filter: zz") {
		t.Fatalf("view:\n%s", v)
	}
}

// TestRegistryWalk checks that every default binding is listed in the help of its scope, global
// ones after the "Global" heading. It reads the registry, so later phases are covered too.
func TestRegistryWalk(t *testing.T) {
	t.Parallel()
	reg := keymap.Defaults()
	toKeys := func(s keymap.Scope) []key.Binding {
		var out []key.Binding
		for _, b := range reg.Bindings(s) {
			out = append(out, b.Binding)
		}
		return out
	}
	for _, s := range reg.Scopes() {
		secs := []Section{{string(s), toKeys(s)}}
		if s != keymap.ScopeGlobal {
			secs = append(secs, Section{"Global", toKeys(keymap.ScopeGlobal)})
		}
		lines := NewModal(string(s), secs, modalKeys(), styles(t)).Lines()
		from := 0
		for _, sec := range secs {
			idx := slices.Index(lines[from:], sec.Title)
			if idx < 0 {
				t.Fatalf("%s: heading %q missing", s, sec.Title)
			}
			from += idx
			for _, b := range sec.Bindings {
				h := b.Help()
				i := slices.Index(lines[from:], h.Key+"  "+h.Desc)
				if i < 0 {
					t.Fatalf("%s: %q missing under %q", s, h.Key+"  "+h.Desc, sec.Title)
				}
			}
		}
	}
}
