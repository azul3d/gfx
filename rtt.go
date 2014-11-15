// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import (
	"fmt"
	"image"
	"sort"
)

// DSFormat specifies a single depth or stencil buffer storage format.
type DSFormat uint8

// String returns a string name for this depth/stencil buffer format. For
// example:
//  Depth24AndStencil8 -> "Depth24AndStencil8"
func (f DSFormat) String() string {
	switch f {
	case ZeroDSFormat:
		return "ZeroDSFormat"
	case Depth16:
		return "Depth16"
	case Depth24:
		return "Depth24"
	case Depth32:
		return "Depth32"
	case Depth24AndStencil8:
		return "Depth24AndStencil8"
	}
	return fmt.Sprintf("DSFormat(%d)", f)
}

// IsDepth tells if f is a valid depth buffer format. It must be one of the
// following:
//  Depth16
//  Depth24
//  Depth32
//  Depth24AndStencil8
func (f DSFormat) IsDepth() bool {
	switch f {
	case Depth16:
		return true
	case Depth24:
		return true
	case Depth32:
		return true
	case Depth24AndStencil8:
		return true
	}
	return false
}

// IsStencil tells if f is a valid stencil buffer format. It must be one of the
// following:
//  Depth24AndStencil8
func (f DSFormat) IsStencil() bool {
	switch f {
	case Depth24AndStencil8:
		return true
	}
	return false
}

// IsCombined tells if f is a combined depth and stencil buffer format. It must
// be one of the following:
//  Depth24AndStencil8
func (f DSFormat) IsCombined() bool {
	switch f {
	case Depth24AndStencil8:
		return true
	}
	return false
}

// DepthBits returns the number of depth bits in this format. For example:
//  Depth16 == 16
//  Depth24AndStencil8 == 24
func (f DSFormat) DepthBits() uint8 {
	switch f {
	case Depth16:
		return 16
	case Depth24:
		return 24
	case Depth32:
		return 32
	case Depth24AndStencil8:
		return 24
	}
	return 0
}

// StencilBits returns the number of stencil bits in this format. For example:
//  Depth24AndStencil8 == 8
//  Depth16 == 0
func (f DSFormat) StencilBits() uint8 {
	switch f {
	case Depth24AndStencil8:
		return 8
	}
	return 0
}

const (
	// ZeroDSFormat is a zero-value depth/stencil format.
	ZeroDSFormat DSFormat = iota

	// Depth16 is the 16-bit depth-buffer-only format.
	Depth16

	// Depth24 is the 24-bit depth-buffer-only format.
	Depth24

	// Depth32 is the 32-bit depth-buffer-only format.
	Depth32

	// Depth24AndStencil8 is the combined 24-bit depth-buffer and 8-bit
	// stencil-buffer format.
	Depth24AndStencil8
)

// RTTConfig represents a configuration used for creating a render-to-texture
// (RTT) canvas. At least one of Color, Depth, or Stencil textures must be non
// nil.
type RTTConfig struct {
	// Bounds is the target resolution of the canvas to render at. If it is an
	// empty rectangle, the renderer's bounds is used instead.
	Bounds image.Rectangle

	// The number of samples to use for multisampling. It should be one of the
	// numbers listed in the GPUInfo.RTTFormats structure.
	Samples int

	// Color, Depth, and Stencil textures, each of these texture's Format
	// fields are explicitly ignored (see the format fields below). If any of
	// these textures are non-nil, the results of that buffer (e.g. the color
	// buffer) are stored into that texture.
	//
	// Specify nil for any you do not intend to use as a texture (e.g. if you
	// want a 16-bit depth buffer but do not intend to use it as a texture, you
	// could set Depth == nil and DepthFormat == Depth16).
	Color, Depth, Stencil *Texture

	// Color format to use for the color buffer, it should be one listed in the
	// GPUInfo.RTTFormats structure.
	ColorFormat TexFormat

	// Depth and Stencil formats to use for the depth and stencil buffers,
	// respectively. They should be ones listed in the GPUInfo.RTTFormats
	// structure.
	//
	// Combined depth and stencil formats can be used (e.g. by setting
	// both DepthFormat and StencilFormat to Depth24AndStencil8), they are
	// often faster and use less memory, but with the caveat that they cannot
	// be used as textures.
	DepthFormat, StencilFormat DSFormat
}

// Valid tells if this render-to-texture (RTT) configuration is valid or not, a
// configuration is considered invalid if:
//  1. All three textures are nil.
//  2. Any non-nil texture is not accompanies by a format.
//  3. Either DepthFormat.IsCombined() or StencilFormat.IsCombined() and the other
//     is not.
func (c RTTConfig) Valid() bool {
	if c.Color == nil && c.Depth == nil && c.Stencil == nil {
		return false
	}
	if c.Color != nil && c.ColorFormat == ZeroTexFormat {
		return false
	}
	if c.Depth != nil && c.DepthFormat == ZeroDSFormat {
		return false
	}
	if c.Stencil != nil && c.StencilFormat == ZeroDSFormat {
		return false
	}

	if c.DepthFormat.IsCombined() != c.StencilFormat.IsCombined() {
		return false
	}
	return true
}

// RTTFormats represents color, depth, and stencil buffer formats applicable to
// render-to-texture (RTT).
type RTTFormats struct {
	// Slice of sample counts available for multisampling.
	Samples []int

	// Slice of color buffer formats.
	ColorFormats []TexFormat

	// Slices of depth and stencil buffer formats.
	DepthFormats, StencilFormats []DSFormat
}

func absInt(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

type chooseTexFormats struct {
	s []TexFormat
	p Precision
}

func (s chooseTexFormats) bpp(t TexFormat) int {
	r, g, b, a := t.Bits()
	return int(r) + int(g) + int(b) + int(a)
}
func (s chooseTexFormats) Len() int      { return len(s.s) }
func (s chooseTexFormats) Swap(i, j int) { s.s[i], s.s[j] = s.s[j], s.s[i] }
func (s chooseTexFormats) Less(ii, jj int) bool {
	i := s.bpp(s.s[ii])
	j := s.bpp(s.s[jj])
	pb := int(s.p.RedBits) + int(s.p.GreenBits) + int(s.p.BlueBits) + int(s.p.AlphaBits)

	// Choose the closest.
	iDist := absInt(pb - i)
	jDist := absInt(pb - j)
	return iDist < jDist
}

type chooseDSFormats struct {
	s                     []DSFormat
	p                     Precision
	preferCombined, depth bool
}

func (s chooseDSFormats) Len() int      { return len(s.s) }
func (s chooseDSFormats) Swap(i, j int) { s.s[i], s.s[j] = s.s[j], s.s[i] }
func (s chooseDSFormats) Less(ii, jj int) bool {
	var i, j, pb int
	if s.depth {
		i = int(s.s[ii].DepthBits())
		j = int(s.s[jj].DepthBits())
		pb = int(s.p.DepthBits)
	} else {
		i = int(s.s[ii].StencilBits())
		j = int(s.s[jj].StencilBits())
		pb = int(s.p.StencilBits)
	}

	// Choose the closest.
	iDist := absInt(pb - i)
	jDist := absInt(pb - j)
	if iDist == jDist {
		// If they are equal, choose based on whether or not we prefer combined
		// formats or not.
		iCombined := s.s[ii].IsCombined()
		if s.preferCombined {
			return iCombined
		}
		return !iCombined
	}
	return iDist < jDist
}

// Choose returns a color, depth, and stencil format from the formats, f. It
// tries to choose the most applicable one to the requested precision. If
// compression is true, it tries to choose the most compressed format.
func (f RTTFormats) Choose(p Precision, compression bool) (color TexFormat, depth, stencil DSFormat) {
	wantColor := (p.RedBits + p.GreenBits + p.BlueBits + p.AlphaBits) > 0
	if wantColor && len(f.ColorFormats) > 0 {
		colors := chooseTexFormats{
			s: f.ColorFormats,
			p: p,
		}
		sort.Sort(colors)
		color = colors.s[0]
	}
	if (p.DepthBits > 0) && len(f.DepthFormats) > 0 {
		depths := chooseDSFormats{
			s:              f.DepthFormats,
			p:              p,
			depth:          true,
			preferCombined: false,
		}
		sort.Sort(depths)
		depth = depths.s[0]
		for _, d := range depths.s {
			fmt.Println(d)
		}
	}
	if (p.StencilBits > 0) && len(f.StencilFormats) > 0 {
		stencils := chooseDSFormats{
			s:              f.StencilFormats,
			p:              p,
			depth:          false,
			preferCombined: false,
		}
		sort.Sort(stencils)
		stencil = stencils.s[0]
	}

	// If we found a combined format, depth+stencil must be the same.
	if depth.IsCombined() {
		stencil = depth
	}
	if stencil.IsCombined() {
		depth = stencil
	}
	return
}

// ChooseConfig is short-hand for:
//  colorFormat, depthFormat, stencilFormat := f.Choose(p, compression)
//  cfg := RTTConfig{
//      ColorFormat: colorFormat,
//      DepthFormat: depthFormat,
//      StencilFormat: stencilFormat,
//  }
func (f RTTFormats) ChooseConfig(p Precision, compression bool) RTTConfig {
	color, depth, stencil := f.Choose(p, compression)
	return RTTConfig{
		ColorFormat:   color,
		DepthFormat:   depth,
		StencilFormat: stencil,
	}
}
