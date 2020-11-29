package nasa

import (
	"context"
)

// DownloadImages -
// Walk backwards from today and fetch all images with their associated
// metadata and save to disk.
func DownloadImages(ctx context.Context) <-chan ImageMeta {
	startDate := "2018-09-06"
	endDate := "2018-09-08"
	return downloadImages(ctx, generateImageMeta(ctx, startDate, endDate))
}
