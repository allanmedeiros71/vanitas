// Package keymap is the single registry of every keyboard shortcut: scopes, actions, default
// keys, footer flags and the user's keymap.toml overrides.
//
// Key strings are Bubble Tea v2 tea.KeyPressMsg.String() values ("ctrl+q", "alt+1", "space",
// "?"); matching is an exact string comparison.
package keymap

import (
	"slices"
	"strings"

	"charm.land/bubbles/v2/key"
)

// Scope is a keyboard context (a panel, an overlay or the global layer).
type Scope string

// Action is one command of a scope with its current keys. Keys is empty when unbound.
type Action struct {
	ID     string
	Scope  Scope
	Desc   string
	Keys   []string
	Footer bool
}

// Binding is a bound action ready for the footer and the help modal.
type Binding struct {
	Action string
	Footer bool
	key.Binding
}

// entry is an action together with its default keys, used when an override is rejected.
type entry struct {
	Action
	defaults []string
}

// Registry holds the actions of every scope. Build one with Defaults.
type Registry struct {
	scopes  []Scope
	actions map[Scope][]*entry
}

func newRegistry(scopes []Scope, actions []Action) *Registry {
	r := &Registry{
		scopes:  slices.Clone(scopes),
		actions: make(map[Scope][]*entry, len(scopes)),
	}
	for _, a := range actions {
		a.Keys = slices.Clone(a.Keys)
		r.actions[a.Scope] = append(r.actions[a.Scope], &entry{Action: a, defaults: slices.Clone(a.Keys)})
	}
	return r
}

// known reports whether s is a declared scope, with or without actions.
func (r *Registry) known(s Scope) bool {
	return slices.Contains(r.scopes, s)
}

// find returns the action id of scope s, or nil.
func (r *Registry) find(s Scope, id string) *entry {
	for _, e := range r.actions[s] {
		if e.ID == id {
			return e
		}
	}
	return nil
}

// Scopes returns the known scopes that have at least one action, in declaration order.
func (r *Registry) Scopes() []Scope {
	var out []Scope
	for _, s := range r.scopes {
		if len(r.actions[s]) > 0 {
			out = append(out, s)
		}
	}
	return out
}

// Actions returns a copy of the actions of s in declaration order, unbound ones included.
func (r *Registry) Actions(s Scope) []Action {
	var out []Action
	for _, e := range r.actions[s] {
		a := e.Action
		a.Keys = slices.Clone(e.Keys)
		out = append(out, a)
	}
	return out
}

// Bindings returns the bound actions of s in declaration order.
func (r *Registry) Bindings(s Scope) []Binding {
	var out []Binding
	for _, e := range r.actions[s] {
		if len(e.Keys) == 0 {
			continue
		}
		display := make([]string, len(e.Keys))
		for i, k := range e.Keys {
			display[i] = DisplayKey(k)
		}
		out = append(out, Binding{
			Action: e.ID,
			Footer: e.Footer,
			Binding: key.NewBinding(
				key.WithKeys(e.Keys...),
				key.WithHelp(strings.Join(display, "/"), e.Desc),
			),
		})
	}
	return out
}

// Match returns the action of s bound to keyStr (a tea.KeyPressMsg.String() value).
func (r *Registry) Match(s Scope, keyStr string) (string, bool) {
	for _, e := range r.actions[s] {
		if slices.Contains(e.Keys, keyStr) {
			return e.ID, true
		}
	}
	return "", false
}

// DisplayKey returns the label shown for a key string: "space" is "␣", others are unchanged.
func DisplayKey(k string) string {
	if k == "space" {
		return "␣"
	}
	return k
}
