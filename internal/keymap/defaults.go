package keymap

// Known scopes, in declaration order. Every scope is known even when it has no actions yet, so
// an override for it reports an unknown action rather than an unknown scope.
const (
	ScopeGlobal       Scope = "global"
	ScopeExplorer     Scope = "explorer"
	ScopeEditorNormal Scope = "editor.normal"
	ScopeEditorInsert Scope = "editor.insert"
	ScopeGrid         Scope = "grid"
	ScopeConnPicker   Scope = "connpicker"
	ScopeHelp         Scope = "help"
	ScopeLeader       Scope = "leader"
)

// knownScopes lists every scope in declaration order.
var knownScopes = []Scope{
	ScopeGlobal, ScopeExplorer, ScopeEditorNormal, ScopeEditorInsert,
	ScopeGrid, ScopeConnPicker, ScopeHelp, ScopeLeader,
}

// defaultActions is the phase 1 keymap; its order is the listing order everywhere.
func defaultActions() []Action {
	return []Action{
		{Scope: ScopeGlobal, ID: "focus_explorer", Keys: []string{"alt+1"}, Desc: "explorer", Footer: true},
		{Scope: ScopeGlobal, ID: "focus_editor", Keys: []string{"alt+2"}, Desc: "editor", Footer: true},
		{Scope: ScopeGlobal, ID: "focus_grid", Keys: []string{"alt+3"}, Desc: "results", Footer: true},
		{Scope: ScopeGlobal, ID: "next_panel", Keys: []string{"tab"}, Desc: "next panel"},
		{Scope: ScopeGlobal, ID: "prev_panel", Keys: []string{"shift+tab"}, Desc: "prev panel"},
		{Scope: ScopeGlobal, ID: "help", Keys: []string{"?"}, Desc: "help", Footer: true},
		{Scope: ScopeGlobal, ID: "leader", Keys: []string{"space"}, Desc: "menu", Footer: true},
		{Scope: ScopeGlobal, ID: "quit", Keys: []string{"ctrl+q"}, Desc: "quit", Footer: true},
		{Scope: ScopeHelp, ID: "close", Keys: []string{"esc"}, Desc: "close", Footer: true},
		{Scope: ScopeHelp, ID: "scroll_up", Keys: []string{"up"}, Desc: "scroll up", Footer: true},
		{Scope: ScopeHelp, ID: "scroll_down", Keys: []string{"down"}, Desc: "scroll down", Footer: true},
		{Scope: ScopeLeader, ID: "toggle_explorer", Keys: []string{"e"}, Desc: "toggle explorer", Footer: true},
		{Scope: ScopeLeader, ID: "switch_theme", Keys: []string{"t"}, Desc: "switch theme", Footer: true},
		{Scope: ScopeLeader, ID: "quit", Keys: []string{"q"}, Desc: "quit", Footer: true},
		{Scope: ScopeLeader, ID: "close", Keys: []string{"esc"}, Desc: "close", Footer: true},
	}
}

// Defaults returns a new registry holding the default keymap. Each call returns an
// independent registry.
func Defaults() *Registry {
	return newRegistry(knownScopes, defaultActions())
}
