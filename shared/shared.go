package shared

import "os"

// FileExists -
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
