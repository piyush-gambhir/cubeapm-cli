//go:build !windows

package config

import (
	"os"

	"github.com/google/renameio/v2"
)

func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	return renameio.WriteFile(path, data, perm)
}
