package nasa

import (
	"context"
)

// DownloadImages -
// Walk backwards from today and fetch all images with their associated
// metadata and save to disk.
// This exported fn just glues the two steps togeher: download image meta, and download images
func DownloadImages(ctx context.Context, startDate, endDate string) <-chan ImageMeta {
	return downloadImages(ctx, generateImageMeta(ctx, startDate, endDate))
}
