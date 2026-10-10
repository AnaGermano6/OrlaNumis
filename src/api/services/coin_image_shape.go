package services

import (
	"bytes"
	"image"
	_ "image/jpeg" // decoders for the shapes checked
	_ "image/png"
)

// LooksLikeCoin tells whether a product photo shows a round coin on a plain
// background rather than its box or blister: the photo's subject must be about
// as wide as high and narrow towards its top and bottom, as a disc is.
func LooksLikeCoin(data []byte) bool {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return false
	}
	bounds := img.Bounds()
	step := max(1, max(bounds.Dx(), bounds.Dy())/160)
	background := func(x, y int) bool {
		r, g, b, a := img.At(x, y).RGBA()
		return a < 0x2000 || (r > 0xEB00 && g > 0xEB00 && b > 0xEB00)
	}
	minX, minY, maxX, maxY := bounds.Max.X, bounds.Max.Y, bounds.Min.X-1, bounds.Min.Y-1
	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			if !background(x, y) {
				minX, maxX, minY, maxY = min(minX, x), max(maxX, x), min(minY, y), max(maxY, y)
			}
		}
	}
	width, height := maxX-minX+1, maxY-minY+1
	if width < bounds.Dx()/4 || height < bounds.Dy()/4 {
		return false
	}
	if aspect := float64(width) / float64(height); aspect < 0.85 || aspect > 1.18 {
		return false
	}
	// How wide the subject is at a given height, from its edges.
	span := func(fraction float64) int {
		y := minY + int(fraction*float64(height-1))
		left, right := -1, -1
		for x := minX; x <= maxX; x++ {
			if !background(x, y) {
				if left < 0 {
					left = x
				}
				right = x
			}
		}
		if left < 0 {
			return 0
		}
		return right - left + 1
	}
	middle := float64(max(span(0.5), 1))
	// A disc is 60% as wide at 10% of its height; a rectangle stays as wide.
	return float64(span(0.1))/middle < 0.85 && float64(span(0.9))/middle < 0.85
}
