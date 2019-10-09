package randimg

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"
)

// RandImgOptions specifies width, height and an optional seed for the creation of a random image
type RandImgOptions struct {
	W, H int
	Seed int64
}

// Circle defines a circle by x and y coordinates and a radius.
type Circle struct {
	X, Y, R          float64
	red, green, blue float64
}

func (c *Circle) getColorAt(x, y float64) color.RGBA {
	var dx, dy float64 = c.X - x, c.Y - y
	d := math.Sqrt(dx*dx+dy*dy) / c.R
	if d > 1 {
		// outside
		return color.RGBA{uint8(0), uint8(0), uint8(0), uint8(0)}
	} else {
		// inside
		return color.RGBA{
			uint8((1 - math.Pow(d, 5)) * 200 * c.red),
			uint8((1 - math.Pow(d, 5)) * 200 * c.green),
			uint8((1 - math.Pow(d, 5)) * 200 * c.blue),
			255,
		}
	}
}

// GetImg creates a random image of the size specified in o (RandImgOptions). If a seed is provided, the result is deterministic.
func GetImg(o RandImgOptions) *image.RGBA {
	circs := make([]*Circle, 20)
	m := image.NewRGBA(image.Rect(0, 0, o.W, o.H))

	if o.Seed == 0 {
		mTime := time.Now()
		mSeed := mTime.UnixNano()
		rand.Seed(mSeed)
	} else {
		rand.Seed(o.Seed)
	}

	for i := range circs {
		circs[i] = &Circle{
			X:     rand.Float64() * float64(o.W),
			Y:     rand.Float64() * float64(o.H),
			R:     rand.Float64()*200 + 10,
			red:   rand.Float64(),
			green: rand.Float64(),
			blue:  rand.Float64(),
		}
	}

	for x := 0; x < o.W; x++ {
		for y := 0; y < o.H; y++ {
			var iCol, cCol color.RGBA
			for _, v := range circs {
				cCol = v.getColorAt(float64(x), float64(y))
				if int32(iCol.R)+int32(cCol.R) <= 255 {
					iCol.R += cCol.R
				} else {
					iCol.R = 255
				}
				if int32(iCol.G)+int32(cCol.G) <= 255 {
					iCol.G += cCol.G
				} else {
					iCol.G = 255
				}
				if int32(iCol.B)+int32(cCol.B) <= 255 {
					iCol.B += cCol.B
				} else {
					iCol.B = 255
				}
			}
			m.Set(x, y, iCol)
		}
	}

	return m
}
