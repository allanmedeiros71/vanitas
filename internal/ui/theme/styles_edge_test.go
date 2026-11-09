package theme

import (
	"image/color"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

var colourEscape = regexp.MustCompile(`\x1b\[[0-9;]*(3[0-9]|4[0-9]|9[0-7]|10[0-7])[;m]`)

func TestOverridesDoNotLeakIntoLaterLoads(t *testing.T) {
	t.Parallel()
	before, _, _ := Load("nord", nil)
	if _, _, err := Load("nord", map[string]string{"zebra": "#010203", "fg": "#040506"}); err != nil {
		t.Fatal(err)
	}
	after, warnings, err := Load("nord", nil)
	if err != nil || len(warnings) != 0 {
		t.Fatalf("Load: %v %v", err, warnings)
	}
	if !reflect.DeepEqual(before.colors, after.colors) {
		t.Error("override changed the built-in palette")
	}
}

func TestEmptyOverridesAreSilent(t *testing.T) {
	t.Parallel()
	base, _, _ := Load("gruvbox", nil)
	for _, ov := range []map[string]string{nil, {}} {
		th, warnings, err := Load("gruvbox", ov)
		if err != nil || len(warnings) != 0 {
			t.Fatalf("overrides %v: err %v warnings %v", ov, err, warnings)
		}
		if !reflect.DeepEqual(th.colors, base.colors) {
			t.Errorf("overrides %v changed colours", ov)
		}
	}
}

func TestNamesAndTokensReturnCopies(t *testing.T) {
	t.Parallel()
	names, tokenList := Names(), Tokens()
	names[0], tokenList[0] = "mutated", "mutated"
	if Names()[0] != "vanitas" || Tokens()[0] != "bg" {
		t.Error("returned slices share state with the package")
	}
}

func TestNoColourStylesEmitNoColourEscapes(t *testing.T) {
	t.Parallel()
	th, _, _ := Load("dracula", nil)
	s := NewStyles(th, true)
	styles := map[string]lipgloss.Style{
		"App": s.App, "Title": s.Title, "Panel": s.Panel, "PanelFocused": s.PanelFocused,
		"PanelTitle": s.PanelTitle, "Muted": s.Muted, "FooterKey": s.FooterKey,
		"FooterDesc": s.FooterDesc, "Status": s.Status, "StatusWarning": s.StatusWarning,
		"StatusError": s.StatusError, "Modal": s.Modal, "ModalTitle": s.ModalTitle,
		"Selected": s.Selected,
	}
	for name, st := range styles {
		if got := st.Render("x"); colourEscape.MatchString(got) {
			t.Errorf("%s renders a colour under NO_COLOR: %q", name, got)
		}
	}
}

func TestColourStylesUseTokens(t *testing.T) {
	t.Parallel()
	th, _, _ := Load("catppuccin-mocha", nil)
	s := NewStyles(th, false)
	checks := []struct {
		name string
		got  color.Color
		want string
	}{
		{"Muted foreground", s.Muted.GetForeground(), th.Color("muted")},
		{"Status foreground", s.Status.GetForeground(), th.Color("fg")},
		{"StatusWarning foreground", s.StatusWarning.GetForeground(), th.Color("warning")},
		{"StatusError foreground", s.StatusError.GetForeground(), th.Color("error")},
		{"Selected background", s.Selected.GetBackground(), th.Color("selection")},
		{"Panel border", s.Panel.GetBorderTopForeground(), th.Color("border")},
		{"PanelFocused border", s.PanelFocused.GetBorderTopForeground(), th.Color("focus")},
		{"Modal border", s.Modal.GetBorderTopForeground(), th.Color("focus")},
	}
	for _, c := range checks {
		wr, wg, wb, _ := lipgloss.Color(c.want).RGBA()
		gr, gg, gb, _ := c.got.RGBA()
		if wr != gr || wg != gg || wb != gb {
			t.Errorf("%s = %v, want %s", c.name, c.got, c.want)
		}
	}
	if s.Selected.GetReverse() {
		t.Error("colour selection should not use reverse video")
	}
}

func TestRenderPanelKeepsWidthWithWideRunes(t *testing.T) {
	t.Parallel()
	th, _, _ := Load("vanitas", nil)
	for _, noColor := range []bool{false, true} {
		s := NewStyles(th, noColor)
		out := s.RenderPanel("表格结果", "数据库连接字符串很长很长", 12, 4, true)
		lines := strings.Split(out, "\n")
		if len(lines) != 4 {
			t.Fatalf("nc=%v: %d lines", noColor, len(lines))
		}
		for _, l := range lines {
			if w := lipgloss.Width(l); w != 12 {
				t.Errorf("nc=%v: width %d: %q", noColor, w, l)
			}
		}
	}
}

func TestRenderPanelMinimalSize(t *testing.T) {
	t.Parallel()
	th, _, _ := Load("vanitas", nil)
	out := NewStyles(th, false).RenderPanel("T", "body", 2, 2, false)
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("%d lines: %q", len(lines), out)
	}
	for _, l := range lines {
		if lipgloss.Width(l) != 2 {
			t.Errorf("width %d: %q", lipgloss.Width(l), l)
		}
	}
}
