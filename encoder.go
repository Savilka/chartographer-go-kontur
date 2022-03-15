package main

import (
	"encoding/binary"
	"errors"
	"image"
	"io"
)

type header struct {
	sigBM           [2]byte
	fileSize        uint32
	resverved       [2]uint16
	pixOffset       uint32
	dibHeaderSize   uint32
	width           uint32
	height          uint32
	colorPlane      uint16
	bpp             uint16
	compression     uint32
	imageSize       uint32
	xPixelsPerMeter uint32
	yPixelsPerMeter uint32
	colorUse        uint32
	colorImportant  uint32
}

func encodeNRGBA(w io.Writer, pix []uint8, dx, dy, stride, step int) error {
	buf := make([]byte, step)
	for y := dy - 1; y >= 0; y-- {
		min := y*stride + 0
		max := y*stride + dx*4
		off := 0
		for i := min; i < max; i += 4 {
			buf[off+2] = pix[i+0]
			buf[off+1] = pix[i+1]
			buf[off+0] = pix[i+2]
			off += 3
		}
		if _, err := w.Write(buf); err != nil {
			return err
		}
	}

	return nil
}

// Encode writes the image m to w in BMP format.
func Encode(w io.Writer, m *image.NRGBA) error {
	d := m.Bounds().Size()
	if d.X < 0 || d.Y < 0 {
		return errors.New("bmp: negative bounds")
	}
	h := &header{
		sigBM:         [2]byte{'B', 'M'},
		fileSize:      14 + 40,
		pixOffset:     14 + 40,
		dibHeaderSize: 40,
		width:         uint32(d.X),
		height:        uint32(d.Y),
		colorPlane:    1,
	}

	var step int
	var palette []byte

	step = (3*d.X + 3) &^ 3
	h.bpp = 24

	h.imageSize = uint32(d.Y * step)
	h.fileSize += h.imageSize

	if err := binary.Write(w, binary.LittleEndian, h); err != nil {
		return err
	}
	if palette != nil {
		if err := binary.Write(w, binary.LittleEndian, palette); err != nil {
			return err
		}
	}

	if d.X == 0 || d.Y == 0 {
		return nil
	}

	return encodeNRGBA(w, m.Pix, d.X, d.Y, m.Stride, step)
}
