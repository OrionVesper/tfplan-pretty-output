package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

func StorageRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not determine home directory: %w", err)
	}
	root := filepath.Join(home, ".tfplan-pretty-output")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", fmt.Errorf("could not create %s: %w", root, err)
	}
	return root, nil
}
