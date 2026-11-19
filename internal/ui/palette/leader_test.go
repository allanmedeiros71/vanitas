package palette

import (
	"slices"
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/allanmedeiros71/vanitas/internal/keymap"
	"github.com/allanmedeiros71/vanitas/internal/ui/theme"
)

func newLeader(t *testing.T) Leader {
	t.Helper()
	th, _, err := theme.Load(theme.Names()[0], nil)
	if err != nil {
		t.Fatal(err)
	}
	var cmds []keymap.Binding
	var closeB key.Binding
	for _, b := range keymap.Defaults().Bindings(keymap.ScopeLeader) {
		if b.Action == "close" {
			closeB = b.Binding
			continue
		}
		cmds = append(cmds, b)
	}
	return NewLeader(LeaderKeys{Close: closeB, Commands: cmds}, theme.NewStyles(th, false))
}

func TestLeaderLines(t *testing.T) {
	t.Parallel()
	got := newLeader(t).Lines()
	want := []string{"Menu", "e  toggle explorer", "t  switch theme", "q  quit", "esc  close"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %q", got)
	}
	v := ansi.Strip(newLeader(t).View())
	for _, w := range want {
		if !strings.Contains(v, w) {
			t.Fatalf("view lacks %q:\n%s", w, v)
		}
	}
}

func TestLeaderUpdate(t *testing.T) {
	t.Parallel()
	l := newLeader(t)
	tests := []struct {
		name string
		msg  tea.Msg
		want tea.Msg
	}{
		{"command", tea.KeyPressMsg{Code: 't', Text: "t"}, LeaderSelectedMsg{Action: "switch_theme"}},
		{"quit", tea.KeyPressMsg{Code: 'q', Text: "q"}, LeaderSelectedMsg{Action: "quit"}},
		{"close", tea.KeyPressMsg{Code: tea.KeyEscape}, LeaderClosedMsg{}},
		{"other", tea.KeyPressMsg{Code: 'z', Text: "z"}, LeaderClosedMsg{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, cmd := l.Update(tc.msg)
			if cmd == nil || cmd() != tc.want {
				t.Fatalf("cmd result mismatch, want %v", tc.want)
			}
		})
	}
	if _, cmd := l.Update(tea.WindowSizeMsg{}); cmd != nil {
		t.Fatal("non-key message produced a command")
	}
}
