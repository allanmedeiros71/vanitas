package keymap

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/BurntSushi/toml"
)

// Overrides maps scope → action id → keys, as read from keymap.toml. A single "none" key means
// the action is unbound.
type Overrides map[Scope]map[string][]string

// none is the value that unbinds an action.
const none = "none"

// ParseOverrides decodes a keymap.toml document. Each table is a scope; each value is a string
// or an array of strings. Values of another type are left out with a warning. Scope and action
// names are checked by Apply. A syntax error names keymap.toml and the line.
func ParseOverrides(data []byte) (Overrides, []string, error) {
	var doc map[string]any
	if err := toml.Unmarshal(data, &doc); err != nil {
		return nil, nil, fmt.Errorf("keymap.toml: %w", err)
	}
	o := Overrides{}
	var warnings []string
	for _, name := range sortedKeys(doc) {
		table, ok := doc[name].(map[string]any)
		if !ok {
			// A key outside any table is not a scope.
			warnings = append(warnings, fmt.Sprintf("keymap: unknown scope %q", name))
			continue
		}
		s := Scope(name)
		o[s] = map[string][]string{}
		for _, action := range sortedKeys(table) {
			keys, ok := toKeys(table[action])
			if !ok {
				warnings = append(warnings, invalidValue(s, action))
				continue
			}
			o[s][action] = keys
		}
	}
	return o, warnings, nil
}

// toKeys converts a decoded TOML value into a key list when it is a string or an array of
// strings. Emptiness is checked by Apply.
func toKeys(v any) ([]string, bool) {
	switch v := v.(type) {
	case string:
		return []string{v}, true
	case []any:
		keys := make([]string, 0, len(v))
		for _, item := range v {
			k, ok := item.(string)
			if !ok {
				return nil, false
			}
			keys = append(keys, k)
		}
		return keys, true
	}
	return nil, false
}

// LoadOverrides reads and parses the keymap file at path. A missing file gives no overrides
// and no error.
func LoadOverrides(path string) (Overrides, []string, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if errors.Is(err, os.ErrNotExist) {
		return Overrides{}, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("keymap.toml: %w", err)
	}
	return ParseOverrides(data)
}

// Apply merges o into the registry and returns the warnings, in a deterministic order:
// unknown scopes (sorted), then each known scope in declaration order.
//
// Per scope: the current keys of every validly overridden action are removed; then the
// overridden actions are processed by id. An action whose new keys are held by another action
// of the same scope keeps those of its default keys that are still free and the override is
// ignored; otherwise it gets its new keys.
func (r *Registry) Apply(o Overrides) []string {
	var warnings []string
	for _, s := range sortedKeys(o) {
		if !r.known(s) {
			warnings = append(warnings, fmt.Sprintf("keymap: unknown scope %q", s))
		}
	}
	for _, s := range r.scopes {
		if actions, ok := o[s]; ok {
			warnings = append(warnings, r.applyScope(s, actions)...)
		}
	}
	return warnings
}

func (r *Registry) applyScope(s Scope, overrides map[string][]string) []string {
	var warnings []string
	var valid []*entry
	newKeys := map[string][]string{}
	for _, id := range sortedKeys(overrides) {
		e := r.find(s, id)
		if e == nil {
			warnings = append(warnings, fmt.Sprintf("keymap: unknown action %q in [%s]", id, s))
			continue
		}
		keys, ok := normalize(overrides[id])
		if !ok {
			warnings = append(warnings, invalidValue(s, id))
			continue
		}
		valid = append(valid, e)
		newKeys[id] = keys
	}
	for _, e := range valid {
		e.Keys = nil
	}
	for _, e := range valid {
		keys := newKeys[e.ID]
		if other, k := r.holder(s, e, keys); other != "" {
			warnings = append(warnings, fmt.Sprintf(
				"keymap: [%s] key %q for %q conflicts with %q; override ignored", s, k, e.ID, other))
			keys = r.free(s, e, e.defaults)
		}
		e.Keys = keys
	}
	return warnings
}

// normalize validates an override's keys and removes duplicates. A single "none" unbinds.
// An empty list, an empty key, or "none" mixed with other keys is invalid.
func normalize(keys []string) ([]string, bool) {
	if len(keys) == 0 {
		return nil, false
	}
	var out []string
	for _, k := range keys {
		if k == "" || (k == none && len(keys) > 1) {
			return nil, false
		}
		if k != none && !slices.Contains(out, k) {
			out = append(out, k)
		}
	}
	return out, true
}

// holder returns the first action of s other than self holding one of keys, and that key.
func (r *Registry) holder(s Scope, self *entry, keys []string) (string, string) {
	for _, k := range keys {
		for _, e := range r.actions[s] {
			if e != self && slices.Contains(e.Keys, k) {
				return e.ID, k
			}
		}
	}
	return "", ""
}

// free returns the keys not held by any action of s other than self.
func (r *Registry) free(s Scope, self *entry, keys []string) []string {
	var out []string
	for _, k := range keys {
		if other, _ := r.holder(s, self, []string{k}); other == "" {
			out = append(out, k)
		}
	}
	return out
}

func invalidValue(s Scope, action string) string {
	return fmt.Sprintf("keymap: invalid value for %q in [%s]", action, s)
}

func sortedKeys[K ~string, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
