package main

import (
	"bytes"
	"fmt"
	"image/png"
)

func convertImage(data []byte) []byte {
	bw := &bytes.Buffer{}
	bw.Write([]byte{
		66, 77, byte(1342 & 0xff), byte(1342 >> 8), 0, 0, 0, 0, 0, 0, 62, byte(62 >> 8),
		0, 0, 40, byte(40 >> 8), 0, 0, byte(256 & 0xff), byte(256 >> 8), 0, 0, 40, byte(40 >> 8),
		0, 0, 1, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, byte(255), byte(255), byte(255), 0,
	})

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		fmt.Printf("Error converting png to raw image: %s\n", err)
		return nil
	}
	w := NewBitWriter(bw)
	for y := 0; y < 40; y++ {
		for x := 0; x < 256; x++ {
			_, b, _, _ := img.At(x, y).RGBA()
			if b == 65535 {
				w.WriteBits(1, 1)
			} else {
				w.WriteBits(0, 1)
			}
		}
	}
	w.Close()

	return bw.Bytes()
}
