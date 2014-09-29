// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

// State represents a generic set of graphics state properties to be used when
// rendering a graphics object. Changes to such properties across multiple draw
// calls (called 'graphics state changes' or 'render state changes') have a
// performance cost.
//
// The performance penalty mentioned depends on several factors (graphics
// hardware, drivers, the specific property being changed, etc). The important
// factor to recognize is that multiple draw calls are faster when the objects
// being draw would cause less changes to the graphics state than the
// previously drawn object.
type State struct {
	// A single alpha transparency mode describing how transparent parts of
	// of the object are to be rendered.
	// Must be one of: NoAlpha, AlphaBlend, BinaryAlpha, AlphaToCoverage
	AlphaMode AlphaMode

	// Blend represents how blending between existing (source) and new
	// (destination) pixels in the color buffer occurs when AlphaMode ==
	// AlphaBlend.
	Blend BlendState

	// Whether or not red/green/blue/alpha should be written to the color
	// buffer or not when rendering this object.
	WriteRed, WriteGreen, WriteBlue, WriteAlpha bool

	// Whether or not dithering should be used when rendering the object.
	Dithering bool

	// Whether or not depth testing and depth writing should be enabled when
	// rendering the object.
	DepthTest, DepthWrite bool

	// The comparison operator to use for depth testing against existing pixels
	// in the depth buffer.
	DepthCmp Cmp

	// Whether or not stencil testing should be enabled when rendering the
	// object.
	StencilTest bool

	// Whether or not (and how) face culling should occur when rendering
	// the object.
	// Must be one of: BackFaceCulling, FrontFaceCulling, NoFaceCulling
	FaceCulling FaceCullMode

	// The stencil state for front and back facing pixels, respectively.
	StencilFront, StencilBack StencilState
}

// Compare compares this state against the other one using DefaultState as a
// reference when inequality occurs and returns whether or not this state
// should sort before the other one for purposes of state sorting.
func (s State) Compare(other State) bool {
	if s == other {
		return true
	}
	if s.AlphaMode != other.AlphaMode {
		return s.AlphaMode == DefaultState.AlphaMode
	}
	if s.Blend != other.Blend {
		return s.Blend.Compare(other.Blend)
	}
	if s.DepthTest != other.DepthTest {
		return s.DepthTest == DefaultState.DepthTest
	}
	if s.StencilTest != other.StencilTest {
		return s.StencilTest == DefaultState.StencilTest
	}
	if s.StencilFront != other.StencilFront {
		return s.StencilFront.Compare(other.StencilFront)
	}
	if s.StencilBack != other.StencilBack {
		return s.StencilBack.Compare(other.StencilBack)
	}
	if s.DepthWrite != other.DepthWrite {
		return s.DepthWrite == DefaultState.DepthWrite
	}
	if s.DepthCmp != other.DepthCmp {
		return s.DepthCmp == DefaultState.DepthCmp
	}
	if s.FaceCulling != other.FaceCulling {
		return s.FaceCulling == DefaultState.FaceCulling
	}
	if s.WriteRed != other.WriteRed {
		return s.WriteRed == DefaultState.WriteRed
	}
	if s.WriteGreen != other.WriteGreen {
		return s.WriteGreen == DefaultState.WriteGreen
	}
	if s.WriteBlue != other.WriteBlue {
		return s.WriteBlue == DefaultState.WriteBlue
	}
	if s.WriteAlpha != other.WriteAlpha {
		return s.WriteAlpha == DefaultState.WriteAlpha
	}
	if s.Dithering != other.Dithering {
		return s.Dithering == DefaultState.Dithering
	}
	return true
}

// The default state that should be used for graphics objects.
var DefaultState = State{
	AlphaMode:    NoAlpha,
	Blend:        DefaultBlendState,
	WriteRed:     true,
	WriteGreen:   true,
	WriteBlue:    true,
	WriteAlpha:   true,
	Dithering:    true,
	DepthTest:    true,
	DepthWrite:   true,
	DepthCmp:     Less,
	StencilTest:  false,
	FaceCulling:  BackFaceCulling,
	StencilFront: DefaultStencilState,
	StencilBack:  DefaultStencilState,
}
