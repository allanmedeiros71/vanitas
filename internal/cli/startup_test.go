package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/allanmedeiros71/vanitas/internal/config"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestStartupDefaultsCreateNothing(t *testing.T) {
	t.Parallel()
	paths := config.PathsIn(t.TempDir())
	opts, closeLog, err := startupIn(paths, "", false)
	if err != nil {
		t.Fatal(err)
	}
	closeLog()
	if opts.Theme.Name != "vanitas" || opts.Registry == nil || len(opts.Warnings) != 0 {
		t.Fatalf("unexpected options %+v", opts)
	}
	if entries, _ := os.ReadDir(paths.Dir); len(entries) != 0 {
		t.Fatalf("startup created files: %v", entries)
	}
}

func TestStartupUnknownThemeFallsBack(t *testing.T) {
	t.Parallel()
	paths := config.PathsIn(t.TempDir())
	write(t, paths.Settings, `theme = "solarized"`)
	opts, _, err := startupIn(paths, "", true)
	if err != nil {
		t.Fatal(err)
	}
	want := `unknown theme "solarized", using "vanitas"`
	if opts.Theme.Name != "vanitas" || !slices.Contains(opts.Warnings, want) || !opts.NoColor {
		t.Fatalf("theme %q warnings %q", opts.Theme.Name, opts.Warnings)
	}
}

func TestStartupCollectsWarnings(t *testing.T) {
	t.Parallel()
	paths := config.PathsIn(t.TempDir())
	write(t, paths.Settings, "bogus = 1\n[theme_overrides]\nnope = \"#000000\"\n")
	write(t, paths.Keymap, "[global]\nhelp = \"ctrl+q\"\n")
	opts, _, err := startupIn(paths, "", false)
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(opts.Warnings, "|")
	for _, want := range []string{"unknown key", "conflicts with \"quit\"", "unknown token"} {
		if !strings.Contains(joined, want) {
			t.Errorf("warnings %q lack %q", joined, want)
		}
	}
}

func TestStartupInvalidFilesFail(t *testing.T) {
	t.Parallel()
	paths := config.PathsIn(t.TempDir())
	write(t, paths.Settings, "page_size = = 3")
	if _, _, err := startupIn(paths, "debug", false); err == nil || !strings.Contains(err.Error(), "settings.toml") {
		t.Fatalf("settings error %v", err)
	}
	if got, _ := os.ReadFile(paths.Settings); string(got) != "page_size = = 3" {
		t.Fatal("settings file was modified")
	}
	paths = config.PathsIn(t.TempDir())
	write(t, paths.Keymap, "[global\n")
	if _, _, err := startupIn(paths, "", false); err == nil || !strings.Contains(err.Error(), "keymap.toml") {
		t.Fatalf("keymap error %v", err)
	}
}

func TestDebugLogFile(t *testing.T) {
	paths := config.PathsIn(filepath.Join(t.TempDir(), "nested"))
	write2 := func() {
		if err := os.MkdirAll(paths.Dir, 0o700); err != nil {
			t.Fatal(err)
		}
		write(t, paths.Settings, "bogus = 1\n[display_columns]\n\"a.b\" = \"SECRETVALUE\"\n")
	}
	write2()
	_, closeLog, err := startupIn(paths, "debug", false)
	if err != nil {
		t.Fatal(err)
	}
	closeLog()
	data, err := os.ReadFile(paths.Log)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "startup warning") || strings.Contains(string(data), "SECRETVALUE") {
		t.Fatalf("unexpected log content %q", data)
	}
	if runtime.GOOS != "windows" {
		if fi, err := os.Stat(paths.Log); err != nil || fi.Mode().Perm() != 0o600 {
			t.Fatalf("log mode %v (%v)", fi.Mode(), err)
		}
	}
}

func TestDebugLogCreatesDirectory(t *testing.T) {
	paths := config.PathsIn(filepath.Join(t.TempDir(), "new"))
	_, closeLog, err := startupIn(paths, "debug", false)
	if err != nil {
		t.Fatal(err)
	}
	closeLog()
	if _, err := os.Stat(paths.Log); err != nil {
		t.Fatal(err)
	}
}

func TestDebugLogFailure(t *testing.T) {
	t.Parallel()
	blocker := filepath.Join(t.TempDir(), "file")
	write(t, blocker, "x")
	if _, _, err := startupIn(config.PathsIn(filepath.Join(blocker, "sub")), "debug", false); err == nil {
		t.Fatal("expected a log directory error")
	}
}

func TestStartupResolvesFromEnvironment(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("VANITAS_CONFIG_DIR", dir)
	t.Setenv("VANITAS_LOG", "")
	t.Setenv("NO_COLOR", "1")
	opts, closeLog, err := Startup()
	if err != nil {
		t.Fatal(err)
	}
	closeLog()
	if opts.Paths.Dir != dir || !opts.NoColor {
		t.Fatalf("paths %q noColor %v", opts.Paths.Dir, opts.NoColor)
	}
}

func TestRootCommandInvalidSettings(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("VANITAS_CONFIG_DIR", dir)
	write(t, filepath.Join(dir, "settings.toml"), "page_size = = 3")
	_, errOut, err := run(t, BuildInfo{})
	if err == nil || !strings.Contains(errOut, "settings.toml") {
		t.Fatalf("err %v, stderr %q", err, errOut)
	}
}
