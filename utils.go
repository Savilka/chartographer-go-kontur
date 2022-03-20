package main

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"golang.org/x/image/bmp"
	"image"
	"strconv"
)

func createBlackImage(width, height int) image.NRGBA {
	buf := make([]uint8, 4*width*height)
	var i int64
	for i = 3; i <= int64(height*width*4); i += 4 {
		buf[i] = 255
	}
	return image.NRGBA{
		Pix:    buf,
		Stride: 4 * width,
		Rect:   struct{ Min, Max image.Point }{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: width, Y: height}},
	}
}

func writeBmpIntoTheBody(fragmentImgBg *image.NRGBA, c *gin.Context) error {
	buf := new(bytes.Buffer)
	err := bmp.Encode(buf, fragmentImgBg)
	if err != nil {
		return err
	}

	c.Header("Content-Length", strconv.Itoa(len(buf.Bytes())))
	c.Data(200, "image/bmp", buf.Bytes())
	return nil
}
