package main

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
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

// makeIcon draws a 32×32 hexagon-patterned circle using the given fill colors
// and returns the result wrapped in an ICO container for Windows compatibility.
func makeIcon(inner, outer color.NRGBA) []byte {
	const size = 32 // 32×32 is standard for Windows tray icons
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
	return makeICO(buf.Bytes(), size, size)
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
