package main

import (
	"image"
	"image/color"
)

func createBlackImage(width, height int) *image.NRGBA {
	out := image.NewNRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			out.Set(x, y, color.NRGBA{R: 0, G: 0, B: 0, A: 255})
		}
	}

	return out
}
