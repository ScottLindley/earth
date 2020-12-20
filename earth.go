package main

import (
	"context"
	"earth/animation"
	"earth/interpolation"
	"earth/nasa"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	// TODO: maybe make these arguments?
	startDate := "2018-09-06"
	endDate := "2018-09-08"
	// The GIF takes a crazy long time to finish and is over 1GB
	// for the dates specified above. So we'll default to video only.
	generateGIF := false

	ctx, cancel := context.WithCancel(context.Background())
	handleSigInt(cancel)

	ims := nasa.DownloadImages(ctx, startDate, endDate)
	done := interpolation.InterpolateImages(ctx, ims)
	<-done

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		log.Println("Generating video...")
		animation.BuildVideo()
		log.Println("video done!")
		wg.Done()
	}()

	if generateGIF {
		wg.Add(1)
		go func() {
			log.Println("Generating gif...")
			animation.BuildGIF()
			log.Println("gif done!")
			wg.Done()
		}()
	}

	wg.Wait()

	log.Println("Done!")
}

func handleSigInt(cancel context.CancelFunc) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT)
	go func() {
		<-sigc
		log.Println("\nSIGINT acknowledged, draining pipeline...")
		cancel()
	}()
}
