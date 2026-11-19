package keymap

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// keysOf returns the current keys of every action of s, by id.
func keysOf(r *Registry, s Scope) map[string][]string {
	out := map[string][]string{}
	for _, a := range r.Actions(s) {
		out[a.ID] = a.Keys
	}
	return out
}

func TestApply(t *testing.T) {
	t.Parallel()
	conflict := func(s, k, a, o string) string {
		return "keymap: [" + s + "] key \"" + k + "\" for \"" + a + "\" conflicts with \"" + o + "\"; override ignored"
	}
	tests := []struct {
		name     string
		o        Overrides
		warnings []string
		want     map[Scope]map[string][]string // only the actions checked
	}{
		{
			name: "single override",
			o:    Overrides{ScopeGlobal: {"help": {"f1"}}},
			want: map[Scope]map[string][]string{ScopeGlobal: {"help": {"f1"}, "quit": {"ctrl+q"}}},
		},
		{
			name: "multi-key",
			o:    Overrides{ScopeGlobal: {"quit": {"ctrl+q", "f10"}}},
			want: map[Scope]map[string][]string{ScopeGlobal: {"quit": {"ctrl+q", "f10"}}},
		},
		{
			name: "none unbinds",
			o:    Overrides{ScopeLeader: {"switch_theme": {"none"}}},
			want: map[Scope]map[string][]string{ScopeLeader: {"switch_theme": nil, "quit": {"q"}}},
		},
		{
			name:     "conflict with a non-overridden action",
			o:        Overrides{ScopeGlobal: {"help": {"ctrl+q"}}},
			warnings: []string{conflict("global", "ctrl+q", "help", "quit")},
			want:     map[Scope]map[string][]string{ScopeGlobal: {"help": {"?"}, "quit": {"ctrl+q"}}},
		},
		{
			name:     "conflict on the second key names it",
			o:        Overrides{ScopeGlobal: {"help": {"f1", "tab"}}},
			warnings: []string{conflict("global", "tab", "help", "next_panel")},
			want:     map[Scope]map[string][]string{ScopeGlobal: {"help": {"?"}, "next_panel": {"tab"}}},
		},
		{
			name:     "conflict between overridden actions: lower id wins",
			o:        Overrides{ScopeGlobal: {"quit": {"f5"}, "help": {"f5"}}},
			warnings: []string{conflict("global", "f5", "quit", "help")},
			want:     map[Scope]map[string][]string{ScopeGlobal: {"help": {"f5"}, "quit": {"ctrl+q"}}},
		},
		{
			name: "rejected action whose defaults are all taken ends unbound",
			o:    Overrides{ScopeGlobal: {"help": {"ctrl+q"}, "focus_grid": {"?"}}},
			// focus_grid takes "?" first; help conflicts with quit and its default "?" is taken.
			warnings: []string{conflict("global", "ctrl+q", "help", "quit")},
			want: map[Scope]map[string][]string{ScopeGlobal: {
				"focus_grid": {"?"}, "help": nil, "quit": {"ctrl+q"},
			}},
		},
		{
			name: "swap",
			o:    Overrides{ScopeGlobal: {"next_panel": {"shift+tab"}, "prev_panel": {"tab"}}},
			want: map[Scope]map[string][]string{ScopeGlobal: {"next_panel": {"shift+tab"}, "prev_panel": {"tab"}}},
		},
		{
			name: "duplicate keys are removed",
			o:    Overrides{ScopeGlobal: {"quit": {"f10", "ctrl+q", "f10"}}},
			want: map[Scope]map[string][]string{ScopeGlobal: {"quit": {"f10", "ctrl+q"}}},
		},
		{
			name: "own default key is not a conflict",
			o:    Overrides{ScopeGlobal: {"quit": {"ctrl+q"}}},
			want: map[Scope]map[string][]string{ScopeGlobal: {"quit": {"ctrl+q"}}},
		},
		{
			name: "same key in different scopes",
			o:    Overrides{ScopeLeader: {"quit": {"x"}}, ScopeHelp: {"close": {"x"}}},
			want: map[Scope]map[string][]string{ScopeLeader: {"quit": {"x"}}, ScopeHelp: {"close": {"x"}}},
		},
		{
			name:     "unknown scopes are sorted",
			o:        Overrides{"zeta": {"a": {"b"}}, "sidebar": {"open": {"o"}}},
			warnings: []string{`keymap: unknown scope "sidebar"`, `keymap: unknown scope "zeta"`},
		},
		{
			name: "unknown actions, including the wrong scope and a known scope without actions",
			o: Overrides{
				ScopeLeader:   {"help": {"h"}, "launch": {"l"}},
				ScopeExplorer: {"open": {"o"}},
			},
			warnings: []string{
				`keymap: unknown action "open" in [explorer]`,
				`keymap: unknown action "help" in [leader]`,
				`keymap: unknown action "launch" in [leader]`,
			},
			want: map[Scope]map[string][]string{ScopeLeader: {"quit": {"q"}}},
		},
		{
			name: "invalid values keep the defaults",
			o: Overrides{ScopeGlobal: {
				"quit": {}, "help": {""}, "leader": {"none", "f2"}, "focus_grid": {"f3", ""},
			}},
			warnings: []string{
				`keymap: invalid value for "focus_grid" in [global]`,
				`keymap: invalid value for "help" in [global]`,
				`keymap: invalid value for "leader" in [global]`,
				`keymap: invalid value for "quit" in [global]`,
			},
			want: map[Scope]map[string][]string{ScopeGlobal: {
				"quit": {"ctrl+q"}, "help": {"?"}, "leader": {"space"}, "focus_grid": {"alt+3"},
			}},
		},
		{
			name: "invalid value does not free its key for others",
			o:    Overrides{ScopeGlobal: {"quit": {}, "help": {"ctrl+q"}}},
			warnings: []string{
				`keymap: invalid value for "quit" in [global]`,
				conflict("global", "ctrl+q", "help", "quit"),
			},
			want: map[Scope]map[string][]string{ScopeGlobal: {"quit": {"ctrl+q"}, "help": {"?"}}},
		},
		{
			name: "valid override frees its old key for another action",
			o:    Overrides{ScopeGlobal: {"quit": {"f10"}, "help": {"ctrl+q"}}},
			want: map[Scope]map[string][]string{ScopeGlobal: {"quit": {"f10"}, "help": {"ctrl+q"}}},
		},
		{
			name: "unknown scope warnings come before scope warnings",
			o:    Overrides{"sidebar": {"x": {"x"}}, ScopeGlobal: {"nope": {"x"}}},
			warnings: []string{
				`keymap: unknown scope "sidebar"`,
				`keymap: unknown action "nope" in [global]`,
			},
		},
		{
			name: "known scopes in declaration order",
			o:    Overrides{ScopeLeader: {"nope": {"x"}}, ScopeHelp: {"nope": {"x"}}, ScopeGlobal: {"nope": {"x"}}},
			warnings: []string{
				`keymap: unknown action "nope" in [global]`,
				`keymap: unknown action "nope" in [help]`,
				`keymap: unknown action "nope" in [leader]`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := Defaults()
			got := r.Apply(tt.o)
			if !reflect.DeepEqual(got, tt.warnings) {
				t.Errorf("warnings = %q, want %q", got, tt.warnings)
			}
			for s, actions := range tt.want {
				cur := keysOf(r, s)
				for id, keys := range actions {
					if !reflect.DeepEqual(cur[id], keys) {
						t.Errorf("[%s] %s keys = %q, want %q", s, id, cur[id], keys)
					}
				}
			}
		})
	}
}

func TestApplyTwiceUsesDefaultsOnReject(t *testing.T) {
	t.Parallel()
	r := Defaults()
	r.Apply(Overrides{ScopeGlobal: {"help": {"f1"}}})
	w := r.Apply(Overrides{ScopeGlobal: {"help": {"ctrl+q"}}})
	if len(w) != 1 {
		t.Fatalf("warnings = %q", w)
	}
	if got := keysOf(r, ScopeGlobal)["help"]; !reflect.DeepEqual(got, []string{"?"}) {
		t.Errorf("help keys = %q, want [?]", got)
	}
}

func TestParseOverrides(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		doc      string
		want     Overrides
		warnings []string
	}{
		{name: "empty", doc: "", want: Overrides{}},
		{
			name: "string and array",
			doc:  "[global]\nhelp = \"f1\"\nquit = [\"ctrl+q\", \"f10\"]\n[leader]\nquit = \"none\"\n",
			want: Overrides{
				ScopeGlobal: {"help": {"f1"}, "quit": {"ctrl+q", "f10"}},
				ScopeLeader: {"quit": {"none"}},
			},
		},
		{
			name: "unknown names pass through to Apply",
			doc:  "[sidebar]\nopen = \"o\"\n",
			want: Overrides{"sidebar": {"open": {"o"}}},
		},
		{
			name: "empty values pass through to Apply",
			doc:  "[global]\nquit = []\nhelp = \"\"\n",
			want: Overrides{ScopeGlobal: {"quit": {}, "help": {""}}},
		},
		{
			name: "wrong types",
			doc:  "[global]\nquit = 3\nhelp = [\"f1\", 2]\nleader = true\n[global.sub]\nx = \"y\"\n",
			want: Overrides{ScopeGlobal: {}},
			warnings: []string{
				`keymap: invalid value for "help" in [global]`,
				`keymap: invalid value for "leader" in [global]`,
				`keymap: invalid value for "quit" in [global]`,
				`keymap: invalid value for "sub" in [global]`,
			},
		},
		{
			name:     "key outside a table",
			doc:      "quit = \"x\"\n[global]\nhelp = \"f1\"\n",
			want:     Overrides{ScopeGlobal: {"help": {"f1"}}},
			warnings: []string{`keymap: unknown scope "quit"`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, w, err := ParseOverrides([]byte(tt.doc))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("overrides = %#v, want %#v", got, tt.want)
			}
			if !reflect.DeepEqual(w, tt.warnings) {
				t.Errorf("warnings = %q, want %q", w, tt.warnings)
			}
		})
	}
}

func TestParseOverridesSyntaxError(t *testing.T) {
	t.Parallel()
	o, w, err := ParseOverrides([]byte("[global]\nhelp = \"f1\"\nquit = = 1\n"))
	if err == nil {
		t.Fatal("no error")
	}
	if !strings.Contains(err.Error(), "keymap.toml") || !strings.Contains(err.Error(), "line 3") {
		t.Errorf("error %q lacks file name or line", err)
	}
	if o != nil || w != nil {
		t.Errorf("got overrides %v and warnings %v with an error", o, w)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	o, w, err := LoadOverrides(filepath.Join(dir, "keymap.toml"))
	if err != nil || w != nil || o == nil || len(o) != 0 {
		t.Errorf("missing file: %v %v %v, want empty overrides", o, w, err)
	}

	path := filepath.Join(dir, "keymap.toml")
	if err := os.WriteFile(path, []byte("[global]\nhelp = \"f1\"\nquit = 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	o, w, err = LoadOverrides(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(o, Overrides{ScopeGlobal: {"help": {"f1"}}}) {
		t.Errorf("overrides = %v", o)
	}
	if !reflect.DeepEqual(w, []string{`keymap: invalid value for "quit" in [global]`}) {
		t.Errorf("warnings = %q", w)
	}
}

func TestLoadOverridesReadError(t *testing.T) {
	t.Parallel()
	o, w, err := LoadOverrides(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "keymap.toml") {
		t.Errorf("error = %v, want one naming keymap.toml", err)
	}
	if o != nil || w != nil {
		t.Errorf("got %v %v with an error", o, w)
	}
}
