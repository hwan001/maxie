package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
)

// makeIcon draws a 22×22 hexagon-patterned circle using the given fill colors.
func makeIcon(inner, outer color.NRGBA) []byte {
	const size = 22
	img := image.NewNRGBA(image.Rect(0, 0, size, size))

	cx, cy := float64(size)/2, float64(size)/2
	r := float64(size)/2 - 1

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			dist := math.Sqrt(dx*dx + dy*dy)

			if dist <= r {
				// Inner hexagon pattern using Manhattan distance proxy
				hex := math.Abs(dx)*0.866 + math.Abs(dy)*0.5
				if hex <= r*0.7 {
					img.Set(x, y, inner)
				} else {
					img.Set(x, y, outer)
				}
			}
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

// getIdleIcon returns a bright cyan/teal icon used when the agent is idle.
func getIdleIcon() []byte {
	return makeIcon(
		color.NRGBA{R: 0, G: 210, B: 210, A: 255},
		color.NRGBA{R: 0, G: 140, B: 160, A: 255},
	)
}

// getScanIcon returns a bright orange icon used while scanning is in progress.
func getScanIcon() []byte {
	return makeIcon(
		color.NRGBA{R: 255, G: 160, B: 0, A: 255},
		color.NRGBA{R: 200, G: 100, B: 0, A: 255},
	)
}

// getIcon is kept for backward compatibility; returns the idle icon.
func getIcon() []byte {
	return getIdleIcon()
}
