package app

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Overlays are drawn over the shell; the view must still fill the window exactly, with and
// without the explorer, at the minimum size and at larger sizes.
func TestOverlaysKeepWindowSize(t *testing.T) {
	t.Parallel()
	for _, sz := range [][2]int{{80, 24}, {120, 40}, {200, 60}} {
		for _, hidden := range []bool{false, true} {
			for _, overlay := range []string{"?", "space"} {
				m := sized(New(testOptions(t, "vanitas")), sz[0], sz[1])
				if hidden {
					m = presses(m, "space", "e")
				}
				m = presses(m, overlay)
				v := m.View().Content
				if w, h := lipgloss.Width(v), lipgloss.Height(v); w != sz[0] || h != sz[1] {
					t.Errorf("%dx%d hidden=%v overlay=%q: view is %dx%d", sz[0], sz[1], hidden, overlay, w, h)
				}
			}
		}
	}
}

// The leader popup is anchored bottom-right: its title and rows must appear in the right half of
// the body, and the title bar and status bar rows must stay untouched.
func TestLeaderPopupIsBottomRight(t *testing.T) {
	t.Parallel()
	m := presses(sized(New(testOptions(t, "vanitas")), 120, 40), "space")
	lines := strings.Split(ansi.Strip(m.View().Content), "\n")
	if len(lines) != 40 {
		t.Fatalf("got %d lines", len(lines))
	}
	if !strings.HasPrefix(lines[0], "vanitas ─ no connection") {
		t.Fatalf("title bar overdrawn: %q", lines[0])
	}
	found := -1
	for i, l := range lines {
		if idx := strings.Index(l, "Menu"); idx >= 0 {
			found = i
			if idx < 60 {
				t.Fatalf("popup title at column %d, want the right side: %q", idx, l)
			}
		}
	}
	if found < 20 || found > 37 {
		t.Fatalf("popup row %d, want it in the lower part of the body", found)
	}
}
