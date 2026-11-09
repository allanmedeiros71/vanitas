// Package app is the root Bubble Tea model: it lays out the shell, routes keys through the keymap
// registry and owns the overlays, focus and theme.
package app

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/allanmedeiros71/vanitas/internal/config"
	"github.com/allanmedeiros71/vanitas/internal/keymap"
	"github.com/allanmedeiros71/vanitas/internal/ui/help"
	"github.com/allanmedeiros71/vanitas/internal/ui/palette"
	"github.com/allanmedeiros71/vanitas/internal/ui/theme"
)

// Options is everything the app needs, already loaded, so tests never touch the real home.
type Options struct {
	Paths    config.Paths
	Settings config.Settings
	Registry *keymap.Registry
	Theme    theme.Theme
	Warnings []string
	NoColor  bool
}

type panel int

const (
	panelExplorer panel = iota
	panelEditor
	panelResults
	panelCount
)

var panelNames = [panelCount]string{"Explorer", "Editor", "Results"}

type overlayKind int

const (
	overlayNone overlayKind = iota
	overlayHelp
	overlayLeader
)

type statusKind int

const (
	statusInfo statusKind = iota
	statusWarning
	statusError
)

// Model is the root model. Its zero value is safe to use: it renders nothing until it knows the
// window size and falls back to the default keymap.
type Model struct {
	opts           Options
	st             theme.Styles
	focus          panel
	explorerHidden bool
	overlay        overlayKind
	helpModal      help.Modal
	leader         palette.Leader
	width, height  int
	status         string
	statusKind     statusKind
}

// New builds the root model from loaded options.
func New(opts Options) Model {
	m := Model{opts: opts, st: theme.NewStyles(opts.Theme, opts.NoColor)}
	if len(opts.Warnings) > 0 {
		m.status = strings.Join(opts.Warnings, " · ")
		m.statusKind = statusWarning
	}
	return m
}

// Focused names the focused panel: Explorer, Editor or Results.
func (m Model) Focused() string { return panelNames[m.focus] }

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyPressMsg:
		return m.onKey(msg)
	case help.ClosedMsg, palette.LeaderClosedMsg:
		m.overlay = overlayNone
	case palette.LeaderSelectedMsg:
		m.overlay = overlayNone
		return m.runLeader(msg.Action)
	case settingsSavedMsg:
		m = m.onSaved(msg)
	}
	return m, nil
}

// View implements tea.Model.
func (m Model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	return v
}

func (m Model) render() string {
	if m.width == 0 && m.height == 0 {
		return ""
	}
	if m.tooSmall() {
		return m.smallView()
	}
	g := layout(m.width, m.height, !m.explorerHidden)
	title := m.st.Title.Render("vanitas ─ no connection")
	return strings.Join([]string{
		fitLine(title, m.width),
		m.body(g),
		m.statusBar(),
		fitLine(m.footer(), m.width),
	}, "\n")
}

func (m Model) statusBar() string {
	style := m.st.Status
	switch m.statusKind {
	case statusWarning:
		style = m.st.StatusWarning
	case statusError:
		style = m.st.StatusError
	}
	return fitLine(style.Render(ansi.Truncate(m.status, m.width, "…")), m.width)
}

// footer shows the footer-flagged bindings of the open overlay, or of the focused scope then global.
func (m Model) footer() string {
	reg := m.registry()
	var scopes []keymap.Scope
	switch m.overlay {
	case overlayHelp:
		scopes = []keymap.Scope{keymap.ScopeHelp}
	case overlayLeader:
		scopes = []keymap.Scope{keymap.ScopeLeader}
	default:
		scopes = []keymap.Scope{m.focusScope(), keymap.ScopeGlobal}
	}
	var bs []key.Binding
	for _, sc := range scopes {
		for _, b := range reg.Bindings(sc) {
			if b.Footer {
				bs = append(bs, b.Binding)
			}
		}
	}
	return help.RenderFooter(bs, m.width, m.st)
}
