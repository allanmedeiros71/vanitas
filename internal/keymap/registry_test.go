package keymap

import (
	"reflect"
	"testing"
)

func TestDefaultsNoSharedKeysWithinScope(t *testing.T) {
	t.Parallel()
	r := Defaults()
	for _, s := range knownScopes {
		seen := map[string]string{}
		for _, a := range r.Actions(s) {
			if a.Desc == "" {
				t.Errorf("[%s] %s has no description", s, a.ID)
			}
			if len(a.Keys) == 0 {
				t.Errorf("[%s] %s has no default key", s, a.ID)
			}
			for _, k := range a.Keys {
				if other, ok := seen[k]; ok {
					t.Errorf("[%s] key %q shared by %s and %s", s, k, other, a.ID)
				}
				seen[k] = a.ID
			}
		}
	}
}

func TestDefaultsKnowsEveryScope(t *testing.T) {
	t.Parallel()
	r := Defaults()
	for _, s := range []Scope{
		"global", "explorer", "editor.normal", "editor.insert",
		"grid", "connpicker", "help", "leader",
	} {
		if !r.known(s) {
			t.Errorf("scope %q is not known", s)
		}
	}
	if r.known("sidebar") {
		t.Error(`scope "sidebar" must not be known`)
	}
}

func TestDefaultsTable(t *testing.T) {
	t.Parallel()
	r := Defaults()
	want := map[Scope][]Action{
		ScopeGlobal: {
			{ID: "focus_explorer", Scope: ScopeGlobal, Desc: "explorer", Keys: []string{"alt+1"}, Footer: true},
			{ID: "focus_editor", Scope: ScopeGlobal, Desc: "editor", Keys: []string{"alt+2"}, Footer: true},
			{ID: "focus_grid", Scope: ScopeGlobal, Desc: "results", Keys: []string{"alt+3"}, Footer: true},
			{ID: "next_panel", Scope: ScopeGlobal, Desc: "next panel", Keys: []string{"tab"}},
			{ID: "prev_panel", Scope: ScopeGlobal, Desc: "prev panel", Keys: []string{"shift+tab"}},
			{ID: "help", Scope: ScopeGlobal, Desc: "help", Keys: []string{"?"}, Footer: true},
			{ID: "leader", Scope: ScopeGlobal, Desc: "menu", Keys: []string{"space"}, Footer: true},
			{ID: "quit", Scope: ScopeGlobal, Desc: "quit", Keys: []string{"ctrl+q"}, Footer: true},
		},
		ScopeHelp: {
			{ID: "close", Scope: ScopeHelp, Desc: "close", Keys: []string{"esc"}, Footer: true},
			{ID: "scroll_up", Scope: ScopeHelp, Desc: "scroll up", Keys: []string{"up"}, Footer: true},
			{ID: "scroll_down", Scope: ScopeHelp, Desc: "scroll down", Keys: []string{"down"}, Footer: true},
		},
		ScopeLeader: {
			{ID: "toggle_explorer", Scope: ScopeLeader, Desc: "toggle explorer", Keys: []string{"e"}, Footer: true},
			{ID: "switch_theme", Scope: ScopeLeader, Desc: "switch theme", Keys: []string{"t"}, Footer: true},
			{ID: "quit", Scope: ScopeLeader, Desc: "quit", Keys: []string{"q"}, Footer: true},
			{ID: "close", Scope: ScopeLeader, Desc: "close", Keys: []string{"esc"}, Footer: true},
		},
	}
	for s, w := range want {
		if got := r.Actions(s); !reflect.DeepEqual(got, w) {
			t.Errorf("Actions(%s) = %+v, want %+v", s, got, w)
		}
	}
	if got := r.Actions(ScopeGrid); got != nil {
		t.Errorf("Actions(grid) = %+v, want none", got)
	}
}

func TestScopesOnlyWithActionsInOrder(t *testing.T) {
	t.Parallel()
	got := Defaults().Scopes()
	want := []Scope{ScopeGlobal, ScopeHelp, ScopeLeader}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Scopes() = %v, want %v", got, want)
	}
}

func TestDefaultsIndependent(t *testing.T) {
	t.Parallel()
	a, b := Defaults(), Defaults()
	a.Apply(Overrides{ScopeGlobal: {"quit": {"f10"}}})
	if id, ok := b.Match(ScopeGlobal, "ctrl+q"); !ok || id != "quit" {
		t.Errorf("second registry changed: ctrl+q → %q %v", id, ok)
	}
	if _, ok := b.Match(ScopeGlobal, "f10"); ok {
		t.Error("second registry sees f10")
	}
	if _, ok := Defaults().Match(ScopeGlobal, "f10"); ok {
		t.Error("a new registry sees f10")
	}
}

func TestActionsReturnsCopies(t *testing.T) {
	t.Parallel()
	r := Defaults()
	acts := r.Actions(ScopeGlobal)
	acts[0].Keys[0] = "x"
	acts[0].ID = "y"
	if id, ok := r.Match(ScopeGlobal, "alt+1"); !ok || id != "focus_explorer" {
		t.Errorf("mutating Actions result changed the registry: %q %v", id, ok)
	}
}

func TestMatch(t *testing.T) {
	t.Parallel()
	r := Defaults()
	tests := []struct {
		scope  Scope
		key    string
		want   string
		wantOK bool
	}{
		{ScopeGlobal, "ctrl+q", "quit", true},
		{ScopeGlobal, "?", "help", true},
		{ScopeGlobal, "space", "leader", true},
		{ScopeGlobal, "alt+3", "focus_grid", true},
		{ScopeLeader, "q", "quit", true},
		{ScopeLeader, "esc", "close", true},
		{ScopeHelp, "down", "scroll_down", true},
		{ScopeGlobal, "q", "", false},
		{ScopeGlobal, "CTRL+Q", "", false},
		{ScopeGrid, "ctrl+q", "", false},
		{"nope", "ctrl+q", "", false},
	}
	for _, tt := range tests {
		got, ok := r.Match(tt.scope, tt.key)
		if got != tt.want || ok != tt.wantOK {
			t.Errorf("Match(%s, %q) = %q %v, want %q %v", tt.scope, tt.key, got, ok, tt.want, tt.wantOK)
		}
	}
}

func TestBindings(t *testing.T) {
	t.Parallel()
	r := Defaults()
	r.Apply(Overrides{
		ScopeGlobal: {"focus_editor": {"none"}, "quit": {"ctrl+q", "space"}, "leader": {"f2"}},
	})
	got := r.Bindings(ScopeGlobal)
	type row struct {
		action string
		footer bool
		keys   []string
		help   string
		desc   string
	}
	want := []row{
		{"focus_explorer", true, []string{"alt+1"}, "alt+1", "explorer"},
		{"focus_grid", true, []string{"alt+3"}, "alt+3", "results"},
		{"next_panel", false, []string{"tab"}, "tab", "next panel"},
		{"prev_panel", false, []string{"shift+tab"}, "shift+tab", "prev panel"},
		{"help", true, []string{"?"}, "?", "help"},
		{"leader", true, []string{"f2"}, "f2", "menu"},
		{"quit", true, []string{"ctrl+q", "space"}, "ctrl+q/␣", "quit"},
	}
	if len(got) != len(want) {
		t.Fatalf("Bindings = %d rows, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		g := got[i]
		h := g.Help()
		if g.Action != w.action || g.Footer != w.footer || !reflect.DeepEqual(g.Keys(), w.keys) ||
			h.Key != w.help || h.Desc != w.desc || !g.Enabled() {
			t.Errorf("row %d = {%s %v %v %q %q}, want %+v", i, g.Action, g.Footer, g.Keys(), h.Key, h.Desc, w)
		}
	}
	if b := r.Bindings("nope"); b != nil {
		t.Errorf("Bindings(unknown) = %+v", b)
	}
}

func TestDisplayKey(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{"space": "␣", "ctrl+q": "ctrl+q", "?": "?", "": ""} {
		if got := DisplayKey(in); got != want {
			t.Errorf("DisplayKey(%q) = %q, want %q", in, got, want)
		}
	}
}
