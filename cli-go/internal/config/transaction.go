package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/gofrs/flock"
)

var updateMu sync.Mutex

// Update performs a cross-process locked load-modify-save transaction.
func Update(mutator func(*Config) error) error {
	updateMu.Lock()
	defer updateMu.Unlock()
	if err := os.MkdirAll(ConfigDir(), 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	lockPath := ConfigPath() + ".lock"
	fileLock := flock.New(lockPath)
	if err := fileLock.Lock(); err != nil {
		return fmt.Errorf("locking config: %w", err)
	}
	defer fileLock.Unlock()
	if err := os.Chmod(lockPath, 0o600); err != nil {
		return fmt.Errorf("securing config lock: %w", err)
	}
	cfg, err := Load()
	if err != nil {
		return err
	}
	if err := mutator(cfg); err != nil {
		return err
	}
	return Save(cfg)
}
