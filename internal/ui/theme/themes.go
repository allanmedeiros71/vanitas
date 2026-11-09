package theme

var names = []string{
	"vanitas", "light", "tokyo-night", "rose-pine", "gruvbox", "nord", "catppuccin-mocha", "dracula",
}

// Names returns the built-in theme names in cycle order.
func Names() []string { return append([]string(nil), names...) }

// mk builds a palette from values given in the order of the tokens slice.
func mk(v ...string) map[string]string {
	m := make(map[string]string, len(tokens))
	for i, t := range tokens {
		m[t] = v[i]
	}
	return m
}

var builtin = map[string]map[string]string{
	// Own palette (DESIGN.md).
	"vanitas": mk("#14161B", "#E6E1D6", "#7D8590", "#2D323B", "#C9A227", "#D9B95B", "#2A2F38",
		"#191C22", "#6B7280", "#E0A458", "#7FB77E", "#D16969", "#E06C75", "#E5C07B", "#98C379",
		"#D14444", "#C9A227", "#98C379", "#D19A66", "#6B7280", "#61AFEF", "#ABB2BF"),
	// Neutral light palette, modelled on GitHub Primer light colours.
	"light": mk("#FAFAFA", "#24292F", "#57606A", "#D0D7DE", "#0969DA", "#8250DF", "#DDF4FF",
		"#F6F8FA", "#6E7781", "#9A6700", "#1A7F37", "#CF222E", "#CF222E", "#9A6700", "#1A7F37",
		"#CF222E", "#CF222E", "#0A3069", "#0550AE", "#6E7781", "#8250DF", "#24292F"),
	// Tokyo Night "night" variant (github.com/folke/tokyonight.nvim).
	"tokyo-night": mk("#1A1B26", "#C0CAF5", "#565F89", "#3B4261", "#7AA2F7", "#BB9AF7", "#283457",
		"#1F2335", "#545C7E", "#E0AF68", "#9ECE6A", "#F7768E", "#DB4B4B", "#E0AF68", "#9ECE6A",
		"#F7768E", "#BB9AF7", "#9ECE6A", "#FF9E64", "#565F89", "#7AA2F7", "#89DDFF"),
	// Rosé Pine main variant (rosepinetheme.com/palette).
	"rose-pine": mk("#191724", "#E0DEF4", "#6E6A86", "#403D52", "#C4A7E7", "#EBBCBA", "#26233A",
		"#1F1D2E", "#6E6A86", "#F6C177", "#9CCFD8", "#EB6F92", "#EB6F92", "#F6C177", "#9CCFD8",
		"#EB6F92", "#31748F", "#F6C177", "#EBBCBA", "#6E6A86", "#C4A7E7", "#908CAA"),
	// Gruvbox dark, medium contrast (github.com/morhetz/gruvbox).
	"gruvbox": mk("#282828", "#EBDBB2", "#928374", "#504945", "#FABD2F", "#FE8019", "#3C3836",
		"#32302F", "#928374", "#FE8019", "#B8BB26", "#FB4934", "#FB4934", "#FABD2F", "#B8BB26",
		"#FB4934", "#FB4934", "#B8BB26", "#D3869B", "#928374", "#8EC07C", "#FE8019"),
	// Nord (nordtheme.com/docs/colors-and-palettes).
	"nord": mk("#2E3440", "#D8DEE9", "#7B88A1", "#434C5E", "#88C0D0", "#8FBCBB", "#3B4252",
		"#333A47", "#616E88", "#EBCB8B", "#A3BE8C", "#BF616A", "#BF616A", "#EBCB8B", "#A3BE8C",
		"#BF616A", "#81A1C1", "#A3BE8C", "#B48EAD", "#616E88", "#88C0D0", "#81A1C1"),
	// Catppuccin Mocha (github.com/catppuccin/catppuccin).
	"catppuccin-mocha": mk("#1E1E2E", "#CDD6F4", "#6C7086", "#45475A", "#B4BEFE", "#CBA6F7",
		"#313244", "#181825", "#6C7086", "#FAB387", "#A6E3A1", "#F38BA8", "#F38BA8", "#F9E2AF",
		"#A6E3A1", "#F38BA8", "#CBA6F7", "#A6E3A1", "#FAB387", "#6C7086", "#89B4FA", "#89DCEB"),
	// Dracula (draculatheme.com/contribute).
	"dracula": mk("#282A36", "#F8F8F2", "#6272A4", "#44475A", "#BD93F9", "#FF79C6", "#44475A",
		"#21222C", "#6272A4", "#FFB86C", "#50FA7B", "#FF5555", "#FF5555", "#F1FA8C", "#50FA7B",
		"#FF5555", "#FF79C6", "#F1FA8C", "#BD93F9", "#6272A4", "#50FA7B", "#FF79C6"),
}
