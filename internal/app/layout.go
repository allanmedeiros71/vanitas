package app

import (
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Minimum terminal size the shell lays out for.
const (
	minWidth  = 80
	minHeight = 24
)

// geometry is the cell budget of the shell regions for one window size.
type geometry struct {
	explorerW, mainW, bodyH, editorH, resultsH int
}

// layout splits a w x h window: one title row, the body, a status row and a footer row. The main
// area gives its first row to the tab bar line and splits the rest between Editor and Results.
func layout(w, h int, explorer bool) geometry {
	g := geometry{mainW: w, bodyH: max(0, h-3)}
	if explorer {
		g.explorerW = min(max(20, w/4), w)
		g.mainW = w - g.explorerW
	}
	rest := max(0, g.bodyH-1)
	g.editorH = rest / 2
	g.resultsH = rest - g.editorH
	return g
}

// fitLine truncates s to w cells and pads it with spaces to exactly w cells.
func fitLine(s string, w int) string {
	s = ansi.Truncate(s, w, "")
	return s + strings.Repeat(" ", max(0, w-lipgloss.Width(s)))
}

func (m Model) tooSmall() bool { return m.width < minWidth || m.height < minHeight }

// smallView is the whole screen when the terminal is below the minimum size.
func (m Model) smallView() string {
	text := "Terminal too small: " + strconv.Itoa(m.width) + "x" + strconv.Itoa(m.height) +
		" (minimum " + strconv.Itoa(minWidth) + "x" + strconv.Itoa(minHeight) + ")"
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, text)
}

// body draws the explorer and main area, or the open overlay over them.
func (m Model) body(g geometry) string {
	if m.overlay == overlayHelp {
		return m.helpModal.View(m.width, g.bodyH)
	}
	main := lipgloss.JoinVertical(lipgloss.Left,
		fitLine(m.st.Muted.Render("No tabs open"), g.mainW),
		m.st.RenderPanel("Editor", m.st.Muted.Render("No query tab open"), g.mainW, g.editorH, m.focus == panelEditor),
		m.st.RenderPanel("Results", m.st.Muted.Render("No results"), g.mainW, g.resultsH, m.focus == panelResults),
	)
	out := main
	if !m.explorerHidden {
		explorer := m.st.RenderPanel("Explorer", m.st.Muted.Render("No connection"), g.explorerW, g.bodyH, m.focus == panelExplorer)
		out = lipgloss.JoinHorizontal(lipgloss.Top, explorer, main)
	}
	if m.overlay == overlayLeader {
		out = overlayBottomRight(out, m.leader.View())
	}
	return out
}

// overlayBottomRight draws top over the bottom-right corner of base.
func overlayBottomRight(base, top string) string {
	lines := strings.Split(base, "\n")
	tl := strings.Split(top, "\n")
	tw := lipgloss.Width(top)
	x := max(0, lipgloss.Width(base)-tw)
	y := max(0, len(lines)-len(tl))
	for i, t := range tl {
		row := y + i
		if row >= len(lines) {
			break
		}
		lines[row] = ansi.Truncate(lines[row], x, "") + t + ansi.TruncateLeft(lines[row], x+tw, "")
	}
	return strings.Join(lines, "\n")
}
