package services

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// syntheticPhoto draws a grey subject on a white background.
func syntheticPhoto(t *testing.T, width, height int, inside func(x, y int) bool) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := color.RGBA{255, 255, 255, 255}
			if inside(x, y) {
				c = color.RGBA{150, 140, 120, 255}
			}
			img.Set(x, y, c)
		}
	}
	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}

func TestLooksLikeCoin(t *testing.T) {
	disc := syntheticPhoto(t, 200, 200, func(x, y int) bool { return (x-100)*(x-100)+(y-100)*(y-100) < 80*80 })
	box := syntheticPhoto(t, 200, 200, func(x, y int) bool { return x > 30 && x < 170 && y > 20 && y < 180 })
	card := syntheticPhoto(t, 200, 200, func(x, y int) bool { return x > 20 && x < 180 && y > 20 && y < 180 })
	blank := syntheticPhoto(t, 200, 200, func(int, int) bool { return false })
	for name, c := range map[string]struct {
		data []byte
		want bool
	}{"disc": {disc, true}, "box": {box, false}, "card": {card, false}, "blank": {blank, false}, "not an image": {[]byte("no"), false}} {
		if got := LooksLikeCoin(c.data); got != c.want {
			t.Errorf("%s: LooksLikeCoin = %v, want %v", name, got, c.want)
		}
	}
}
