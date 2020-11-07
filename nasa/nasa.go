package nasa

import (
	"context"
	"time"
)

// DownloadImages -
// Walk backwards from today and fetch all images with their associated
// metadata and save to disk.
func DownloadImages(ctx context.Context) <-chan bool {
	startDate := getYesterdayDateString()
	return downloadImages(ctx, generateImageMeta(ctx, startDate))
}

func getYesterdayDateString() string {
	t := time.Now()
	t = t.Add(-time.Hour * 24)
	return t.Format(layoutISO)
}
