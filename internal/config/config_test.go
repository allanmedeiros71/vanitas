package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestResolveDir(t *testing.T) {
	t.Parallel()
	abs := func(p string) string {
		a, err := filepath.Abs(filepath.FromSlash(p))
		if err != nil {
			t.Fatal(err)
		}
		return a
	}
	h, x, ad, b := abs("/h"), abs("/x"), abs("/ad"), abs("/b")
	tests := []struct {
		name, goos, home string
		env              map[string]string
		want             string
		wantErr          bool
	}{
		{"override", "linux", h, map[string]string{"VANITAS_CONFIG_DIR": filepath.Join(abs("/a"), "..", "b"), "XDG_CONFIG_HOME": x}, b, false},
		{"xdg", "linux", h, map[string]string{"XDG_CONFIG_HOME": x}, filepath.Join(x, "vanitas"), false},
		{"xdg relative", "linux", h, map[string]string{"XDG_CONFIG_HOME": "x"}, filepath.Join(h, ".config", "vanitas"), false},
		{"windows", "windows", h, map[string]string{"APPDATA": ad}, filepath.Join(ad, "vanitas"), false},
		{"windows no appdata", "windows", h, nil, "", true},
		{"darwin", "darwin", h, nil, filepath.Join(h, "Library", "Application Support", "vanitas"), false},
		{"linux", "linux", h, nil, filepath.Join(h, ".config", "vanitas"), false},
		{"no home", "linux", "", nil, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := resolveDir(tt.goos, func(k string) string { return tt.env[k] }, tt.home)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolvePathsUsesEnvAndCreatesNothing(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "conf")
	t.Setenv("VANITAS_CONFIG_DIR", dir)
	p, err := ResolvePaths()
	if err != nil {
		t.Fatal(err)
	}
	if p.Dir != dir || p.Settings != filepath.Join(dir, "settings.toml") ||
		p.Keymap != filepath.Join(dir, "keymap.toml") || p.Connections != filepath.Join(dir, "connections.toml") ||
		p.State != filepath.Join(dir, "state.db") || p.Log != filepath.Join(dir, "vanitas.log") {
		t.Errorf("unexpected paths %+v", p)
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Errorf("directory was created: %v", err)
	}
}

func entries(t *testing.T, dir string) []string {
	t.Helper()
	es, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, e := range es {
		names = append(names, e.Name())
	}
	return names
}

func TestWriteFileAtomic(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "sub")
	path := filepath.Join(dir, "f.toml")
	if err := WriteFileAtomic(path, []byte("one"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic(path, []byte("two"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "two" {
		t.Errorf("content %q", got)
	}
	if names := entries(t, dir); !reflect.DeepEqual(names, []string{"f.toml"}) {
		t.Errorf("directory has %v", names)
	}
}

func TestWriteFileAtomicFailureKeepsOriginal(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "f.toml")
	// The target is a non-empty directory, so the final rename fails even as root.
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	inner := filepath.Join(path, "orig")
	if err := os.WriteFile(inner, []byte("orig"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := WriteFileAtomic(path, []byte("new"), 0o600)
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("error %q does not name the path", err)
	}
	got, _ := os.ReadFile(inner)
	if string(got) != "orig" {
		t.Errorf("original changed: %q", got)
	}
	if names := entries(t, dir); !reflect.DeepEqual(names, []string{"f.toml"}) {
		t.Errorf("directory has %v", names)
	}
}

func TestWriteFileAtomicRenameFailureRemovesTemp(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	if err := os.Mkdir(target, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "x"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic(target, []byte("d"), 0o600); err == nil {
		t.Fatal("expected error renaming onto a non-empty directory")
	}
	if names := entries(t, dir); !reflect.DeepEqual(names, []string{"target"}) {
		t.Errorf("directory has %v", names)
	}
}

func TestWriteFileAtomicMkdirFailure(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteFileAtomic(filepath.Join(file, "child"), []byte("d"), 0o600); err == nil {
		t.Fatal("expected error")
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "settings.toml")
	want := Settings{
		Theme: "nord", VimMode: false, PageSize: 42, QueryTimeout: Duration{90 * time.Second},
		Zebra: false, FKDisplay: false, Mouse: false, ExternalEditor: "nvim",
		DisplayColumns: map[string]string{"public.users": "name"},
		ThemeOverrides: map[string]string{"accent": "#112233"},
	}
	if err := SaveSettings(path, want); err != nil {
		t.Fatal(err)
	}
	got, warns, err := LoadSettings(path)
	if err != nil || len(warns) != 0 {
		t.Fatalf("err=%v warnings=%v", err, warns)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestLoadSettings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, content, wantErr string
		wantWarn               []string
	}{
		{name: "unknown nested", content: "[foo]\nbar = 1\n", wantWarn: []string{`settings.toml: unknown key "foo.bar"`}},
		{name: "syntax", content: "a = 1\nb = = 2\n", wantErr: "line 2"},
		{name: "page low", content: "page_size = 0\n", wantErr: "page_size must be between 1 and 100000, got 0"},
		{name: "page high", content: "page_size = 100001\n", wantErr: "got 100001"},
		{name: "negative timeout", content: "query_timeout = \"-5s\"\n", wantErr: `invalid query_timeout "-5s"`},
		{name: "bad timeout", content: "query_timeout = \"abc\"\n", wantErr: `invalid query_timeout "abc"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "settings.toml")
			if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
				t.Fatal(err)
			}
			_, warns, err := LoadSettings(path)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) || !strings.Contains(err.Error(), "settings.toml") {
					t.Fatalf("err = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(warns, tt.wantWarn) {
				t.Errorf("warnings = %v, want %v", warns, tt.wantWarn)
			}
		})
	}
}

func TestLoadSettingsReadError(t *testing.T) {
	t.Parallel()
	if _, _, err := LoadSettings(t.TempDir()); err == nil {
		t.Fatal("expected error reading a directory")
	}
}

func TestLoadSettingsMissingGivesDefaults(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	s, w, err := LoadSettings(filepath.Join(dir, "settings.toml"))
	if err != nil || len(w) != 0 || !reflect.DeepEqual(s, DefaultSettings()) {
		t.Fatalf("s=%+v w=%v err=%v", s, w, err)
	}
	if names := entries(t, dir); len(names) != 0 {
		t.Errorf("created %v", names)
	}
}
