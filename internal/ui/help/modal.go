package help

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/allanmedeiros71/vanitas/internal/ui/theme"
)

// Section is a titled group of bindings in the help modal.
type Section struct {
	Title    string
	Bindings []key.Binding
}

// ModalKeys are the keymap bindings that drive the modal.
type ModalKeys struct{ Close, Up, Down key.Binding }

// ClosedMsg is produced by the command returned when the modal is closed.
type ClosedMsg struct{}

// backspace edits the filter as text input; it is not a keymap action and not rebindable.
var backspace = key.NewBinding(key.WithKeys("backspace"))

const noMatch = "No matching shortcuts"

type rowKind int

const (
	kindTitle rowKind = iota
	kindFilter
	kindHeading
	kindItem
	kindEmpty
)

type row struct {
	text string
	kind rowKind
}

// Modal is the searchable help overlay.
type Modal struct {
	title    string
	sections []Section
	keys     ModalKeys
	st       theme.Styles
	filter   string
	offset   int
}

// NewModal builds a modal titled "Help — <title>" listing sections.
func NewModal(title string, sections []Section, keys ModalKeys, st theme.Styles) Modal {
	return Modal{title: title, sections: sections, keys: keys, st: st}
}

// Update handles scrolling, closing and filter typing.
func (m Modal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(k, m.keys.Close):
		return m, func() tea.Msg { return ClosedMsg{} }
	case key.Matches(k, m.keys.Up):
		if m.offset > 0 {
			m.offset--
		}
	case key.Matches(k, m.keys.Down):
		if m.offset < len(m.rows())-1 {
			m.offset++
		}
	case key.Matches(k, backspace):
		if r := []rune(m.filter); len(r) > 0 {
			m.filter = string(r[:len(r)-1])
			m.offset = 0
		}
	case k.Text != "":
		m.filter += k.Text
		m.offset = 0
	}
	return m, nil
}

func (m Modal) rows() []row {
	rows := []row{{"Help — " + m.title, kindTitle}}
	f := "Filter: " + m.filter
	if m.filter == "" {
		f = "Filter: type to filter"
	}
	rows = append(rows, row{f, kindFilter})
	needle := strings.ToLower(m.filter)
	found := false
	for _, s := range m.sections {
		var items []row
		for _, b := range s.Bindings {
			h := b.Help()
			if needle != "" && !strings.Contains(strings.ToLower(h.Key), needle) &&
				!strings.Contains(strings.ToLower(h.Desc), needle) {
				continue
			}
			items = append(items, row{h.Key + "  " + h.Desc, kindItem})
		}
		if len(items) == 0 {
			continue
		}
		found = true
		rows = append(rows, row{s.Title, kindHeading})
		rows = append(rows, items...)
	}
	if !found {
		rows = append(rows, row{noMatch, kindEmpty})
	}
	return rows
}

// Lines returns the full unscrolled content rows as plain text.
func (m Modal) Lines() []string {
	rows := m.rows()
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = r.text
	}
	return out
}

func (m Modal) style(r row) string {
	switch r.kind {
	case kindTitle:
		return m.st.ModalTitle.Render(r.text)
	case kindHeading:
		return m.st.Title.Render(r.text)
	case kindFilter:
		if m.filter == "" {
			return "Filter: " + m.st.Muted.Render("type to filter")
		}
	case kindEmpty:
		return m.st.Muted.Render(r.text)
	}
	return r.text
}

// View draws the modal centred in a width x height area, at most 70% wide and 80% tall.
func (m Modal) View(width, height int) string {
	if width < 4 || height < 4 {
		return ""
	}
	rows := m.rows()
	boxW, maxH := max(width*70/100, 4), max(height*80/100, 4)
	boxH := min(len(rows)+2, maxH)
	inner, innerH := boxW-2, boxH-2
	start := min(m.offset, max(0, len(rows)-innerH))
	lines := make([]string, 0, innerH)
	for _, r := range rows[start:min(start+innerH, len(rows))] {
		lines = append(lines, ansi.Truncate(m.style(r), inner, ""))
	}
	box := m.st.Modal.Width(boxW).Height(boxH).Render(strings.Join(lines, "\n"))
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
