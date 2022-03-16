package main

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"image"
	"image/color"
	"strconv"
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

func writeBmpIntoTheBody(fragmentImgBg *image.NRGBA, c *gin.Context) error {
	buf := new(bytes.Buffer)
	err := Encode(buf, fragmentImgBg)
	if err != nil {
		return err
	}

	c.Header("Content-Length", strconv.Itoa(len(buf.Bytes())))
	c.Data(200, "image/bmp", buf.Bytes())
	return nil
}
