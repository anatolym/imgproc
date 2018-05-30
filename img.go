package imgproc

import (
	"bytes"
	"fmt"
	"image"
	// Initializing packages for supporting GIF, JPEG and PNG formats.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

// Color represents a color of an image.
type Color struct {
	Hex   string
	Count uint
}

// ColorList represents slice of Colors.
type ColorList []Color

// insert inserts new `color` on position `key` with shifting to the right.
func (cl ColorList) insert(color Color, key int) {
	for i := len(cl) - 1; i > key; i-- {
		cl[i] = cl[i-1]
	}
	cl[key] = color
}

// rgbSet represents RGB color.
type rgbSet [3]byte

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
	colors := make(map[rgbSet]uint)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rgb := rgbSet{byte(r >> 8), byte(g >> 8), byte(b >> 8)}
			colors[rgb]++
		}
	}
	var size int
	if len(colors) >= n {
		size = n
	} else {
		size = len(colors)
	}

	cl := make(ColorList, size)
	for rgb, count := range colors {
		for key := size - 1; key >= 0; key-- {
			if cl[key].Count > count {
				if key < size-1 {
					cl.insert(Color{toHex(rgb), count}, key+1)
				}
				break
			} else if key == 0 {
				cl.insert(Color{toHex(rgb), count}, 0)
			}
		}
	}

	return cl, nil
}

// toHex converts array of RGB byte values to HEX string.
func toHex(rgb rgbSet) string {
	return fmt.Sprintf("%02X%02X%02X", rgb[0], rgb[1], rgb[2])
}
