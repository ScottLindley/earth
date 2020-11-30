package animation

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

// BuildVideo - creates a mp4 from all the PNGs in the images directory
func BuildVideo() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("error getting working directory: ", err)
	}
	args := strings.Split("-framerate 60 -pattern_type glob -i "+dir+"/images/*.png -c:v libx264 -pix_fmt yuv420p earth.mp4 -y", " ")
	cmd := exec.Command("ffmpeg", args...)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

// BuildGIF - creates a gif from all the PNGs in the images directory
func BuildGIF() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("error getting working directory: ", err)
	}
	args := strings.Split("-pattern_type glob -i "+dir+"/images/*.png earth.gif -y", " ")
	cmd := exec.Command("ffmpeg", args...)
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
