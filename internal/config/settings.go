package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
)

// Duration is a time.Duration that reads and writes as a string such as "30s".
type Duration struct{ time.Duration }

// String formats the duration like time.Duration.
func (d Duration) String() string { return d.Duration.String() }

// MarshalText implements encoding.TextMarshaler.
func (d Duration) MarshalText() ([]byte, error) { return []byte(d.String()), nil }

// UnmarshalText implements encoding.TextUnmarshaler.
func (d *Duration) UnmarshalText(b []byte) error {
	v, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	d.Duration = v
	return nil
}

// Settings is the content of settings.toml.
type Settings struct {
	Theme          string            `toml:"theme"`
	VimMode        bool              `toml:"vim_mode"`
	PageSize       int               `toml:"page_size"`
	QueryTimeout   Duration          `toml:"query_timeout"`
	Zebra          bool              `toml:"zebra"`
	FKDisplay      bool              `toml:"fk_display"`
	Mouse          bool              `toml:"mouse"`
	ExternalEditor string            `toml:"external_editor"`
	DisplayColumns map[string]string `toml:"display_columns"`
	ThemeOverrides map[string]string `toml:"theme_overrides"`
}

// DefaultSettings returns the settings used when settings.toml is absent or partial.
func DefaultSettings() Settings {
	return Settings{
		Theme:          "vanitas",
		VimMode:        true,
		PageSize:       500,
		Zebra:          true,
		FKDisplay:      true,
		Mouse:          true,
		DisplayColumns: map[string]string{},
		ThemeOverrides: map[string]string{},
	}
}

const maxPageSize = 100000

// fileSettings shadows query_timeout with a string so the error can name the bad value.
type fileSettings struct {
	Settings
	QueryTimeout *string `toml:"query_timeout"`
}

// LoadSettings reads path. A missing file gives the defaults. The second result lists warnings.
func LoadSettings(path string) (Settings, []string, error) {
	base := filepath.Base(path)
	data, err := os.ReadFile(filepath.Clean(path))
	if errors.Is(err, os.ErrNotExist) {
		return DefaultSettings(), nil, nil
	}
	if err != nil {
		return Settings{}, nil, fmt.Errorf("%s: %w", base, err)
	}
	fs := fileSettings{Settings: DefaultSettings()}
	md, err := toml.NewDecoder(bytes.NewReader(data)).Decode(&fs)
	if err != nil {
		return Settings{}, nil, fmt.Errorf("%s: %w", base, err)
	}
	s := fs.Settings
	if fs.QueryTimeout != nil {
		var d Duration
		if err := d.UnmarshalText([]byte(*fs.QueryTimeout)); err != nil || d.Duration < 0 {
			return Settings{}, nil, fmt.Errorf("%s: invalid query_timeout %q", base, *fs.QueryTimeout)
		}
		s.QueryTimeout = d
	}
	if s.PageSize < 1 || s.PageSize > maxPageSize {
		return Settings{}, nil, fmt.Errorf("%s: page_size must be between 1 and %d, got %d", base, maxPageSize, s.PageSize)
	}
	var warnings []string
	und := md.Undecoded()
	for i, k := range und {
		// A table is reported through its leaf keys only.
		if i+1 < len(und) && len(und[i+1]) > len(k) && und[i+1][:len(k)].String() == k.String() {
			continue
		}
		warnings = append(warnings, fmt.Sprintf("%s: unknown key %q", base, k.String()))
	}
	return s, warnings, nil
}

// SaveSettings writes s to path atomically with mode 0600. Comments in an existing file are lost.
func SaveSettings(path string, s Settings) error {
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(s); err != nil {
		return fmt.Errorf("encode settings: %w", err)
	}
	return WriteFileAtomic(path, buf.Bytes(), 0o600)
}
