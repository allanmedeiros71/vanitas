package acceptance

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/cucumber/godog"

	"github.com/allanmedeiros71/vanitas/internal/keymap"
	"github.com/allanmedeiros71/vanitas/internal/ui/help"
	"github.com/allanmedeiros71/vanitas/internal/ui/palette"
	"github.com/allanmedeiros71/vanitas/internal/ui/theme"
)

type helpWorld struct {
	st       theme.Styles
	bs       []key.Binding
	footer   string
	modal    help.Modal
	closed   bool
	leader   palette.Leader
	leadMsg  tea.Msg
	inLeader bool
}

func newHelpWorld() (*helpWorld, error) {
	th, _, err := theme.Load(theme.Names()[0], nil)
	if err != nil {
		return nil, err
	}
	return &helpWorld{st: theme.NewStyles(th, false)}, nil
}

func helpBinding(k, desc string) key.Binding {
	return key.NewBinding(key.WithKeys(k), key.WithHelp(keymap.DisplayKey(k), desc))
}

var quoted = regexp.MustCompile(`"([^"]*)"`)

func (w *helpWorld) bindingsList(list string) {
	w.bs = nil
	for _, m := range quoted.FindAllStringSubmatch(list, -1) {
		k, desc, _ := strings.Cut(m[1], " ")
		w.bs = append(w.bs, helpBinding(k, desc))
	}
}

func (w *helpWorld) manyBindings(n int) {
	w.bs = nil
	for i := 1; i <= n; i++ {
		w.bs = append(w.bs, helpBinding("f"+strconv.Itoa(i), "x"))
	}
}

func (w *helpWorld) renderFooter(width int) {
	w.footer = ansi.Strip(help.RenderFooter(w.bs, width, w.st))
}

func (w *helpWorld) footerIs(want string) error {
	if w.footer != want {
		return fmt.Errorf("footer = %q", w.footer)
	}
	return nil
}

func (w *helpWorld) footerShows(n int) error {
	if got := len(strings.Split(w.footer, "  ")); got != n {
		return fmt.Errorf("footer has %d items: %q", got, w.footer)
	}
	return nil
}

func (w *helpWorld) footerLacks(s string) error {
	if strings.Contains(w.footer, s) {
		return fmt.Errorf("footer contains %q", s)
	}
	return nil
}

func toKeyBindings(bs []keymap.Binding) []key.Binding {
	out := make([]key.Binding, len(bs))
	for i, b := range bs {
		out[i] = b.Binding
	}
	return out
}

func helpModalKeys(reg *keymap.Registry) help.ModalKeys {
	var k help.ModalKeys
	for _, b := range reg.Bindings(keymap.ScopeHelp) {
		switch b.Action {
		case "close":
			k.Close = b.Binding
		case "scroll_up":
			k.Up = b.Binding
		case "scroll_down":
			k.Down = b.Binding
		}
	}
	return k
}

func (w *helpWorld) walkScopes() error {
	reg := keymap.Defaults()
	for _, s := range reg.Scopes() {
		secs := []help.Section{{Title: string(s), Bindings: toKeyBindings(reg.Bindings(s))}}
		if s != keymap.ScopeGlobal {
			secs = append(secs, help.Section{Title: "Global", Bindings: toKeyBindings(reg.Bindings(keymap.ScopeGlobal))})
		}
		lines := help.NewModal(string(s), secs, helpModalKeys(reg), w.st).Lines()
		from := 0
		for _, sec := range secs {
			idx := slices.Index(lines[from:], sec.Title)
			if idx < 0 {
				return fmt.Errorf("scope %s: heading %q missing", s, sec.Title)
			}
			from += idx
			for _, b := range sec.Bindings {
				h := b.Help()
				if !slices.Contains(lines[from:], h.Key+"  "+h.Desc) {
					return fmt.Errorf("scope %s: %q missing under %q", s, h.Key, sec.Title)
				}
			}
		}
	}
	return nil
}

func (w *helpWorld) walkOK() error { return nil }

func (w *helpWorld) helpFor(scope string) {
	reg := keymap.Defaults()
	w.modal = help.NewModal(scope, []help.Section{{Title: scope, Bindings: toKeyBindings(reg.Bindings(keymap.Scope(scope)))}},
		helpModalKeys(reg), w.st)
}

func (w *helpWorld) typeText(s string) {
	for _, r := range s {
		w.modal, _ = w.modal.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
}

func pressMsg(name string) tea.KeyPressMsg {
	if name == "esc" {
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	}
	r := []rune(name)[0]
	return tea.KeyPressMsg{Code: r, Text: name}
}

func (w *helpWorld) pressInModal(name string) {
	var cmd tea.Cmd
	w.modal, cmd = w.modal.Update(pressMsg(name))
	if cmd != nil {
		_, w.closed = cmd().(help.ClosedMsg)
	}
}

func (w *helpWorld) helpLists(k, desc string) error {
	if !slices.Contains(w.modal.Lines()[2:], k+"  "+desc) {
		return fmt.Errorf("rows %q lack %q", w.modal.Lines(), k+"  "+desc)
	}
	return nil
}

func (w *helpWorld) helpLacks(desc string) error {
	for _, l := range w.modal.Lines()[2:] {
		if strings.HasSuffix(l, "  "+desc) {
			return fmt.Errorf("row %q present", l)
		}
	}
	return nil
}

func (w *helpWorld) helpShows(s string) error {
	if !slices.Contains(w.modal.Lines(), s) {
		return fmt.Errorf("rows %q lack %q", w.modal.Lines(), s)
	}
	return nil
}

func (w *helpWorld) modalClosed() error {
	if !w.closed {
		return fmt.Errorf("modal did not report closed")
	}
	return nil
}

func (w *helpWorld) leaderDefault() {
	var l palette.LeaderKeys
	for _, b := range keymap.Defaults().Bindings(keymap.ScopeLeader) {
		if b.Action == "close" {
			l.Close = b.Binding
			continue
		}
		l.Commands = append(l.Commands, b)
	}
	w.leader = palette.NewLeader(l, w.st)
	w.inLeader = true
}

func (w *helpWorld) pressInLeader(name string) {
	_, cmd := w.leader.Update(pressMsg(name))
	if cmd != nil {
		w.leadMsg = cmd()
	}
}

func (w *helpWorld) popupTitle(s string) error {
	if got := w.leader.Lines()[0]; got != s {
		return fmt.Errorf("title = %q", got)
	}
	return nil
}

func (w *helpWorld) popupLists(k, desc string) error {
	if !slices.Contains(w.leader.Lines(), k+"  "+desc) {
		return fmt.Errorf("rows %q lack %q", w.leader.Lines(), k+"  "+desc)
	}
	return nil
}

func (w *helpWorld) popupAction(a string) error {
	if m, ok := w.leadMsg.(palette.LeaderSelectedMsg); !ok || m.Action != a {
		return fmt.Errorf("message = %#v", w.leadMsg)
	}
	return nil
}

func (w *helpWorld) popupClosed() error {
	if _, ok := w.leadMsg.(palette.LeaderClosedMsg); !ok {
		return fmt.Errorf("message = %#v", w.leadMsg)
	}
	return nil
}

func TestHelpFeature(t *testing.T) {
	runFeature(t, "help", func(sc *godog.ScenarioContext) {
		w := &helpWorld{}
		sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
			nw, err := newHelpWorld()
			if err == nil {
				*w = *nw
			}
			return ctx, err
		})
		sc.Step(`^the bindings (.+)$`, w.bindingsList)
		sc.Step(`^(\d+) bindings "f1 one" to "f10 ten"$`, w.manyBindings)
		sc.Step(`^the footer is rendered (\d+) columns wide$`, w.renderFooter)
		sc.Step(`^the footer is "([^"]*)"$`, w.footerIs)
		sc.Step(`^the footer shows (\d+) shortcuts$`, w.footerShows)
		sc.Step(`^the footer does not contain "([^"]*)"$`, w.footerLacks)
		sc.Step(`^I open the help for each scope of the default keymap$`, w.walkScopes)
		sc.Step(`^every binding of that scope is listed with its key and description$`, w.walkOK)
		sc.Step(`^every "global" binding is listed under "Global"$`, w.walkOK)
		sc.Step(`^the help modal for scope "([^"]*)"$`, w.helpFor)
		sc.Step(`^I type "([^"]*)"$`, w.typeText)
		sc.Step(`^I press "([^"]*)"$`, func(name string) {
			if w.inLeader {
				w.pressInLeader(name)
				return
			}
			w.pressInModal(name)
		})
		sc.Step(`^the help lists "([^"]*)" with "([^"]*)"$`, w.helpLists)
		sc.Step(`^the help does not list "([^"]*)"$`, w.helpLacks)
		sc.Step(`^the help shows "([^"]*)"$`, w.helpShows)
		sc.Step(`^the help modal reports it closed$`, w.modalClosed)
		sc.Step(`^the leader popup with the default keymap$`, w.leaderDefault)
		sc.Step(`^the popup title is "([^"]*)"$`, w.popupTitle)
		sc.Step(`^the popup lists "([^"]*)" with "([^"]*)"$`, w.popupLists)
		sc.Step(`^the popup reports the action "([^"]*)"$`, w.popupAction)
		sc.Step(`^the popup reports it closed without an action$`, w.popupClosed)
	})
}
