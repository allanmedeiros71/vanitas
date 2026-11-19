// Package theme defines the named colour tokens, the built-in themes, user token overrides and
// the lipgloss styles every component uses.
package theme

import (
	"fmt"
	"regexp"
	"sort"
)

var tokens = []string{
	"bg", "fg", "muted", "border", "focus", "accent", "selection", "zebra", "null",
	"changed", "inserted", "deleted", "error", "warning", "success", "production",
	"sql_keyword", "sql_string", "sql_number", "sql_comment", "sql_function", "sql_operator",
}

var hexColour = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// Tokens returns the token names every theme defines, in canonical order.
func Tokens() []string { return append([]string(nil), tokens...) }

// Theme is a named set of colours, one per token.
type Theme struct {
	Name   string
	colors map[string]string
}

// Color returns the #RRGGBB value of token, or "" for an unknown token.
func (t Theme) Color(token string) string { return t.colors[token] }

// Load returns the named theme with the user's token overrides applied. Unknown tokens and
// invalid colours are ignored and reported as warnings sorted by token.
func Load(name string, overrides map[string]string) (Theme, []string, error) {
	base, ok := builtin[name]
	if !ok {
		return Theme{}, nil, fmt.Errorf("unknown theme %q", name)
	}
	colors := make(map[string]string, len(base))
	for k, v := range base {
		colors[k] = v
	}
	keys := make([]string, 0, len(overrides))
	for k := range overrides {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var warnings []string
	for _, k := range keys {
		v := overrides[k]
		if _, known := base[k]; !known {
			warnings = append(warnings, fmt.Sprintf("theme_overrides: unknown token %q", k))
			continue
		}
		if !hexColour.MatchString(v) {
			warnings = append(warnings, fmt.Sprintf("theme_overrides: invalid colour %q for %q", v, k))
			continue
		}
		colors[k] = v
	}
	return Theme{Name: name, colors: colors}, warnings, nil
}
