//go:build unix

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileAtomicPermissions(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "sub")
	path := filepath.Join(dir, "f.toml")
	if err := WriteFileAtomic(path, []byte("one"), 0o600); err != nil {
		t.Fatal(err)
	}
	if fi, _ := os.Stat(dir); fi.Mode().Perm() != 0o700 {
		t.Errorf("dir perm %v", fi.Mode().Perm())
	}
	if fi, _ := os.Stat(path); fi.Mode().Perm() != 0o600 {
		t.Errorf("file perm %v", fi.Mode().Perm())
	}
}
