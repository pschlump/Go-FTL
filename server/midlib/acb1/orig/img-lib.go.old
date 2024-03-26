package Acb1

import (
	"fmt"
	"image"
	"os"
)

// SizeOfImage returns the height and width of the image in pixels.
func SizeOfImg(fn string) (h, w float64) {
	if reader, err := os.Open(fn); err == nil {
		defer reader.Close()
		im, _, err := image.DecodeConfig(reader)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", fn, err)
			return
		}
		fmt.Printf("%s %d %d\n", fn, im.Width, im.Height)
		h, w = float64(im.Height), float64(im.Width)
		return
	} else {
		fmt.Println("Impossible to open the file:", err)
	}
	return
}
