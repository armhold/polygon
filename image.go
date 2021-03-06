package polygen

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/gif" // register image formats
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"os"
)

func MustReadImage(file string) image.Image {
	infile, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer infile.Close()

	img, _, err := image.Decode(infile)
	if err != nil {
		log.Fatal(err)
	}

	return img
}

// Compare compares images by computing the square root of the total sum of individual squared pixel differences.
func Compare(img1, img2 image.Image) (int64, error) {
	if img1.Bounds() != img2.Bounds() {
		return 0, fmt.Errorf("image bounds not equal: %+v, %+v", img1.Bounds(), img2.Bounds())
	}

	accumError := int64(0)
	bounds := img1.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c1 := img1.At(x, y)
			c2 := img2.At(x, y)

			r1, g1, b1, a1 := c1.RGBA()
			r2, g2, b2, a2 := c2.RGBA()

			// TODO: consider ignoring the Alpha, since the colors are pre-multiplied
			sum := sqDiff(r1, r2) + sqDiff(g1, g2) + sqDiff(b1, b2) + sqDiff(a1, a2)
			accumError += int64(sum)
		}
	}

	return int64(math.Sqrt(float64(accumError))), nil
}

// FastCompare compares images by diffing the underlying byte arrays directly.
// This is more than 10x faster than Compare(), but requires a concrete instance of image.RGBA.
func FastCompare(img1, img2 *image.RGBA) (uint64, error) {
	if img1.Bounds() != img2.Bounds() {
		return 0, fmt.Errorf("image bounds not equal: %+v, %+v", img1.Bounds(), img2.Bounds())
	}

	accumError := uint64(0)

	for i := 0; i < len(img1.Pix); i++ {
		accumError += uint64(diffUint8(img1.Pix[i], img2.Pix[i]))
	}

	//return int64(math.Sqrt(float64(accumError))), nil
	return accumError, nil
}

// from http://blog.golang.org/go-imagedraw-package ("Converting an Image to RGBA"),
// modified slightly to be a no-op if the src image is already RGBA
//
func ConvertToRGBA(img image.Image) (result *image.RGBA) {
	result, ok := img.(*image.RGBA)
	if ok {
		return result
	}

	b := img.Bounds()
	result = image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(result, result.Bounds(), img, b.Min, draw.Src)

	return
}

// taken directly from image/color/color.go:
//
// sqDiff returns the squared-difference of x and y, shifted by 2 so that
// adding four of those won't overflow a uint32.
//
// x and y are both assumed to be in the range [0, 0xffff].
func sqDiff(x, y uint32) uint32 {
	var d uint32
	if x > y {
		d = x - y
	} else {
		d = y - x
	}
	return (d * d) >> 2
}

func diffUint8(x, y uint8) uint8 {
	if x > y {
		return x - y
	} else {
		return y - x
	}
}
