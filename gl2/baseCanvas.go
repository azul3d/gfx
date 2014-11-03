// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"image"
	"sync"

	"azul3d.org/gfx.v1"
)

// baseCanvas implements portions of a gfx.Canvas shared between Renderer and
// rttCanvas. It implements all portions of gfx.Canvas except:
//  Downloadable
//  Clear(r image.Rectangle, bg Color)
//  ClearDepth(r image.Rectangle, depth float64)
//  ClearStencil(r image.Rectangle, stencil int)
//  Draw(r image.Rectangle, o *Object, c *Camera)
//  QueryWait()
//  Render()
type baseCanvas struct {
	sync.RWMutex
	msaa      bool            // The MSAA state.
	precision gfx.Precision   // The precision of this canvas.
	bounds    image.Rectangle // The bounding rectangle of this canvas.
}

// Implements gfx.Canvas interface.
func (c *baseCanvas) SetMSAA(msaa bool) {
	c.Lock()
	c.msaa = msaa
	c.Unlock()
}

// Implements gfx.Canvas interface.
func (c *baseCanvas) MSAA() bool {
	c.RLock()
	msaa := c.msaa
	c.RUnlock()
	return msaa
}

// Implements gfx.Canvas interface.
func (c *baseCanvas) Precision() gfx.Precision {
	c.RLock()
	precision := c.precision
	c.RUnlock()
	return precision
}

func (c *baseCanvas) setBounds(b image.Rectangle) {
	c.Lock()
	c.bounds = b
	c.Unlock()
}

// Implements the gfx.Canvas interface.
func (c *baseCanvas) Bounds() image.Rectangle {
	c.RLock()
	bounds := c.bounds
	c.RUnlock()
	return bounds
}
