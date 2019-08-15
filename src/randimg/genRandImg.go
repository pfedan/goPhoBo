package randimg

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"time"
)

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
			uint8((1 - math.Pow(d, 5)) * 255 * c.red),
			uint8((1 - math.Pow(d, 5)) * 255 * c.green),
			uint8((1 - math.Pow(d, 5)) * 255 * c.blue),
			255,
		}
	}
}

func GetImg(w, h int) *image.RGBA {
	rand.Seed(time.Now().UTC().UnixNano())
	circs := make([]*Circle, 20)
	for i := range circs {
		circs[i] = &Circle{
			X:     rand.Float64() * float64(w),
			Y:     rand.Float64() * float64(h),
			R:     rand.Float64()*200 + 10,
			red:   rand.Float64(),
			green: rand.Float64(),
			blue:  rand.Float64(),
		}

	}

	m := image.NewRGBA(image.Rect(0, 0, w, h))
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {

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
