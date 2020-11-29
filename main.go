package main

import (
	"context"
	"earth/gif"
	"earth/interpolation"
	"earth/nasa"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	handleSigInt(cancel)
	ims := nasa.DownloadImages(ctx)
	filePaths := interpolation.InterpolateImages(ctx, ims)
	done := gif.Build(ctx, filePaths)
	<-done
}

func handleSigInt(cancel context.CancelFunc) {
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT)
	go func() {
		<-sigc
		fmt.Println("\n\nSIGINT acknowledged, draining pipeline...")
		cancel()
	}()
}
