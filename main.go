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

	ctx, cancel := context.WithCancel(context.Background())
	handleSigInt(cancel)

	ims := nasa.DownloadImages(ctx, startDate, endDate)
	done := interpolation.InterpolateImages(ctx, ims)
	<-done

	wg := &sync.WaitGroup{}
	wg.Add(2)

	log.Println("Generating video and gif...")

	go func() {
		animation.BuildVideo()
		log.Println("video done!")
		wg.Done()
	}()
	go func() {
		animation.BuildGIF()
		log.Println("gif done!")
		wg.Done()
	}()

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
