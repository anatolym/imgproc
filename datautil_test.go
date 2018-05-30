package imgproc_test

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/anatolym/imgproc"
)

func TestNewFileSrcErrors(t *testing.T) {
	if _, err := imgproc.NewFileSrc("testdata/some_not_existing_file.txt"); err == nil ||
		!strings.Contains(string(err.Error()), "not exists") {
		t.Errorf("Got error '%s', expected error 'not exists'", err)
	}
}
func TestNewFileSrc(t *testing.T) {
	src, err := imgproc.NewFileSrc("testdata/url_list.txt")
	if err != nil {
		t.Errorf("Got error '%s', expected nil", err)
	}
	defer src.Close()
}

func TestCsvResult(t *testing.T) {
	resultFile := path.Join(os.TempDir(), "tst.csv")
	url := "http://example.com/123"
	results := imgproc.ColorList{
		imgproc.Color{Hex: "FF0000", Count: 1},
		imgproc.Color{Hex: "00FF00", Count: 1},
		imgproc.Color{Hex: "0000FF", Count: 1},
	}
	csv, err := imgproc.NewCsvResult(resultFile)
	if err != nil {
		t.Errorf("Got error '%s', expected nil", err)
	}
	if err := csv.Add(url, results); err != nil {
		t.Errorf("Got error '%s', expected nil", err)
	}

	content, err := ioutil.ReadFile(resultFile)
	if err != nil {
		t.Errorf("Got error '%s', expected nil", err)
	}
	contentWant := "http://example.com/123,#FF0000,#00FF00,#0000FF,\n"
	if string(content) != contentWant {
		t.Errorf("Got content '%s', wanted '%s'", content, contentWant)
	}

	if err := os.Remove(resultFile); err != nil {
		t.Errorf("Got error '%s', expected nil", err)
	}
}
