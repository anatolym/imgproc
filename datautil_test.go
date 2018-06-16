package imgproc_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"
	"time"

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

func prepFileServer(t *testing.T) *httptest.Server {
	ts := httptest.NewServer(http.FileServer(http.Dir("testdata")))

	// Checking test server returns correct data.
	res, err := http.Get(ts.URL + "/gopher.png")
	if err != nil {
		t.Fatalf("Got error '%s', expected nil", err)
	}
	respImg, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatalf("Got error '%s', expected nil", err)
	}
	if respImg == nil {
		t.Fatal("Got nil , expected non-empty response body")
	}
	fileImg, err := ioutil.ReadFile("testdata/gopher.png")
	if err != nil {
		t.Fatalf("Got error '%s', expected nil", err)
	}
	if !bytes.Equal(respImg, fileImg) {
		t.Fatal("Files are different")
	}
	return ts
}

func TestGetImgItemCh(t *testing.T) {
	ts := prepFileServer(t)
	defer ts.Close()
	files := []string{
		"gopher.png",
		"img-1-color.jpg",
		"img-3-colors-with-gradient.png",
		"img-4-colors.png",
		"img-gradient.png",
	}
	tmpfile, err := ioutil.TempFile(os.TempDir(), "example.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // clean up

	// Filling tmpfile with image urls.
	urls := make([]string, len(files))
	for i, fn := range files {
		urls[i] = ts.URL + "/" + fn
		io.WriteString(tmpfile, urls[i]+"\n")
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	src, e := imgproc.NewFileSrc(tmpfile.Name(), 1)
	if e != nil {
		t.Errorf("Got error '%s', expected nil", e)
	}

	// Checking image channel returns all the imagas
	done := make(chan struct{})
	defer close(done)
	readCh := func() map[string]struct{} {
		imgCh := src.GetImgItemCh(done)
		var lastImgName string
		timeoutC := time.After(1 * time.Second)
		gotUrls := map[string]struct{}{}
		for {
			select {
			case img, ok := <-imgCh:
				if !ok {
					return gotUrls
				}
				lastImgName = img.Name
				gotUrls[img.Name] = struct{}{}
				t.Logf("processing img '%s'", img.Name)
			case <-timeoutC:
				t.Errorf("timed out on img '%s'", lastImgName)
				return gotUrls
			}
		}
	}
	gotUrls := readCh()
	for _, url := range urls {
		if _, ok := gotUrls[url]; !ok {
			t.Errorf("Url '%s' did not show up in the channel", url)
		}
	}
}

func TestProcessItems(t *testing.T) {
	gopherImg, err := ioutil.ReadFile("testdata/gopher.png")
	if err != nil {
		t.Fatalf("Got error '%s', expected nil", err)
	}
	imgs := []imgproc.ImgItem{
		{Name: "name1", Data: gopherImg},
	}
	imgItemCh := make(imgproc.ImgItemCh)
	go func() {
		for _, img := range imgs {
			imgItemCh <- &img
		}
		close(imgItemCh)
	}()
	done := make(chan struct{})
	defer close(done)
	results := make(map[string]struct{})
	for res := range imgproc.ProcessItems(done, imgItemCh, 1, 1) {
		results[res.Name] = struct{}{}
	}
	for _, img := range imgs {
		if _, ok := results[img.Name]; !ok {
			t.Errorf("Img '%s' did not show up in the channel", img.Name)
		}
	}
}
