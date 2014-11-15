// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import (
	"fmt"
	"image"
	"sync"
)

// TexFormat specifies a single texture storage format.
type TexFormat uint8

// String returns a string name for this texture format. For example:
//  RGBA -> "RGBA"
func (t TexFormat) String() string {
	switch t {
	case ZeroTexFormat:
		return "ZeroTexFormat"
	case RGB:
		return "RGB"
	case RGBA:
		return "RGBA"
	case DXT1:
		return "DXT1"
	case DXT1RGBA:
		return "DXT1RGBA"
	case DXT3:
		return "DXT3"
	case DXT5:
		return "DXT5"
	}
	return fmt.Sprintf("TexFormat(%d)", t)
}

// Bits returns the number of bits per color component in this texture format.
// For example:
//  r, g, b, a := RGB.Bits()
//  r == 8 && g == 8 && b == 8 && a == 0
// A panic will occur if the format is not one of the predefined ones in this
// package.
//
// ZeroTexFormat, DXT1, DXT3, and DXT5 formats will return only zero.
func (t TexFormat) Bits() (r, g, b, a uint8) {
	switch t {
	case RGB:
		return 8, 8, 8, 0
	case RGBA:
		return 8, 8, 8, 8

	case ZeroTexFormat:
		return 0, 0, 0, 0
	case DXT1:
		return 0, 0, 0, 0
	case DXT1RGBA:
		return 0, 0, 0, 0
	case DXT3:
		return 0, 0, 0, 0
	case DXT5:
		return 0, 0, 0, 0
	}
	panic("invalid format")
}

const (
	// ZeroTexFormat is the zero-value texture format, it is not a valid format
	// and is mainly used to catch zero-value errors.
	ZeroTexFormat TexFormat = iota

	// RGBA is a standard 32-bit premultiplied alpha image format.
	RGBA

	// RGB is a standard 24-bit RGB image format with no alpha component.
	RGB

	// DXT1 is a DXT1 texture compression format in RGB form (i.e. fully
	// opaque) each 4x4 block of pixels take up 64-bits of data, as such when
	// compared to a standard 24-bit RGB format it provides a 6:1 compression
	// ratio.
	DXT1

	// DXT1RGBA is a DXT1 texture compression format in RGBA form with 1 bit
	// reserved for alpha (i.e. fully transparent or fully opaque per-pixel
	// transparency).
	DXT1RGBA

	// DXT3 is a RGBA texture compression format with four bits per pixel
	// reserved for alpha. Each 4x4 block of pixels take up 128-bits of data,
	// as such when compared to a standard 32-bit RGBA format it provides a 4:1
	// compression ratio. Color information stored in DXT3 is mostly the same
	// as DXT1.
	DXT3

	// DXT5 is a RGBA format similar to DXT3 except it compresses the alpha
	// chunk in a similar manner to DXT1's color storage. It provides the same
	// 4:1 compression ratio as DXT3.
	DXT5
)

// Downloadable represents a image that can be downloaded from the graphics
// hardware into system memory (e.g. for taking a screen-shot).
type Downloadable interface {
	// Download should download the given intersecting rectangle of this
	// downloadable image from the graphics hardware into system memory and
	// send it to the complete channel when done.
	//
	// If downloading this texture is impossible (i.e. hardware does not
	// support this) then nil will be sent over the channel and all future
	// attempts to download this texture will fail as well.
	//
	// It should be noted that the downloaded image may not be pixel-identical
	// to the previously uploaded source image of a texture, for instance if
	// texture compression was used it may suffer from compression artifacts,
	// etc.
	//
	// Only a texture created from render-to-texture is guaranteed to succeed,
	// others may not (esp. compressed textures). Most renderers support
	// downloading RGB/A textures and sometimes depth/alpha ones.
	Download(r image.Rectangle, complete chan image.Image)
}

// NativeTexture represents the native object of a *Texture, the renderer is
// responsible for creating these and fulfilling the interface.
type NativeTexture interface {
	Destroyable
	Downloadable

	// ChosenFormat tells the texture format chosen by the renderer for storing
	// this texture on the graphics device. It may differ from the Texture's
	// Format field only if the graphics device does not support that format.
	ChosenFormat() TexFormat
}

// Texture represents a single 2D texture that may be applied to a mesh for
// drawing.
//
// Clients are responsible for utilizing the RWMutex of the texture when using
// it or invoking methods.
type Texture struct {
	sync.RWMutex

	// The native object of this texture. Once loaded the renderer using this
	// texture must assign a value to this field. Typically clients should not
	// assign values to this field at all.
	NativeTexture

	// Weather or not this texture is currently loaded or not.
	Loaded bool

	// If true then when this texture is loaded the data image source of it
	// will be kept instead of being set to nil (which allows it to be garbage
	// collected).
	KeepDataOnLoad bool

	// The bounds of the texture, in the case of a texture loaded from a image
	// this should be set to the image's bounds. In the case of rendering to a
	// texture this should be set to the desired canvas resolution.
	Bounds image.Rectangle

	// The source image of the texture, may be nil (i.e. in the case of render
	// to texture, unless downloaded).
	Source image.Image

	// The texture format to use for storing this texture on the GPU, which may
	// result in lossy conversions (e.g. RGB would lose the alpha channel, etc).
	//
	// If the format is not supported then the renderer may use an image format
	// that is similar and is supported (and the format chosen by the renderer
	// can be determined via NativeTexture's ChosenFormat method).
	Format TexFormat

	// The U and V wrap modes of this texture.
	WrapU, WrapV TexWrap

	// The color of the border when a wrap mode is set to BorderColor.
	BorderColor Color

	// The texture filtering used for minification and magnification of the
	// texture.
	MinFilter, MagFilter TexFilter
}

// Copy returns a new copy of this Texture. Explicitly not copied over is the
// native texture, the OnLoad slice, the Loaded status, and the source image
// (because the image type is not strictly known). Because the texture's source
// image is not copied over, you may want to copy it directly over yourself.
//
// The texture's read lock must be held for this method to operate safely.
func (t *Texture) Copy() *Texture {
	return &Texture{
		sync.RWMutex{},
		nil,   // Native texture -- not copied.
		false, // Loaded status -- not copied.
		t.KeepDataOnLoad,
		t.Bounds,
		nil, // Source image -- not copied.
		t.Format,
		t.WrapU,
		t.WrapV,
		t.BorderColor,
		t.MinFilter,
		t.MagFilter,
	}
}

// ClearData sets the data source image, t.Source, of this texture to nil if
// t.KeepDataOnLoad is set to false.
//
// The texture's write lock must be held for this method to operate safely.
func (t *Texture) ClearData() {
	if !t.KeepDataOnLoad {
		t.Source = nil
	}
}

// Reset resets this texture to it's default (NewTexture) state.
//
// The textures's write lock must be held for this method to operate safely.
func (t *Texture) Reset() {
	t.NativeTexture = nil
	t.Loaded = false
	t.KeepDataOnLoad = false
	t.Bounds = image.Rectangle{}
	t.Source = nil
	t.Format = RGBA
	t.WrapU = 0
	t.WrapV = 0
	t.BorderColor = Color{}
	t.MinFilter = 0
	t.MagFilter = 0
}

// Destroy destroys this texture for use by other callees to NewTexture. You
// must not use it after calling this method. This makes an implicit call to
// t.NativeTexture.Destroy.
//
// The texture's write lock must be held for this method to operate safely.
func (t *Texture) Destroy() {
	if t.NativeTexture != nil {
		t.NativeTexture.Destroy()
	}
	t.Reset()
	texturePool.Put(t)
}

var texturePool = sync.Pool{
	New: func() interface{} {
		return &Texture{
			Format: RGBA,
		}
	},
}

// NewTexture returns a new, initialized *Texture object.
func NewTexture() *Texture {
	return texturePool.Get().(*Texture)
}
