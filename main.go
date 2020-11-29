package main

import (
	"context"
	"earth/interpolation"
	"earth/nasa"
	"earth/shared"
	"fmt"
	"image"
	"image/gif"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/andybons/gogif"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	handleSigInt(cancel)
	ims := nasa.DownloadImages(ctx)
	filePaths := interpolation.InterpolateImages(ctx, ims)
	done := buildGIF(ctx, filePaths)
	<-done
}

func buildGIF(ctx context.Context, filePaths <-chan string) <-chan bool {
	done := make(chan bool)

	go func() {
		defer close(done)

		quantizer := gogif.MedianCutQuantizer{NumColor: 64}
		outGif := &gif.GIF{}

		writeGIFAndClose := func() {
			log.Println("writing the gif!...")
			if len(outGif.Image) > 0 {
				f, err := os.Create("earth.gif")
				if err != nil {
					log.Fatal("error creating file for GIF", err)
				}

				err = gif.EncodeAll(f, outGif)
				if err != nil {
					log.Fatal("error writing GIF", err)
				}
				f.Close()
				log.Println("wrote the gif!")
			}
			done <- true
		}

		for {
			select {
			case <-ctx.Done():
				writeGIFAndClose()
				return
			case filePath, ok := <-filePaths:
				if !ok {
					writeGIFAndClose()
					return
				}

				log.Println("file path:", filePath)

				img, err := shared.LoadImage(filePath)
				if err != nil {
					log.Fatal("error loading image for GIF frame", err)
				}

				bounds := img.Bounds()
				palettedImage := image.NewPaletted(bounds, nil)
				quantizer.Quantize(palettedImage, bounds, img, image.ZP)
				outGif.Image = append(outGif.Image, palettedImage)
				outGif.Delay = append(outGif.Delay, 0)
			}
		}
	}()

	return done
}

func handleSigInt(cancel context.CancelFunc) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT)
	go func() {
		<-sigc
		fmt.Println("SIGINT acknowledged, draining pipeline...")
		cancel()
	}()
}
