package imgproc_test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/anatolym/imgproc"
)

func TestImgAnalyzeErrorsN(t *testing.T) {
	cases := []int{0, -42}
	imgData := make([]byte, 0)
	for _, c := range cases {
		_, err := imgproc.Analyze(imgData, c)
		if err == nil {
			t.Error("Got nil, expected error")
		}
	}
}

func TestImgAnalyzeErrorsImgReader(t *testing.T) {
	cases := [][]byte{
		make([]byte, 0),
	}
	for _, c := range cases {
		if _, err := imgproc.Analyze(c, 1); err == nil ||
			!strings.Contains(string(err.Error()), "image.Decode returns an error") {
			t.Errorf("Got error '%s', expected error 'image.Decode returns an error'", err)
		}
	}
}

func TestImgAnalyzeErrorImgDecode(t *testing.T) {
	reader, err := os.Open("testdata/img-1-color-corrupted.jpg")
	if err != nil {
		t.Errorf("os.Open returns an error: %s", err)
	}
	defer reader.Close()
	imgData, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Errorf("ioutil.ReadAll returns an error: %s", err)
	}
	if _, err := imgproc.Analyze(imgData, 1); err == nil ||
		!strings.Contains(string(err.Error()), "image.Decode returns an error: invalid JPEG format") {
		t.Errorf("Got error '%s', expected error 'image.Decode returns an error: invalid JPEG format'", err)
	}
}
func TestImgAnalyze(t *testing.T) {
	cases := []struct {
		file      string
		nIn, nOut int
		hexes     []string
	}{
		{"testdata/gopher.png", 3, 3, []string{"000000", "74CEDC", "73CEDC"}},
		{"testdata/img-1-color.jpg", 5, 1, []string{"66CCFF"}},
		{"testdata/img-3-colors-with-gradient.png", 3, 3, []string{"FF0054", "22FF1E", "F9FFFB"}},
		{"testdata/img-4-colors.png", 5, 4, []string{"FFFF00", "FFFFFF", "0000FF", "00FF00"}},
	}
	for _, c := range cases {
		reader, err := os.Open(c.file)
		if err != nil {
			t.Errorf("os.Open returns an error: %s", err)
		}
		defer reader.Close()
		imgData, err := ioutil.ReadAll(reader)
		if err != nil {
			t.Errorf("ioutil.ReadAll returns an error: %s", err)
		}
		result, err := imgproc.Analyze(imgData, c.nIn)
		if err != nil {
			t.Errorf("Got error '%s', expected nil", err)
		}
		if (len(result)) != c.nOut {
			t.Errorf("Got result slice of length %d, expected %d", len(result), c.nOut)
		}
		t.Log(result)
		for key, color := range result {
			if color.Hex != c.hexes[key] {
				t.Errorf("Got color '%s' for the key %d, expected '%s'", color.Hex, key, c.hexes[key])
			}
		}
	}
}
func BenchmarkAnalyze(b *testing.B) {
	reader, err := os.Open("testdata/img-gradient.png")
	if err != nil {
		b.Errorf("os.Open returns an error: %s", err)
	}
	defer reader.Close()
	imgData, err := ioutil.ReadAll(reader)
	if err != nil {
		b.Errorf("ioutil.ReadAll returns an error: %s", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := imgproc.Analyze(imgData, 3)
		if err != nil {
			b.Errorf("Got error '%s', expected nil", err)
		}
		if (len(result)) != 3 {
			b.Errorf("Got result slice of length %d, expected %d", len(result), 3)
		}
	}
}
