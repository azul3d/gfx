// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import "fmt"

// TexFilter represents a single texture filter to be used for minification or
// magnification of a texture during rendering.
type TexFilter uint8

// String returns a string representation of this TexFilter.
// e.g. Nearest -> "Nearest"
func (t TexFilter) String() string {
	switch t {
	case Nearest:
		return "Nearest"
	case Linear:
		return "Linear"
	}
	return fmt.Sprintf("TexFilter(%d)", t)
}

// Mipmapped tells if the texture filter is a mipmapped one, that is one of:
//  NearestMipmapNearest
//  LinearMipmapNearest
//  NearestMipmapLinear
//  LinearMipmapLinear
func (t TexFilter) Mipmapped() bool {
	switch t {
	case NearestMipmapNearest:
		return true
	case LinearMipmapNearest:
		return true
	case NearestMipmapLinear:
		return true
	case LinearMipmapLinear:
		return true
	}
	return false
}

const (
	// Nearest samples the nearest pixel.
	Nearest TexFilter = iota

	// Linear samples the four closest pixels and linearly interpolates them.
	Linear

	// NearestMipmapNearest samples the pixel from the nearest mipmap, it may
	// not be used as a magnification filter.
	NearestMipmapNearest

	// LinearMipmapNearest (AKA Bilinear filtering) samples the pixel from the
	// nearest mipmap.
	//
	// It may not be used as a magnification filter.
	LinearMipmapNearest

	// NearestMipmapLinear samples the pixel from the two closest mipmaps, and
	// linearly blends the result.
	//
	// It may not be used as a magnification filter.
	NearestMipmapLinear

	// LinearMipmapLinear (AKA Trilinear filtering) bilinearly filters the
	// pixel from the two mipmaps, and linear blends the result.
	//
	// It may not be used as a magnification filter.
	LinearMipmapLinear
)
