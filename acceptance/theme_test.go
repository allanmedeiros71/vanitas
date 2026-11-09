package acceptance

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/cucumber/godog"

	"github.com/allanmedeiros71/vanitas/internal/ui/theme"
)

type themeWorld struct {
	noColor  bool
	loaded   theme.Theme
	warnings []string
	loadErr  error
	output   string
}

var colourSGR = regexp.MustCompile(`\x1b\[[0-9;]*(3[0-9]|4[0-9]|9[0-7]|10[0-7])[;m]`)

func (w *themeWorld) listThemes() {}

func (w *themeWorld) themesAre(list string) error {
	if got := strings.Join(theme.Names(), ", "); got != list {
		return fmt.Errorf("themes = %q", got)
	}
	return nil
}

func (w *themeWorld) load(name string) { w.loaded, w.warnings, w.loadErr = theme.Load(name, nil) }

func (w *themeWorld) loadOverride(name, token, value string) {
	w.loaded, w.warnings, w.loadErr = theme.Load(name, map[string]string{token: value})
}

func (w *themeWorld) everyToken(format string) error {
	if w.loadErr != nil {
		return w.loadErr
	}
	re := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	if format != "#RRGGBB" {
		return fmt.Errorf("unsupported format %q", format)
	}
	for _, tok := range theme.Tokens() {
		if !re.MatchString(w.loaded.Color(tok)) {
			return fmt.Errorf("token %s = %q", tok, w.loaded.Color(tok))
		}
	}
	return nil
}

func luminance(hex string) float64 {
	var sum float64
	for i, k := range []float64{0.2126, 0.7152, 0.0722} {
		v, _ := strconv.ParseUint(hex[1+2*i:3+2*i], 16, 8)
		f := float64(v) / 255
		if f <= 0.03928 {
			f /= 12.92
		} else {
			f = math.Pow((f+0.055)/1.055, 2.4)
		}
		sum += k * f
	}
	return sum
}

func (w *themeWorld) contrast(a, b string, atLeast float64) error {
	la, lb := luminance(w.loaded.Color(a)), luminance(w.loaded.Color(b))
	if la < lb {
		la, lb = lb, la
	}
	if r := (la + 0.05) / (lb + 0.05); r < atLeast {
		return fmt.Errorf("contrast %s/%s = %.2f", a, b, r)
	}
	return nil
}

func (w *themeWorld) failsWith(msg string) error {
	if w.loadErr == nil {
		return fmt.Errorf("load succeeded, want error %q", msg)
	}
	if w.loadErr.Error() != unquote(msg) {
		return fmt.Errorf("unexpected load error: %w", w.loadErr)
	}
	return nil
}

func unquote(s string) string { return strings.ReplaceAll(s, `\"`, `"`) }

func (w *themeWorld) tokenIs(token, want string) error {
	if got := w.loaded.Color(token); got != want {
		return fmt.Errorf("%s = %s", token, got)
	}
	return nil
}

func (w *themeWorld) tokenSame(token, name string) error {
	ref, _, err := theme.Load(name, nil)
	if err != nil {
		return err
	}
	return w.tokenIs(token, ref.Color(token))
}

func (w *themeWorld) hasWarning(msg string) error {
	if !slices.Contains(w.warnings, unquote(msg)) {
		return fmt.Errorf("warnings = %q", w.warnings)
	}
	return nil
}

func (w *themeWorld) renderFocused(title, name string) error {
	th, _, err := theme.Load(name, nil)
	if err != nil {
		return err
	}
	w.output = theme.NewStyles(th, w.noColor).RenderPanel(title, "body", 24, 5, true)
	return nil
}

func (w *themeWorld) noEscapes() error {
	if colourSGR.MatchString(w.output) {
		return fmt.Errorf("colour escape in %q", w.output)
	}
	return nil
}

func (w *themeWorld) thick() error {
	if !strings.Contains(w.output, "┏") {
		return fmt.Errorf("border not thick: %q", w.output)
	}
	return nil
}

func TestThemeFeature(t *testing.T) {
	runFeature(t, "theme", func(sc *godog.ScenarioContext) {
		w := &themeWorld{}
		sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
			*w = themeWorld{}
			return ctx, nil
		})
		sc.Step(`^NO_COLOR is set$`, func() { t.Setenv("NO_COLOR", "1"); w.noColor = theme.NoColorFromEnv() })
		sc.Step(`^I list the themes$`, w.listThemes)
		sc.Step(`^the themes are "([^"]*)"$`, w.themesAre)
		sc.Step(`^I load the theme "([^"]*)"$`, w.load)
		sc.Step(`^I load the theme "([^"]*)" with the override "([^"]*)" = "([^"]*)"$`, w.loadOverride)
		sc.Step(`^every token has a "([^"]*)" colour$`, w.everyToken)
		sc.Step(`^the contrast of "([^"]*)" on "([^"]*)" is at least ([0-9.]+)$`, func(a, b, m string) error {
			f, _ := strconv.ParseFloat(m, 64)
			return w.contrast(a, b, f)
		})
		sc.Step(`^loading the theme fails with the error "((?:[^"\\]|\\.)*)"$`, w.failsWith)
		sc.Step(`^the token "([^"]*)" is "([^"]*)"$`, w.tokenIs)
		sc.Step(`^the token "([^"]*)" is the same as in the theme "([^"]*)"$`, w.tokenSame)
		sc.Step(`^there is the theme warning "((?:[^"\\]|\\.)*)"$`, w.hasWarning)
		sc.Step(`^I render a focused panel titled "([^"]*)" with the theme "([^"]*)"$`, w.renderFocused)
		sc.Step(`^the output has no colour escape sequences$`, w.noEscapes)
		sc.Step(`^the panel border is thick$`, w.thick)
	})
}
