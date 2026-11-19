package palette

import (
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/allanmedeiros71/vanitas/internal/keymap"
	"github.com/allanmedeiros71/vanitas/internal/ui/theme"
)

// The zero value of Leader (no constructor) must not panic in Update, View or Lines.
func TestLeaderZeroValueDoesNotPanic(t *testing.T) {
	t.Parallel()
	var l Leader
	_, _ = l.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	_, _ = l.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	_, _ = l.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	_ = l.View()
	_ = l.Lines()
}

// A zero keymap.Binding among the commands must not match any key and must not panic.
func TestLeaderZeroCommandBindingDoesNotPanic(t *testing.T) {
	t.Parallel()
	l := NewLeader(LeaderKeys{Commands: []keymap.Binding{{}}}, theme.Styles{})
	_, cmd := l.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	if cmd == nil || cmd() != (LeaderClosedMsg{}) {
		t.Fatalf("zero command binding should close the popup, got %#v", cmd)
	}
	_ = l.View()
}
