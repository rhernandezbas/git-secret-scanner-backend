package cloner

import "os"

// Delete removes a cloned repo directory. Safe to call even if path doesn't exist.
func Delete(path string) error {
	if path == "" {
		return nil
	}
	return os.RemoveAll(path)
}
