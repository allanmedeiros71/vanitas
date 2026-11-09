package app

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/allanmedeiros71/vanitas/internal/config"
	"github.com/allanmedeiros71/vanitas/internal/keymap"
	"github.com/allanmedeiros71/vanitas/internal/ui/help"
	"github.com/allanmedeiros71/vanitas/internal/ui/palette"
	"github.com/allanmedeiros71/vanitas/internal/ui/theme"
)

// Action ids, as registered in internal/keymap.
const (
	actFocusExplorer  = "focus_explorer"
	actFocusEditor    = "focus_editor"
	actFocusGrid      = "focus_grid"
	actNextPanel      = "next_panel"
	actPrevPanel      = "prev_panel"
	actHelp           = "help"
	actLeader         = "leader"
	actQuit           = "quit"
	actToggleExplorer = "toggle_explorer"
	actSwitchTheme    = "switch_theme"
	actClose          = "close"
	actScrollUp       = "scroll_up"
	actScrollDown     = "scroll_down"
)

type handler func(Model) (Model, tea.Cmd)

// globalHandlers maps a global action id to what it does.
var globalHandlers = map[string]handler{
	actFocusExplorer: func(m Model) (Model, tea.Cmd) {
		m.explorerHidden = false
		m.focus = panelExplorer
		return m, nil
	},
	actFocusEditor: func(m Model) (Model, tea.Cmd) { m.focus = panelEditor; return m, nil },
	actFocusGrid:   func(m Model) (Model, tea.Cmd) { m.focus = panelResults; return m, nil },
	actNextPanel:   func(m Model) (Model, tea.Cmd) { return m.cycle(1), nil },
	actPrevPanel:   func(m Model) (Model, tea.Cmd) { return m.cycle(-1), nil },
	actHelp:        func(m Model) (Model, tea.Cmd) { return m.openHelp(), nil },
	actLeader:      func(m Model) (Model, tea.Cmd) { return m.openLeader(), nil },
	actQuit:        func(m Model) (Model, tea.Cmd) { return m, tea.Quit },
}

// leaderHandlers maps a leader action id to what it does.
var leaderHandlers = map[string]handler{
	actToggleExplorer: func(m Model) (Model, tea.Cmd) { return m.toggleExplorer(), nil },
	actSwitchTheme:    Model.switchTheme,
	actQuit:           func(m Model) (Model, tea.Cmd) { return m, tea.Quit },
}

func (m Model) registry() *keymap.Registry {
	if m.opts.Registry == nil {
		return keymap.Defaults()
	}
	return m.opts.Registry
}

func (m Model) focusScope() keymap.Scope {
	switch m.focus {
	case panelEditor:
		return keymap.ScopeEditorNormal
	case panelResults:
		return keymap.ScopeGrid
	}
	return keymap.ScopeExplorer
}

// onKey routes a key: below the minimum size only quit works; an open overlay owns the keyboard;
// otherwise the focused scope is tried before global.
func (m Model) onKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case m.tooSmall():
		return m.dispatch(msg, []keymap.Scope{keymap.ScopeGlobal}, map[string]handler{actQuit: globalHandlers[actQuit]})
	case m.overlay == overlayHelp:
		var cmd tea.Cmd
		m.helpModal, cmd = m.helpModal.Update(msg)
		return m, cmd
	case m.overlay == overlayLeader:
		var cmd tea.Cmd
		m.leader, cmd = m.leader.Update(msg)
		return m, cmd
	}
	return m.dispatch(msg, []keymap.Scope{m.focusScope(), keymap.ScopeGlobal}, globalHandlers)
}

func (m Model) dispatch(msg tea.KeyPressMsg, scopes []keymap.Scope, handlers map[string]handler) (tea.Model, tea.Cmd) {
	reg := m.registry()
	for _, s := range scopes {
		if id, ok := reg.Match(s, msg.String()); ok {
			if h, ok := handlers[id]; ok {
				return h(m)
			}
		}
	}
	return m, nil
}

func (m Model) runLeader(action string) (tea.Model, tea.Cmd) {
	if h, ok := leaderHandlers[action]; ok {
		return h(m)
	}
	return m, nil
}

// visible lists the panels that can take focus, in cycle order.
func (m Model) visible() []panel {
	if m.explorerHidden {
		return []panel{panelEditor, panelResults}
	}
	return []panel{panelExplorer, panelEditor, panelResults}
}

func (m Model) cycle(step int) Model {
	vis := m.visible()
	idx := 0
	for i, p := range vis {
		if p == m.focus {
			idx = i
		}
	}
	m.focus = vis[(idx+step+len(vis))%len(vis)]
	return m
}

func (m Model) toggleExplorer() Model {
	m.explorerHidden = !m.explorerHidden
	if m.explorerHidden && m.focus == panelExplorer {
		m.focus = panelEditor
	}
	return m
}

func toBindings(bs []keymap.Binding) []key.Binding {
	out := make([]key.Binding, len(bs))
	for i, b := range bs {
		out[i] = b.Binding
	}
	return out
}

func (m Model) openHelp() Model {
	reg := m.registry()
	var keys help.ModalKeys
	for _, b := range reg.Bindings(keymap.ScopeHelp) {
		switch b.Action {
		case actClose:
			keys.Close = b.Binding
		case actScrollUp:
			keys.Up = b.Binding
		case actScrollDown:
			keys.Down = b.Binding
		}
	}
	var sections []help.Section
	if own := reg.Bindings(m.focusScope()); len(own) > 0 {
		sections = append(sections, help.Section{Title: m.Focused(), Bindings: toBindings(own)})
	}
	sections = append(sections, help.Section{Title: "Global", Bindings: toBindings(reg.Bindings(keymap.ScopeGlobal))})
	m.helpModal = help.NewModal(m.Focused(), sections, keys, m.st)
	m.overlay = overlayHelp
	return m
}

func (m Model) openLeader() Model {
	var keys palette.LeaderKeys
	for _, b := range m.registry().Bindings(keymap.ScopeLeader) {
		if b.Action == actClose {
			keys.Close = b.Binding
			continue
		}
		keys.Commands = append(keys.Commands, b)
	}
	m.leader = palette.NewLeader(keys, m.st)
	m.overlay = overlayLeader
	return m
}

// settingsSavedMsg reports the result of saving settings.toml.
type settingsSavedMsg struct{ err error }

// switchTheme moves to the next built-in theme (wrapping), applies it at once and saves it.
func (m Model) switchTheme() (Model, tea.Cmd) {
	names := theme.Names()
	next := names[0]
	for i, n := range names {
		if n == m.opts.Theme.Name {
			next = names[(i+1)%len(names)]
		}
	}
	th, _, err := theme.Load(next, m.opts.Settings.ThemeOverrides)
	if err != nil {
		m.status, m.statusKind = err.Error(), statusError
		return m, nil
	}
	m.opts.Theme = th
	m.opts.Settings.Theme = next
	m.st = theme.NewStyles(th, m.opts.NoColor)
	m.status, m.statusKind = "theme: "+next, statusInfo
	return m, saveSettings(m.opts.Paths.Settings, m.opts.Settings)
}

// saveSettings writes settings.toml off the update loop.
func saveSettings(path string, s config.Settings) tea.Cmd {
	return func() tea.Msg { return settingsSavedMsg{err: config.SaveSettings(path, s)} }
}

func (m Model) onSaved(msg settingsSavedMsg) Model {
	if msg.err != nil {
		m.status, m.statusKind = fmt.Sprintf("could not save settings: %v", msg.err), statusError
	}
	return m
}
