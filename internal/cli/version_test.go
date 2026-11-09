package cli

import (
	"bytes"
	"strings"
	"testing"
)

func run(t *testing.T, info BuildInfo, args ...string) (string, string, error) {
	t.Helper()
	cmd := NewRootCommand(info)
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), errOut.String(), err
}

func TestVersionOutput(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, BuildInfo{"1.2.3", "abc", "2026-01-02"}, "version")
	if err != nil {
		t.Fatal(err)
	}
	if want := "vanitas 1.2.3 (abc, 2026-01-02)\n"; out != want {
		t.Fatalf("got %q, want %q", out, want)
	}
}

func TestVersionRejectsArgs(t *testing.T) {
	t.Parallel()
	if _, _, err := run(t, BuildInfo{"dev", "none", "unknown"}, "version", "x"); err == nil {
		t.Fatal("expected error")
	}
}

func TestUnknownCommand(t *testing.T) {
	t.Parallel()
	_, _, err := run(t, BuildInfo{}, "frobnicate")
	if err == nil || !strings.Contains(err.Error(), `unknown command "frobnicate"`) {
		t.Fatalf("unexpected error %v", err)
	}
}

func TestRootPrintsHelp(t *testing.T) {
	t.Parallel()
	out, _, err := run(t, BuildInfo{}, "--help")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "version") {
		t.Fatalf("help does not list version: %q", out)
	}
}

func TestExecuteExitCodes(t *testing.T) {
	t.Parallel()
	var out, errOut bytes.Buffer
	if code := execute(BuildInfo{"v", "c", "d"}, []string{"version"}, &out, &errOut); code != 0 {
		t.Fatalf("version exit code %d", code)
	}
	if out.String() != "vanitas v (c, d)\n" {
		t.Fatalf("got %q", out.String())
	}
	errOut.Reset()
	if code := execute(BuildInfo{}, []string{"nope"}, &out, &errOut); code != 1 {
		t.Fatalf("unknown exit code %d", code)
	}
	if !strings.HasPrefix(errOut.String(), "Error: unknown command") {
		t.Fatalf("got %q", errOut.String())
	}
}
