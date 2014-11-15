// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import "fmt"

// TexWrap represents a single way that the extra area of a texture wraps
// around a mesh.
type TexWrap uint8

// String returns a string representation of this TexWrapMode.
// e.g. Repeat -> "Repeat"
func (t TexWrap) String() string {
	switch t {
	case Repeat:
		return "Repeat"
	case Clamp:
		return "Clamp"
	case BorderColor:
		return "BorderColor"
	case Mirror:
		return "Mirror"
	}
	return fmt.Sprintf("TexWrap(%d)", t)
}

const (
	// Repeat repeats the extra area of the texture into infinity.
	Repeat TexWrap = iota

	// Clamp clamps the extra area of the texture by stretching the edge pixels
	// out into infinity.
	Clamp

	// BorderColor represents the extra area of the texture by the border color
	// specified on the Texture object.
	BorderColor

	// Mirror represents the extra area of the texture by mirroring itself into
	// infinity.
	Mirror
)
