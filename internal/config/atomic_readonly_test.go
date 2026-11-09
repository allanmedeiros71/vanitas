//go:build unix

package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestWriteFileAtomicReadOnlyDirKeepsOriginal(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "f.toml")
	if err := os.WriteFile(path, []byte("orig"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o700) })

	err := WriteFileAtomic(path, []byte("new"), 0o600)
	if err == nil {
		t.Fatal("expected an error writing into a read-only directory")
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("error %q does not name the path", err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "orig" {
		t.Errorf("original changed: %q", got)
	}
	if names := entries(t, dir); !reflect.DeepEqual(names, []string{"f.toml"}) {
		t.Errorf("directory has %v", names)
	}
}
