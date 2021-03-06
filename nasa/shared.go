package nasa

import (
	"earth/shared"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
)

const layoutISO = "2006-01-02"

func mkdirIfNotExists(dir string) error {
	if !shared.FileExists(dir) {
		if err := os.Mkdir(dir, 0777); err != nil {
			return err
		}
	}
	return nil
}

func createErrorFromResponsebody(r *http.Response) error {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return errors.New(string(b))
}
