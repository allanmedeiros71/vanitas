package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

func writeSettings(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "settings.toml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadSettingsPageSizeBoundsAccepted(t *testing.T) {
	t.Parallel()
	for _, n := range []string{"1", "100000"} {
		s, warns, err := LoadSettings(writeSettings(t, "page_size = "+n+"\n"))
		if err != nil || len(warns) != 0 {
			t.Fatalf("page_size %s: err=%v warnings=%v", n, err, warns)
		}
		if s.PageSize == 0 || s.PageSize > maxPageSize {
			t.Errorf("page_size %s loaded as %d", n, s.PageSize)
		}
	}
}

func TestLoadSettingsQueryTimeoutZeroAccepted(t *testing.T) {
	t.Parallel()
	s, _, err := LoadSettings(writeSettings(t, "query_timeout = \"0s\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if s.QueryTimeout.Duration != 0 {
		t.Errorf("query_timeout = %v, want 0", s.QueryTimeout.Duration)
	}
}

func TestLoadSettingsQueryTimeoutNotAString(t *testing.T) {
	t.Parallel()
	_, _, err := LoadSettings(writeSettings(t, "query_timeout = 5\n"))
	if err == nil || !strings.Contains(err.Error(), "settings.toml") {
		t.Fatalf("err = %v, want an error naming settings.toml", err)
	}
}

func TestLoadSettingsUnknownKeysOneWarningEach(t *testing.T) {
	t.Parallel()
	_, warns, err := LoadSettings(writeSettings(t, "colour = \"blue\"\nfoo = 1\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := []string{`settings.toml: unknown key "colour"`, `settings.toml: unknown key "foo"`}
	if !reflect.DeepEqual(warns, want) {
		t.Errorf("warnings = %v, want %v", warns, want)
	}
}

func TestSettingsRoundTripDefaultsAndDuration(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "settings.toml")
	want := DefaultSettings()
	want.QueryTimeout = Duration{2*time.Minute + 3*time.Second}
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
