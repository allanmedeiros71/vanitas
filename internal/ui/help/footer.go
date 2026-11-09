// Package help renders the one-line shortcut footer and the searchable help modal. Both are fed
// only by key bindings, so the keymap registry stays the single source of shortcuts.
package help

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/lipgloss/v2"

	"github.com/allanmedeiros71/vanitas/internal/ui/theme"
)

const (
	maxFooterItems = 8
	footerSep      = "  "
)

// RenderFooter renders bindings as "<key> <desc>" items separated by two spaces, at most eight,
// within width display columns. An item that does not fit is dropped with all following ones, so
// the footer never shows a gap. Callers pass the footer-flagged bindings; no filtering is done.
func RenderFooter(bindings []key.Binding, width int, st theme.Styles) string {
	var b strings.Builder
	used, count := 0, 0
	for _, bd := range bindings {
		if count == maxFooterItems {
			break
		}
		h := bd.Help()
		plainW := lipgloss.Width(h.Key + " " + h.Desc)
		need := plainW
		if count > 0 {
			need += lipgloss.Width(footerSep)
		}
		if used+need > width {
			break
		}
		if count > 0 {
			b.WriteString(footerSep)
		}
		b.WriteString(st.FooterKey.Render(h.Key) + st.FooterDesc.Render(" "+h.Desc))
		used += need
		count++
	}
	return b.String()
}
