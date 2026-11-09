package theme

import (
	"image/color"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
)

// NoColorFromEnv reports whether the NO_COLOR convention asks for colourless output.
func NoColorFromEnv() bool { return os.Getenv("NO_COLOR") != "" }

// Styles holds the lipgloss styles every component renders with.
type Styles struct {
	App, Title, Panel, PanelFocused, PanelTitle, Muted lipgloss.Style
	FooterKey, FooterDesc, Status, StatusWarning       lipgloss.Style
	StatusError, Modal, ModalTitle, Selected           lipgloss.Style

	edge, edgeFocused lipgloss.Style
	thickFocus        bool
}

// NewStyles builds the styles of t; with noColor no colour is used at all: focus is shown by a
// thick border and selection by reverse video.
func NewStyles(t Theme, noColor bool) Styles {
	plain := lipgloss.NewStyle()
	fg := func(token string) lipgloss.Style {
		if noColor {
			return plain
		}
		return plain.Foreground(c(t, token))
	}
	rounded, thick := lipgloss.RoundedBorder(), lipgloss.ThickBorder()
	s := Styles{
		App:           plain,
		Title:         fg("accent").Bold(true),
		PanelTitle:    fg("fg").Bold(true),
		Muted:         fg("muted"),
		FooterKey:     fg("accent"),
		FooterDesc:    fg("muted"),
		Status:        fg("fg"),
		StatusWarning: fg("warning"),
		StatusError:   fg("error"),
		ModalTitle:    fg("accent").Bold(true),
		Selected:      plain.Reverse(true),
		edge:          fg("border"),
		edgeFocused:   fg("focus"),
		thickFocus:    noColor,
	}
	s.Panel = fg("border").Border(rounded)
	s.PanelFocused = plain.Border(rounded)
	s.Modal = plain.Border(rounded)
	if noColor {
		s.PanelFocused = plain.Border(thick)
		s.Modal = plain.Border(thick)
		return s
	}
	s.App = plain.Background(c(t, "bg")).Foreground(c(t, "fg"))
	s.Panel = plain.Border(rounded).BorderForeground(c(t, "border"))
	s.PanelFocused = plain.Border(rounded).BorderForeground(c(t, "focus"))
	s.Modal = plain.Border(rounded).BorderForeground(c(t, "focus"))
	s.Selected = plain.Background(c(t, "selection"))
	return s
}

func c(t Theme, token string) color.Color { return lipgloss.Color(t.Color(token)) }

// RenderPanel draws a bordered box of exactly width x height cells with title in the top border.
func (s Styles) RenderPanel(title, body string, width, height int, focused bool) string {
	if width < 2 || height < 2 {
		return ""
	}
	b, edge := lipgloss.RoundedBorder(), s.edge
	if focused {
		edge = s.edgeFocused
		if s.thickFocus {
			b = lipgloss.ThickBorder()
		}
	}
	inner := width - 2
	top := edge.Render(b.TopLeft)
	used := 0
	if title = fitWidth(title, inner-2); title != "" {
		top += edge.Render(b.Top) + s.PanelTitle.Render(title)
		used = lipgloss.Width(title) + 1
	}
	top += edge.Render(strings.Repeat(b.Top, inner-used) + b.TopRight)

	lines := strings.Split(body, "\n")
	out := make([]string, 0, height)
	out = append(out, top)
	left, right := edge.Render(b.Left), edge.Render(b.Right)
	cell := lipgloss.NewStyle().MaxWidth(inner)
	for i := range height - 2 {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		line = cell.Render(line)
		out = append(out, left+line+strings.Repeat(" ", inner-lipgloss.Width(line))+right)
	}
	out = append(out, edge.Render(b.BottomLeft+strings.Repeat(b.Bottom, inner)+b.BottomRight))
	return strings.Join(out, "\n")
}

// fitWidth trims s to at most n cells.
func fitWidth(s string, n int) string {
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r)) > n {
		r = r[:len(r)-1]
	}
	return string(r)
}
