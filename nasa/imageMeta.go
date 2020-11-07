package nasa

import (
	"context"
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
	Image string
	Date  string
}

// There are no more available images before this date.
const finalEndDate = "2015-06-13"

// The directory where we'll write our files
const cacheDir = "metadata"

func getImageMetaForDate(date string) ([]ImageMeta, error) {
	// fmt.Println(date)

	var b []byte
	var err error
	path := cacheDir + "/" + date
	if fileExists(path) {
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
	path := cacheDir + "/" + date

	if err := mkdirIfNotExists(cacheDir); err != nil {
		return err
	}

	if fileExists(path) {
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

func generateImageMeta(ctx context.Context, startDate string) <-chan ImageMeta {
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
				if date == finalEndDate {
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
