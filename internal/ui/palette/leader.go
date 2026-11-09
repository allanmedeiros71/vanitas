// Package palette holds the which-key leader popup.
package palette

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/allanmedeiros71/vanitas/internal/keymap"
	"github.com/allanmedeiros71/vanitas/internal/ui/theme"
)

// LeaderKeys are the popup's bindings: Close plus the commands (leader scope except close).
type LeaderKeys struct {
	Close    key.Binding
	Commands []keymap.Binding
}

// LeaderSelectedMsg reports the chosen command's action id.
type LeaderSelectedMsg struct{ Action string }

// LeaderClosedMsg reports that the popup was dismissed without an action.
type LeaderClosedMsg struct{}

// Leader is the which-key popup opened by the leader key.
type Leader struct {
	keys LeaderKeys
	st   theme.Styles
}

// NewLeader builds the popup.
func NewLeader(keys LeaderKeys, st theme.Styles) Leader { return Leader{keys: keys, st: st} }

// Update runs the command bound to the pressed key; any other key closes the popup.
func (l Leader) Update(msg tea.Msg) (Leader, tea.Cmd) {
	k, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return l, nil
	}
	for _, c := range l.keys.Commands {
		if key.Matches(k, c.Binding) {
			a := c.Action
			return l, func() tea.Msg { return LeaderSelectedMsg{Action: a} }
		}
	}
	return l, func() tea.Msg { return LeaderClosedMsg{} }
}

func item(b key.Binding) string {
	h := b.Help()
	return h.Key + "  " + h.Desc
}

// Lines returns the popup rows as plain text: the title, the commands, then the close row.
func (l Leader) Lines() []string {
	out := []string{"Menu"}
	for _, c := range l.keys.Commands {
		out = append(out, item(c.Binding))
	}
	return append(out, item(l.keys.Close))
}

// View draws the bordered popup.
func (l Leader) View() string {
	lines := l.Lines()
	lines[0] = l.st.ModalTitle.Render(lines[0])
	return l.st.Modal.Padding(0, 1).Render(strings.Join(lines, "\n"))
}
