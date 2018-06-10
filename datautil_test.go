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
	if _, err := imgproc.NewFileSrc("testdata/some_not_existing_file.txt", 1); err == nil ||
		!strings.Contains(string(err.Error()), "not exists") {
		t.Errorf("Got error '%s', expected error 'not exists'", err)
	}
}
func TestNewFileSrc(t *testing.T) {
	_, err := imgproc.NewFileSrc("testdata/url_list.txt", 1)
	if err != nil {
		t.Errorf("Got error '%s', expected nil", err)
	}
}

func TestNewCsvResult(t *testing.T) {
	resultFile := path.Join(os.TempDir(), "tst.csv")
	if _, err := imgproc.NewCsvResult(resultFile, 0); err == nil ||
		!strings.Contains(string(err.Error()), "Column number should be above 0") {
		t.Errorf("Got error '%s', expected error 'Column number should be above 0'", err)
	}
	if _, err := imgproc.NewCsvResult(resultFile, -1); err == nil ||
		!strings.Contains(string(err.Error()), "Column number should be above 0") {
		t.Errorf("Got error '%s', expected error 'Column number should be above 0'", err)
	}
}

func TestCsvResult(t *testing.T) {
	resultFile := path.Join(os.TempDir(), "tst.csv")
	res := imgproc.ResultItem{
		Name: "http://example.com/123",
		Results: imgproc.ColorList{
			imgproc.Color{Hex: "FF0000", Count: 1},
			imgproc.Color{Hex: "00FF00", Count: 1},
			imgproc.Color{Hex: "0000FF", Count: 1}},
	}

	cases := []struct {
		colNum      int
		contentWant string
	}{
		{3, "http://example.com/123,#FF0000,#00FF00,#0000FF,\n"},
		{1, "http://example.com/123,#FF0000,\n"},
		{5, "http://example.com/123,#FF0000,#00FF00,#0000FF,,,\n"},
	}

	for _, c := range cases {
		csv, err := imgproc.NewCsvResult(resultFile, c.colNum)
		if err != nil {
			t.Errorf("Got error '%s', expected nil", err)
		}
		if err := csv.Add(&res); err != nil {
			t.Errorf("Got error '%s', expected nil", err)
		}

		content, err := ioutil.ReadFile(resultFile)
		if err != nil {
			t.Errorf("Got error '%s', expected nil", err)
		}
		if string(content) != c.contentWant {
			t.Errorf("Got content '%s', wanted '%s'", content, c.contentWant)
		}

		if err := os.Remove(resultFile); err != nil {
			t.Errorf("Got error '%s', expected nil", err)
		}
	}
}
