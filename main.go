package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/llgcode/draw2d/draw2dimg"
)

func hexColor(c color.Color) string {
	rgba := color.RGBAModel.Convert(c).(color.RGBA)
	return fmt.Sprintf("0x%02X%02X%02X", rgba.R, rgba.G, rgba.B)
}

func intColor(c color.RGBA) int {
	r := (int(c.R) << 16) & 0xff0000
	g := (int(c.G) << 8) & 0x00ff00
	b := int(c.B) & 0x0000ff
	return r | g | b
	//return int(c.R + c.G + c.B) / 3
}

const scale = 8

type Bucket struct {
	Color  color.RGBA
	Pixels []Pixel
}

type Pixel struct {
	x, y int
}

var inFile, outFile string
var normalSort bool

func init() {
	flag.StringVar(&inFile, "in", "", "GIF, PNG or JPEG to read from")
	flag.StringVar(&outFile, "out", "", "PNG to write to")
	flag.BoolVar(&normalSort, "normal", false, "Normal colour sort (might give better results)")
}

func loadImage(file *os.File) (image.Image, error) {
	var img image.Image
	var err error
	filename := strings.ToLower(file.Name())
	switch {
	case strings.HasSuffix(filename, "png"):
		img, err = png.Decode(file)
	case strings.HasSuffix(filename, "gif"):
		img, err = gif.Decode(file)
	case strings.HasSuffix(filename, "jpg"):
		fallthrough
	case strings.HasSuffix(filename, "jpeg"):
		img, err = jpeg.Decode(file)
	default:
		err = fmt.Errorf("can not determine image format: %q", filename)
	}
	return img, err
}

func main() {
	flag.Parse()

	if inFile == "" || outFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	if _, err := os.Stat(inFile); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	start := time.Now()
	dt := time.Now()
	in, err := os.Open(inFile)
	if err != nil {
		log.Fatalf("unable to open %q", inFile)
	}
	defer in.Close()

	fmt.Println("Load", time.Since(dt))

	dt = time.Now()

	img, err := loadImage(in)
	if err != nil {
		log.Fatalf("unable to decode %q: %s", inFile, err)
	}
	fmt.Println("Decode", time.Since(dt))

	dt = time.Now()

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	var cols []int
	buckets := make(map[int]Bucket)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
			h := intColor(c)
			if _, ok := buckets[h]; !ok {
				cols = append(cols, h)
				buckets[h] = Bucket{
					Color:  c,
					Pixels: []Pixel{},
				}
			}
			b := buckets[h]
			b.Pixels = append(b.Pixels, Pixel{x, y})
			buckets[h] = b
		}
	}
	fmt.Println("Bucketing", time.Since(dt))

	dt = time.Now()
	if normalSort {
		sort.Sort(sort.IntSlice(cols))
		fmt.Println("Sorting", time.Since(dt))
	} else {
		sort.Sort(sort.Reverse(sort.IntSlice(cols)))
		fmt.Println("Reverse sorting", time.Since(dt))
	}

	dt = time.Now()
	// Initialize the graphic context on an RGBA image
	dest := image.NewRGBA(image.Rect(0, 0, w*scale, h*scale))
	gc := draw2dimg.NewGraphicContext(dest)

	for _, h := range cols {
		b := buckets[h]

		//gc.SetFillColor(color.RGBA{0x44, 0xff, 0x44, 0xff})
		gc.BeginPath()
		gc.SetStrokeColor(b.Color)
		l := len(b.Pixels)
		if l == 1 {
			p := b.Pixels[0]
			gc.SetLineWidth(scale)
			gc.MoveTo(float64(p.x-scale/2)*scale, float64(p.y)*scale)
			gc.LineTo(float64(p.x)*scale, float64(p.y)*scale)
		} else {
			for i, p := range b.Pixels {
				//gc.SetLineWidth(float64(l - i))
				gc.SetLineWidth(scale)
				if i == 0 {
					gc.MoveTo(float64(p.x-scale/2)*scale, float64(p.y)*scale)
				} else {
					gc.LineTo(float64(p.x)*scale, float64(p.y)*scale)
				}
			}
		}
		gc.Stroke()
		gc.Close()
	}
	fmt.Println("Drawing", time.Since(dt))

	dt = time.Now()
	// Save to file
	err = draw2dimg.SaveToPngFile(outFile, dest)
	if err != nil {
		log.Fatalf("unable to save %q", outFile)
	}
	fmt.Println("Saving", time.Since(dt))
	fmt.Println("Overall", time.Since(start))
	fmt.Println("Done painting " + outFile)
}
