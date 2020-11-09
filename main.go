package main

import (
	"context"
	"earth/interpolation"
	"earth/nasa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// ctx, cancel := context.WithCancel(context.Background())
	// handleSigInt(cancel)
	// <-nasa.DownloadImages(ctx)

	b, _ := ioutil.ReadFile("metadata/2020-10-31.json")
	ims := []nasa.ImageMeta{}
	json.Unmarshal(b, &ims)

	lng1 := ims[0].CentroidCoordinates.Lng
	lng2 := ims[1].CentroidCoordinates.Lng

	lng := (lng1 + lng2) / 2

	err := interpolation.GenerateFrame(ims[0], ims[1], lng)
	if err != nil {
		log.Println(err)
	}
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
