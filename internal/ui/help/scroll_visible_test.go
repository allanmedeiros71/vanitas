package help

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

// Up from the bottom of a scrolled modal must move the visible window at once; the scroll offset
// must not be consumed silently while the view stays put.
func TestModalScrollUpVisibleFromBottom(t *testing.T) {
	t.Parallel()
	var bs []key.Binding
	for _, k := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"} {
		bs = append(bs, bind(k, "desc "+k))
	}
	m := NewModal("x", []Section{{"s", bs}}, modalKeys(), styles(t))
	for range len(m.Lines()) + 2 {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	before := ansi.Strip(m.View(40, 8))
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	if after := ansi.Strip(m.View(40, 8)); after == before {
		t.Fatalf("one Up from the bottom did not change the view:\n%s", after)
	}
}
