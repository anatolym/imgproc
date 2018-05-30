package imgproc

import (
	"bytes"
	"fmt"
	"image"
	"sort"
	// Initializing packages for supporting GIF, JPEG and PNG formats.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// Color represents a color of an image.
type Color struct {
	Hex   string
	Count int
}

// ColorList represents slice of Colors.
type ColorList []Color

func (cl ColorList) Len() int           { return len(cl) }
func (cl ColorList) Less(i, j int) bool { return cl[i].Count > cl[j].Count }
func (cl ColorList) Swap(i, j int)      { cl[i], cl[j] = cl[j], cl[i] }

// Analyze performs image analysis and returns slice of top n most prevalent colors of an image.
func Analyze(data []byte, n int) (ColorList, error) {
	if n <= 0 {
		return nil, fmt.Errorf("Size of result slice should be above 0")
	}
	reader := bytes.NewReader(data)
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("image.Decode returns an error: %s", err)
	}

	bounds := img.Bounds()
	colors := make(map[string]int)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			hex := fmt.Sprintf("%02X%02X%02X", r>>8, g>>8, b>>8)
			colors[hex]++
		}
	}

	cl := make(ColorList, len(colors))
	i := 0
	for hex, count := range colors {
		cl[i] = Color{hex, count}
		i++
	}
	sort.Sort(cl)
	var size int
	if len(cl) >= n {
		size = n
	} else {
		size = len(cl)
	}
	return cl[0:size], nil
}
