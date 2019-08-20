package randimg

import (
	"image"
	"image/color"
	"reflect"
	"testing"
)

func TestCircle_getColorAt(t *testing.T) {
	type args struct {
		x float64
		y float64
	}
	tests := []struct {
		name string
		c    *Circle
		args args
		want color.RGBA
	}{
		{name: "Center",
			c:    &Circle{R: 100, X: 100, Y: 100, red: 1.0, green: 1.0, blue: 1.0},
			args: args{x: 100, y: 100},
			want: color.RGBA{R: 200, G: 200, B: 200, A: 255},
		},
		{name: "Outside",
			c:    &Circle{R: 100, X: 100, Y: 100, red: 1.0, green: 1.0, blue: 1.0},
			args: args{x: 0, y: 0},
			want: color.RGBA{R: 0, G: 0, B: 0, A: 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.getColorAt(tt.args.x, tt.args.y); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Circle.getColorAt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetImg(t *testing.T) {
	type args struct {
		o RandImgOptions
	}
	tests := []struct {
		name string
		args args
		want *image.RGBA
	}{
		{
			name: "noSeed",
			args: args{o: RandImgOptions{H: 50, W: 50, Seed: 0}},
			want: image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{50, 50}}), 
		},
		{
			name: "withSeed",
			args: args{o: RandImgOptions{H: 800, W: 600, Seed: 1}},
			want: image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{800, 600}}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetImg(tt.args.o); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetImg() = %T, want %T", got, tt.want)
			}
		})
	}
}
