package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"

	"golang.org/x/image/vector"
)

// makeICO wraps raw PNG bytes in a minimal single-image ICO file container.
// Windows LoadImageW (used by getlantern/systray) only accepts ICO format;
// passing raw PNG causes "Unable to set icon: The operation completed successfully".
// PNG-in-ICO is supported on Windows Vista and later.
func makeICO(pngData []byte, w, h int) []byte {
	const (
		icoDirSize  = 6
		icoDirEntry = 16
		dataOffset  = icoDirSize + icoDirEntry
	)

	out := make([]byte, dataOffset+len(pngData))

	// ICONDIR header (6 bytes)
	binary.LittleEndian.PutUint16(out[0:], 0) // idReserved = 0
	binary.LittleEndian.PutUint16(out[2:], 1) // idType = 1 (icon)
	binary.LittleEndian.PutUint16(out[4:], 1) // idCount = 1

	// ICONDIRENTRY (16 bytes, starting at offset 6)
	bw := w
	if bw >= 256 {
		bw = 0 // 0 encodes as 256 in the ICO spec
	}
	bh := h
	if bh >= 256 {
		bh = 0
	}
	out[6] = uint8(bw) // bWidth
	out[7] = uint8(bh) // bHeight
	out[8] = 0         // bColorCount (0 = >8bpp)
	out[9] = 0         // bReserved
	binary.LittleEndian.PutUint16(out[10:], 1)                    // wPlanes
	binary.LittleEndian.PutUint16(out[12:], 32)                   // wBitCount
	binary.LittleEndian.PutUint32(out[14:], uint32(len(pngData))) // dwBytesInRes
	binary.LittleEndian.PutUint32(out[18:], dataOffset)           // dwImageOffset

	copy(out[dataOffset:], pngData)
	return out
}

// drawBadge renders a 32×32 badge icon styled after the favicon-badge.svg:
// a rounded rectangle background with two wing-shaped paths in the foreground.
func drawBadge(bg, fg color.RGBA) []byte {
	const (
		size = 32
		rx   = 7.0 // corner radius proportional to svg rx=14 on 64px → 7px on 32px
	)

	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Fill rounded rectangle background pixel-by-pixel.
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			px, py := float64(x)+0.5, float64(y)+0.5
			inCornerX := px < rx || px > float64(size)-rx
			inCornerY := py < rx || py > float64(size)-rx
			if inCornerX && inCornerY {
				cx := rx
				if px > float64(size)-rx {
					cx = float64(size) - rx
				}
				cy := rx
				if py > float64(size)-rx {
					cy = float64(size) - rx
				}
				dx, dy := px-cx, py-cy
				if dx*dx+dy*dy > rx*rx {
					continue
				}
			}
			img.SetRGBA(x, y, bg)
		}
	}

	// tf maps svg path coordinates (from the 64×64 viewBox with group transform
	// translate(32,32) scale(0.62) translate(-32.5,-36)) to 32×32 icon pixels.
	tf := func(x, y float32) (float32, float32) {
		return 16 + 0.31*(x-32.5), 16 + 0.31*(y-36)
	}

	ras := vector.NewRasterizer(size, size)
	src := image.NewUniform(fg)
	bounds := image.Rect(0, 0, size, size)

	// cubeTo is a helper that applies tf to all three control/endpoint pairs
	// before forwarding to ras.CubeTo (Go forbids passing multi-return values inline).
	cubeTo := func(ras *vector.Rasterizer, x1, y1, x2, y2, x, y float32) {
		ax, ay := tf(x1, y1)
		bx, by := tf(x2, y2)
		cx, cy := tf(x, y)
		ras.CubeTo(ax, ay, bx, by, cx, cy)
	}

	// Path 1 – left/lower wing shape.
	{
		px, py := tf(28.75, 17.25)
		ras.MoveTo(px, py)
		cubeTo(ras, 28.92, 18.62, 29.04, 22.67, 29.75, 25.50)
		cubeTo(ras, 30.46, 28.33, 31.79, 31.71, 33.00, 34.25)
		cubeTo(ras, 34.21, 36.79, 35.46, 39.12, 37.00, 40.75)
		cubeTo(ras, 38.54, 42.38, 42.12, 42.38, 42.25, 44.00)
		cubeTo(ras, 42.38, 45.62, 39.58, 48.58, 37.75, 50.50)
		cubeTo(ras, 35.92, 52.42, 32.96, 54.79, 31.25, 55.50)
		cubeTo(ras, 29.54, 56.21, 29.04, 55.88, 27.50, 54.75)
		cubeTo(ras, 25.96, 53.62, 23.75, 51.38, 22.00, 48.75)
		cubeTo(ras, 20.25, 46.12, 18.33, 42.58, 17.00, 39.00)
		cubeTo(ras, 15.67, 35.42, 14.46, 29.75, 14.00, 27.25)
		cubeTo(ras, 13.54, 24.75, 13.71, 25.04, 14.25, 24.00)
		cubeTo(ras, 14.79, 22.96, 14.83, 22.12, 17.25, 21.00)
		cubeTo(ras, 19.67, 19.88, 26.67, 16.50, 28.75, 17.25)
		ras.ClosePath()
		_, _ = px, py
	}
	ras.Draw(img, bounds, src, image.Point{})
	ras.Reset(size, size)

	// Path 2 – upper/right wing shape.
	{
		px, py := tf(39.75, 16.50)
		ras.MoveTo(px, py)
		cubeTo(ras, 41.04, 16.62, 45.50, 16.62, 47.50, 17.25)
		cubeTo(ras, 49.50, 17.88, 51.25, 18.29, 51.75, 20.25)
		cubeTo(ras, 52.25, 22.21, 51.50, 26.25, 50.50, 29.00)
		cubeTo(ras, 49.50, 31.75, 47.33, 35.92, 45.75, 36.75)
		cubeTo(ras, 44.17, 37.58, 42.46, 35.92, 41.00, 34.00)
		cubeTo(ras, 39.54, 32.08, 37.79, 27.88, 37.00, 25.25)
		cubeTo(ras, 36.21, 22.62, 35.79, 19.71, 36.25, 18.25)
		cubeTo(ras, 36.71, 16.79, 37.88, 16.67, 39.75, 16.50)
		ras.ClosePath()
		_, _ = px, py
	}
	ras.Draw(img, bounds, src, image.Point{})

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return makeICO(buf.Bytes(), size, size)
}

// makeIcon is retained for call-sites that pass explicit colors.
// The inner color is used as background; outer color is ignored in the badge design.
func makeIcon(inner, outer color.NRGBA) []byte {
	_ = outer
	return drawBadge(
		color.RGBA{inner.R, inner.G, inner.B, inner.A},
		color.RGBA{255, 255, 255, 255},
	)
}

// getIdleIcon returns the indigo badge icon used when the agent is idle.
func getIdleIcon() []byte {
	return drawBadge(
		color.RGBA{0x4E, 0x46, 0xDD, 255}, // #4E46DD – matches favicon-badge.svg
		color.RGBA{255, 255, 255, 255},
	)
}

// getScanIcon returns an amber badge icon used while scanning is in progress.
func getScanIcon() []byte {
	return drawBadge(
		color.RGBA{0xE8, 0x7C, 0x00, 255}, // amber
		color.RGBA{255, 255, 255, 255},
	)
}

// getIcon is kept for backward compatibility; returns the idle icon.
func getIcon() []byte {
	return getIdleIcon()
}
