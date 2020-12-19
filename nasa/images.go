package nasa

import (
	"context"
	"earth/shared"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func downloadImages(ctx context.Context, imageMetas <-chan ImageMeta) <-chan ImageMeta {
	out := make(chan ImageMeta)

	if err := mkdirIfNotExists(imageFileDir); err != nil {
		log.Fatal("Error creating images directory", err)
	}

	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case im, ok := <-imageMetas:
				if !ok {
					return
				}

				if err := downloadImageIfNotExists(im); err != nil {
					log.Println("error downloading image", im.Date, err)
				}
				out <- im
			}
		}
	}()

	return out
}

const imageFileDir = "images"

func downloadImageIfNotExists(im ImageMeta) error {
	path := BuildImageFilePath(im)

	if shared.FileExists(path) {
		return nil
	}

	datePath := strings.Split(strings.ReplaceAll(im.Date, "-", "/"), " ")[0]
	url := "https://epic.gsfc.nasa.gov/archive/natural/" + datePath + "/png/" + im.Image + ".png"

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return createErrorFromResponsebody(resp)
	}

	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(f, resp.Body)
	return err
}

// BuildImageFilePath - given image meta data, create the files system path
func BuildImageFilePath(im ImageMeta) string {
	return imageFileDir + "/" + im.Date + "_000.png"
}
