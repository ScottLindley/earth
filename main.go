package main

import (
	"context"
	"earth/nasa"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	handleSigInt(cancel)
	<-nasa.DownloadImages(ctx)
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
