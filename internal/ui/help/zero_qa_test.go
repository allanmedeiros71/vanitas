package help

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/allanmedeiros71/vanitas/internal/ui/theme"
)

// The zero value of Modal must not panic in any method.
func TestModalZeroValueAllMethods(t *testing.T) {
	t.Parallel()
	var m Modal
	for _, k := range []tea.KeyPressMsg{
		{Code: tea.KeyUp},
		{Code: tea.KeyDown},
		{Code: tea.KeyBackspace},
		{Code: 'a', Text: "a"},
		{Code: tea.KeyEscape},
	} {
		m, _ = m.Update(k)
	}
	_ = m.Lines()
	for _, sz := range [][2]int{{0, 0}, {3, 3}, {40, 8}, {120, 40}, {80, 24}} {
		_ = m.View(sz[0], sz[1])
	}
}

// RenderFooter with no bindings, and with a zero binding, on the zero theme.Styles must not panic.
func TestFooterZeroValueDoesNotPanic(t *testing.T) {
	t.Parallel()
	if got := RenderFooter(nil, 0, theme.Styles{}); got != "" {
		t.Fatalf("empty footer = %q, want \"\"", got)
	}
	_ = RenderFooter([]key.Binding{{}}, 80, theme.Styles{})
}
