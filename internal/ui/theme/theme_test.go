package theme

import (
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

func lum(t *testing.T, hex string) float64 {
	t.Helper()
	var ch [3]float64
	for i := range ch {
		v, err := strconv.ParseUint(hex[1+2*i:3+2*i], 16, 8)
		if err != nil {
			t.Fatalf("bad hex %q", hex)
		}
		f := float64(v) / 255
		if f <= 0.03928 {
			ch[i] = f / 12.92
		} else {
			ch[i] = math.Pow((f+0.055)/1.055, 2.4)
		}
	}
	return 0.2126*ch[0] + 0.7152*ch[1] + 0.0722*ch[2]
}

func contrast(t *testing.T, a, b string) float64 {
	la, lb := lum(t, a), lum(t, b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func TestNamesOrder(t *testing.T) {
	t.Parallel()
	want := []string{"vanitas", "light", "tokyo-night", "rose-pine", "gruvbox", "nord", "catppuccin-mocha", "dracula"}
	if !reflect.DeepEqual(Names(), want) {
		t.Fatalf("Names() = %v", Names())
	}
	if len(Tokens()) != 22 || Tokens()[0] != "bg" || Tokens()[21] != "sql_operator" {
		t.Fatalf("Tokens() = %v", Tokens())
	}
}

func TestEveryThemeCompleteAndContrast(t *testing.T) {
	t.Parallel()
	for _, n := range Names() {
		th, w, err := Load(n, nil)
		if err != nil || len(w) != 0 {
			t.Fatalf("%s: %v %v", n, err, w)
		}
		if len(th.colors) != len(Tokens()) {
			t.Errorf("%s: %d colours, want %d", n, len(th.colors), len(Tokens()))
		}
		for _, tok := range Tokens() {
			if !hexColour.MatchString(th.Color(tok)) {
				t.Errorf("%s.%s = %q", n, tok, th.Color(tok))
			}
		}
		if r := contrast(t, th.Color("fg"), th.Color("bg")); r < 4.5 {
			t.Errorf("%s fg/bg contrast %.2f", n, r)
		}
		if r := contrast(t, th.Color("focus"), th.Color("bg")); r < 3.0 {
			t.Errorf("%s focus/bg contrast %.2f", n, r)
		}
	}
}

func TestVanitasPalette(t *testing.T) {
	t.Parallel()
	th, _, _ := Load("vanitas", nil)
	want := map[string]string{
		"bg": "#14161B", "fg": "#E6E1D6", "muted": "#7D8590", "border": "#2D323B",
		"focus": "#C9A227", "accent": "#D9B95B", "selection": "#2A2F38", "zebra": "#191C22", "null": "#6B7280",
		"changed": "#E0A458", "inserted": "#7FB77E", "deleted": "#D16969", "error": "#E06C75",
		"warning": "#E5C07B", "success": "#98C379", "production": "#D14444", "sql_keyword": "#C9A227",
		"sql_string": "#98C379", "sql_number": "#D19A66", "sql_comment": "#6B7280",
		"sql_function": "#61AFEF", "sql_operator": "#ABB2BF",
	}
	for k, v := range want {
		if th.Color(k) != v {
			t.Errorf("%s = %s, want %s", k, th.Color(k), v)
		}
	}
}

func TestLoad(t *testing.T) {
	t.Parallel()
	if _, _, err := Load("solarized", nil); err == nil || err.Error() != `unknown theme "solarized"` {
		t.Fatalf("err = %v", err)
	}
	th, w, err := Load("nord", map[string]string{
		"zebra": "#aabbcc", "fg": "blue", "sparkle": "#112233", "bg": "#12345", "accent": "#GGGGGG",
	})
	if err != nil {
		t.Fatal(err)
	}
	if th.Color("zebra") != "#aabbcc" {
		t.Errorf("zebra = %s", th.Color("zebra"))
	}
	base, _, _ := Load("nord", nil)
	if th.Color("fg") != base.Color("fg") {
		t.Error("invalid override applied")
	}
	want := []string{
		`theme_overrides: invalid colour "#GGGGGG" for "accent"`,
		`theme_overrides: invalid colour "#12345" for "bg"`,
		`theme_overrides: invalid colour "blue" for "fg"`,
		`theme_overrides: unknown token "sparkle"`,
	}
	if !reflect.DeepEqual(w, want) {
		t.Errorf("warnings = %q", w)
	}
	if th.Color("nope") != "" {
		t.Error("unknown token has colour")
	}
}

func TestNoColorFromEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "")
	if NoColorFromEnv() {
		t.Error("empty NO_COLOR counts")
	}
	t.Setenv("NO_COLOR", "1")
	if !NoColorFromEnv() {
		t.Error("NO_COLOR=1 ignored")
	}
}

func TestRenderPanel(t *testing.T) {
	t.Parallel()
	th, _, _ := Load("vanitas", nil)
	for _, noColor := range []bool{false, true} {
		s := NewStyles(th, noColor)
		for _, focused := range []bool{false, true} {
			for _, sz := range [][2]int{{30, 6}, {12, 3}, {8, 2}} {
				out := s.RenderPanel("Explorer", "line one\nsecond line that is far too long to fit\n\n\nextra", sz[0], sz[1], focused)
				lines := strings.Split(out, "\n")
				if len(lines) != sz[1] {
					t.Fatalf("%v: %d lines", sz, len(lines))
				}
				for _, l := range lines {
					if lipgloss.Width(l) != sz[0] {
						t.Errorf("%v nc=%v f=%v: width %d: %q", sz, noColor, focused, lipgloss.Width(l), l)
					}
				}
				if sz[0] >= 12 && !strings.Contains(lines[0], "Expl") {
					t.Errorf("title missing: %q", lines[0])
				}
				if thick := strings.Contains(out, "┏"); thick != (noColor && focused) {
					t.Errorf("nc=%v f=%v thick=%v", noColor, focused, thick)
				}
			}
		}
	}
	s := NewStyles(th, false)
	if s.RenderPanel("x", "y", 1, 5, false) != "" || s.RenderPanel("x", "y", 5, 1, false) != "" {
		t.Error("degenerate size should be empty")
	}
	if !strings.Contains(NewStyles(th, true).RenderPanel("", "", 6, 3, false), "╭────╮") {
		t.Error("untitled top border")
	}
}

func TestNoColorStyles(t *testing.T) {
	t.Parallel()
	th, _, _ := Load("dracula", nil)
	nc, col := NewStyles(th, true), NewStyles(th, false)
	if !nc.Selected.GetReverse() || col.Selected.GetReverse() {
		t.Error("selection style")
	}
	if got := nc.Title.Render("x"); strings.Contains(got, "38;") || strings.Contains(got, "48;") {
		t.Errorf("colour in no-colour title: %q", got)
	}
	if !strings.Contains(col.Title.Render("x"), "\x1b[") {
		t.Error("colour title has no escape")
	}
	if nc.Modal.GetBorderStyle() != lipgloss.ThickBorder() || col.Modal.GetBorderStyle() != lipgloss.RoundedBorder() {
		t.Error("modal border")
	}
}
