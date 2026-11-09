package keymap

import (
	"reflect"
	"testing"
)

// A rejected override gives its defaults back, and a later override in the same scope that
// wants one of them is rejected in turn (processing is by action id, one override at a time).
func TestApplyRejectedDefaultsBlockLaterOverride(t *testing.T) {
	t.Parallel()
	r := Defaults()
	w := r.Apply(Overrides{ScopeGlobal: {
		"help":   {"ctrl+q"},
		"leader": {"?"},
	}})
	wantWarnings := []string{
		`keymap: [global] key "ctrl+q" for "help" conflicts with "quit"; override ignored`,
		`keymap: [global] key "?" for "leader" conflicts with "help"; override ignored`,
	}
	if !reflect.DeepEqual(w, wantWarnings) {
		t.Fatalf("warnings = %q, want %q", w, wantWarnings)
	}
	want := map[string][]string{"help": {"?"}, "leader": {"space"}, "quit": {"ctrl+q"}}
	cur := keysOf(r, ScopeGlobal)
	for id, keys := range want {
		if !reflect.DeepEqual(cur[id], keys) {
			t.Errorf("%s keys = %q, want %q", id, cur[id], keys)
		}
	}
}

// A rejected action keeps only the default keys that are still free; the others are dropped.
func TestApplyRejectedKeepsOnlyFreeDefaults(t *testing.T) {
	t.Parallel()
	r := newRegistry([]Scope{"s"}, []Action{
		{Scope: "s", ID: "a", Desc: "a", Keys: []string{"z"}},
		{Scope: "s", ID: "b", Desc: "b", Keys: []string{"p", "q"}},
		{Scope: "s", ID: "c", Desc: "c", Keys: []string{"q"}},
	})
	w := r.Apply(Overrides{"s": {"b": {"z"}}})
	wantWarnings := []string{`keymap: [s] key "z" for "b" conflicts with "a"; override ignored`}
	if !reflect.DeepEqual(w, wantWarnings) {
		t.Fatalf("warnings = %q, want %q", w, wantWarnings)
	}
	cur := keysOf(r, "s")
	want := map[string][]string{"a": {"z"}, "b": {"p"}, "c": {"q"}}
	for id, keys := range want {
		if !reflect.DeepEqual(cur[id], keys) {
			t.Errorf("%s keys = %q, want %q", id, cur[id], keys)
		}
	}
}
