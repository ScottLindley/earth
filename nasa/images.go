package nasa

import (
	"context"
	"earth/shared"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// helper for reporting number of images downloaded per second
func rate(count int, start time.Time) {
	// fmt.Printf("%f/s\n", float64(count)/time.Since(start).Seconds())
}

func downloadImages(ctx context.Context, imageMetas <-chan ImageMeta) <-chan ImageMeta {
	out := make(chan ImageMeta)

	count := 0
	start := time.Now()
	worker := func(wg *sync.WaitGroup) {
		defer wg.Done()
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
				count++
				rate(count, start)
				out <- im
			}
		}
	}

	go func() {
		defer close(out)
		// Make this syncronous for now to ensure images
		// are order correctly for interpolation.
		numWorkers := 1
		wg := sync.WaitGroup{}
		wg.Add(numWorkers)

		for i := 0; i < numWorkers; i++ {
			go worker(&wg)
		}
		wg.Wait()
	}()

	return out
}

const imageFileDir = "images"

func downloadImageIfNotExists(im ImageMeta) error {
	path := BuildImageFilePath(im)

	if err := mkdirIfNotExists(imageFileDir); err != nil {
		return err
	}

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
	return imageFileDir + "/" + im.Date + ".png"
}
