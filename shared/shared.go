package shared

import (
	"image"
	"os"
)

// FileExists -
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// LoadImage - loads an image into memory
func LoadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}
