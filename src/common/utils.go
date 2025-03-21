package common

import (
	"os"
	"path/filepath"
)

// Expand paths like ~/ to absolute paths
func ExpandPath(path string) (string, error) {
	if len(path) > 0 {
		if path[0] == '~' {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			return filepath.Join(home, path[1:]), nil
		}
		return filepath.Abs(path)
	}
	return path, nil
}
