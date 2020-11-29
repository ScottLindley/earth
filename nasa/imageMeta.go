package nasa

import (
	"context"
	"earth/shared"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// ImageMeta -
// Meta data that describes an available image taken on a particular date.
type ImageMeta struct {
	Image               string
	Date                string
	CentroidCoordinates struct {
		Lat float64 `json:"lat"`
		Lng float64 `json:"lon"`
	} `json:"centroid_coordinates"`
	SatellitePosition struct {
		X float64
		Y float64
		Z float64
	} `json:"dscovr_j2000_position"`
}

// The directory where we'll write our files
const cacheDir = "metadata"

func getImageMetaForDate(date string) ([]ImageMeta, error) {
	var b []byte
	var err error
	path := buildFilePath(date)
	if shared.FileExists(path) {
		b, err = ioutil.ReadFile(path)
	} else {
		resp, err := http.Get("https://epic.gsfc.nasa.gov/api/natural/date/" + date)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, createErrorFromResponsebody(resp)
		}

		b, err = ioutil.ReadAll(resp.Body)
	}

	if err = writeImageMetaIfNotExists(date, b); err != nil {
		return nil, err
	}

	ims := []ImageMeta{}
	if err = json.Unmarshal(b, &ims); err != nil {
		return nil, err
	}

	return ims, nil
}

func writeImageMetaIfNotExists(date string, data []byte) error {
	path := buildFilePath(date)

	if err := mkdirIfNotExists(cacheDir); err != nil {
		return err
	}

	if shared.FileExists(path) {
		return nil
	}

	f, err := os.Create(path)
	defer f.Close()
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	return err
}

func generateImageMeta(ctx context.Context, startDate, endDate string) <-chan ImageMeta {
	out := make(chan ImageMeta)

	go func() {
		defer close(out)
		date := startDate
		for {
			select {
			case <-ctx.Done():
				return
			default:
				ims, err := getImageMetaForDate(date)
				if err != nil {
					log.Fatal(err)
				}

				for _, im := range ims {
					out <- im
				}

				date, err = getPreviousDate(date)
				if err != nil {
					log.Fatal(err)
				}
				// We've reached the end, close the pipeline
				if date == endDate {
					return
				}
			}
		}
	}()

	return out
}

func getPreviousDate(date string) (string, error) {
	t, err := time.Parse(layoutISO, date)
	if err != nil {
		return "", err
	}
	t = t.Add(-time.Hour * 24)
	return t.Format(layoutISO), nil
}

func buildFilePath(date string) string {
	return cacheDir + "/" + date + ".json"
}
