package help

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func qaTenBindings() []key.Binding {
	var bs []key.Binding
	for _, k := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"} {
		bs = append(bs, bind(k, "desc "+k))
	}
	return bs
}

// Every Up press that can move the view (offset above 0) changes the visible rows, while the
// title and filter rows stay pinned.
func TestModalUpChangesViewUntilTop(t *testing.T) {
	t.Parallel()
	m := NewModal("x", []Section{{"s", qaTenBindings()}}, modalKeys(), styles(t))
	for range 40 {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	prev := ansi.Strip(m.View(40, 8))
	for range 9 {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
		cur := ansi.Strip(m.View(40, 8))
		if cur == prev {
			t.Fatalf("Up at offset %d did not change the view:\n%s", m.offset, cur)
		}
		if !strings.Contains(cur, "Help — x") || !strings.Contains(cur, "Filter:") {
			t.Fatalf("title or filter not pinned:\n%s", cur)
		}
		prev = cur
	}
	if m.offset != 0 {
		t.Fatalf("offset %d after 9 Up presses, want 0", m.offset)
	}
}

// Down with no View before it must still end on the last row once rendered.
func TestModalDownWithoutViewShowsLastRow(t *testing.T) {
	t.Parallel()
	m := NewModal("x", []Section{{"s", qaTenBindings()}}, modalKeys(), styles(t))
	for range 40 {
		m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	}
	if v := ansi.Strip(m.View(40, 8)); !strings.Contains(v, "desc j") {
		t.Fatalf("last row not visible:\n%s", v)
	}
}

// The zero value of Modal is a usable value type: Update and View must not panic.
func TestModalZeroValueDoesNotPanic(t *testing.T) {
	t.Parallel()
	var m Modal
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	_ = m.View(40, 8)
}
