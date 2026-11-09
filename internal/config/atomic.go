package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFileAtomic writes data to path through a temp file and a rename, so a failure
// never leaves a partial file. The parent directory is created with 0700 if missing.
func WriteFileAtomic(path string, data []byte, perm os.FileMode) (err error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create directory for %s: %w", path, err)
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file for %s: %w", path, err)
	}
	defer func() {
		if err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
		}
	}()
	if _, err = tmp.Write(data); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err = tmp.Sync(); err != nil {
		return fmt.Errorf("sync %s: %w", path, err)
	}
	if err = tmp.Close(); err != nil {
		return fmt.Errorf("close %s: %w", path, err)
	}
	if err = os.Chmod(tmp.Name(), perm); err != nil {
		return fmt.Errorf("chmod %s: %w", path, err)
	}
	if err = os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	return nil
}
