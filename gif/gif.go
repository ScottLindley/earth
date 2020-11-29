package gif

import (
	"context"
	"earth/shared"
	"image"
	"image/gif"
	"log"
	"os"
	"sort"
	"sync"

	"github.com/andybons/gogif"
)

// Frame - a processed (quantized) frame
type Frame struct {
	img      *image.Paletted
	filePath string
}

func buildFrames(ctx context.Context, filePaths <-chan string) <-chan Frame {
	out := make(chan Frame)

	worker := func(wg *sync.WaitGroup, quantizer *gogif.MedianCutQuantizer) {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case filePath, ok := <-filePaths:
				if !ok {
					return
				}

				img, err := shared.LoadImage(filePath)
				if err != nil {
					log.Fatal("error loading image for GIF frame", err)
				}

				bounds := img.Bounds()
				palettedImage := image.NewPaletted(bounds, nil)
				quantizer.Quantize(palettedImage, bounds, img, image.Point{})
				out <- Frame{img: palettedImage, filePath: filePath}
			}
		}
	}

	go func() {
		defer close(out)

		quantizer := &gogif.MedianCutQuantizer{NumColor: 64}

		numWorkers := 10
		wg := &sync.WaitGroup{}
		wg.Add(numWorkers)

		for i := 0; i < numWorkers; i++ {
			go worker(wg, quantizer)
		}
		wg.Wait()
	}()

	return out
}

// Build - takes a stream of file paths and stitches the images into a single GIF
func Build(ctx context.Context, filePaths <-chan string) <-chan bool {
	frames := buildFrames(ctx, filePaths)
	done := make(chan bool)

	go func() {
		defer close(done)

		outGif := &gif.GIF{}

		seenFilePaths := []string{}
		filePathToImage := make(map[string]*image.Paletted)

		writeGIFAndClose := func() {
			if len(seenFilePaths) > 0 {
				f, err := os.Create("earth.gif")
				if err != nil {
					log.Fatal("error creating file for GIF", err)
				}

				sort.Strings(seenFilePaths)

				for _, filePath := range seenFilePaths {
					outGif.Image = append(outGif.Image, filePathToImage[filePath])
					outGif.Delay = append(outGif.Delay, 0)
				}

				err = gif.EncodeAll(f, outGif)
				if err != nil {
					log.Fatal("error writing GIF", err)
				}
				f.Close()
			}
			done <- true
		}

		for {
			select {
			case <-ctx.Done():
				writeGIFAndClose()
				return
			case frame, ok := <-frames:
				if !ok {
					writeGIFAndClose()
					return
				}

				log.Println("file path:", frame.filePath)

				seenFilePaths = append(seenFilePaths, frame.filePath)
				filePathToImage[frame.filePath] = frame.img
			}
		}
	}()

	return done
}
