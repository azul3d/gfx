// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import (
	"image/color"
	"math"
)

var fMaxUint16 = float32(math.MaxUint16)

// Color represents a normalized (values are in the range of 0.0 to 1.0)
// 32-bit floating point RGBA color data type.
type Color struct {
	R, G, B, A float32
}

// RGBA implements the color.Color interface.
func (c Color) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R * fMaxUint16)
	g = uint32(c.G * fMaxUint16)
	b = uint32(c.B * fMaxUint16)
	a = uint32(c.A * fMaxUint16)
	return
}

func colorModel(c color.Color) color.Color {
	if _, ok := c.(Color); ok {
		return c
	}
	r, g, b, a := c.RGBA()
	return Color{
		R: float32(r) / float32(fMaxUint16),
		G: float32(g) / float32(fMaxUint16),
		B: float32(b) / float32(fMaxUint16),
		A: float32(a) / float32(fMaxUint16),
	}
}

// ColorModel represents the graphics color model (i.e. normalized 32-bit
// floating point values RGBA color).
var ColorModel = color.ModelFunc(colorModel)
