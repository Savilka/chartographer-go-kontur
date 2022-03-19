package main

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"golang.org/x/image/bmp"
	"image"
	"strconv"
)

func createBlackImage(width, height int) *image.NRGBA {
	buf1 := make([]uint8, height*width*4)
	var i int64
	for i = 3; i <= int64(height*width*4); i += 4 {
		buf1[i] = 255
	}
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	img.Pix = buf1
	img.Stride = 4 * width

	return img
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
