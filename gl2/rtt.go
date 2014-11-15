// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"errors"
	"fmt"
	"image"
	"runtime"
	"sync"

	"azul3d.org/gfx.v2"
	"azul3d.org/gfx.v2/gl2/internal/gl"
)

func (r *Renderer) freeFBOs() {
	// Lock the list.
	r.fbosToFree.Lock()

	if len(r.fbosToFree.slice) > 0 {
		// Free the FBOs.
		gl.DeleteFramebuffers(int32(len(r.fbosToFree.slice)), &r.fbosToFree.slice[0])

		// Flush and execute OpenGL commands.
		gl.Flush()
		//gl.Execute()
	}

	// Slice to zero, and unlock.
	r.fbosToFree.slice = r.fbosToFree.slice[:0]
	r.fbosToFree.Unlock()
}

func (r *Renderer) freeRenderbuffers() {
	// Lock the list.
	r.renderbuffersToFree.Lock()

	if len(r.renderbuffersToFree.slice) > 0 {
		// Free the FBOs.
		gl.DeleteRenderbuffers(int32(len(r.renderbuffersToFree.slice)), &r.renderbuffersToFree.slice[0])

		// Flush and execute OpenGL commands.
		gl.Flush()
		//gl.Execute()
	}

	// Slice to zero, and unlock.
	r.renderbuffersToFree.slice = r.renderbuffersToFree.slice[:0]
	r.renderbuffersToFree.Unlock()
}

// rttCanvas is the gfx.Canvas returned by RenderToTexture.
type rttCanvas struct {
	*baseCanvas
	r   *Renderer
	cfg gfx.RTTConfig

	// Frame buffer ID.
	fbo uint32

	// Render buffer ID's (rbColor is only a valid render buffer if e.g. the
	// cfg.Color field is nil).
	//
	// rbDepthAndStencil is only set if cfg.DepthFormat.IsCombined()
	rbColor, rbDepth, rbStencil, rbDepthAndStencil uint32

	// Decremented until zero, then all textures are free'd and all of the
	// canvas methods are no-op.
	textureCount struct {
		sync.RWMutex
		count int
	}
}

func (r *rttCanvas) freeTexture(n *nativeTexture) {
	r.textureCount.Lock()
	if r.textureCount.count == 0 {
		r.textureCount.Unlock()
		return
	}
	r.textureCount.count--
	if r.textureCount.count == 0 {
		// Everything is free now.
		if r.cfg.Color != nil {
			finalizeTexture(r.cfg.Color.NativeTexture.(*nativeTexture))
		}
		if r.cfg.Depth != nil {
			finalizeTexture(r.cfg.Depth.NativeTexture.(*nativeTexture))
		}
		if r.cfg.Stencil != nil {
			finalizeTexture(r.cfg.Stencil.NativeTexture.(*nativeTexture))
		}

		// Add the FBO to the free list.
		if r.fbo != 0 {
			r.r.fbosToFree.Lock()
			r.r.fbosToFree.slice = append(r.r.fbosToFree.slice, r.fbo)
			r.r.fbosToFree.Unlock()
		}

		// Add the render buffers to the free list.
		freeRb := func(id uint32) {
			if id == 0 {
				return
			}
			r.r.renderbuffersToFree.Lock()
			r.r.renderbuffersToFree.slice = append(r.r.renderbuffersToFree.slice, id)
			r.r.renderbuffersToFree.Unlock()
		}
		freeRb(r.rbColor)
		freeRb(r.rbDepth)
		freeRb(r.rbStencil)
		freeRb(r.rbDepthAndStencil)
	}
	r.textureCount.Unlock()
}

func finalizeRTTTexture(n *nativeTexture) {
	n.rttCanvas.freeTexture(n)
}

// Tells if all textures have been free'd and canvas methods are considered
// no-op.
func (r *rttCanvas) noop() bool {
	r.textureCount.RLock()
	if r.textureCount.count == 0 {
		return true
	}
	r.textureCount.RUnlock()
	return false
}

// Short methods that just call the hooked methods. We insert calls to rttBegin
// and rttEnd (they are executed via r.r.RenderExec, i.e. legal for GL
// rendering commands to be invoked).

// Implements gfx.Canvas interface.
func (r *rttCanvas) Clear(rect image.Rectangle, bg gfx.Color) {
	if r.noop() {
		return
	}
	r.r.hookedClear(rect, bg, r.rttBegin, r.rttEnd)
}

// Implements gfx.Canvas interface.
func (r *rttCanvas) ClearDepth(rect image.Rectangle, depth float64) {
	r.r.hookedClearDepth(rect, depth, r.rttBegin, r.rttEnd)
}

// Implements gfx.Canvas interface.
func (r *rttCanvas) ClearStencil(rect image.Rectangle, stencil int) {
	r.r.hookedClearStencil(rect, stencil, r.rttBegin, r.rttEnd)
}

// Implements gfx.Canvas interface.
func (r *rttCanvas) Draw(rect image.Rectangle, o *gfx.Object, c *gfx.Camera) {
	r.r.hookedDraw(rect, o, c, r.rttBegin, r.rttEnd)
}

// Implements gfx.Canvas interface.
func (r *rttCanvas) QueryWait() {
	r.r.hookedQueryWait(r.rttBegin, r.rttEnd)
}

// Implements gfx.Canvas interface.
func (r *rttCanvas) Render() {
	r.r.hookedRender(nil, func() {
		// Generate mipmaps for any texture with a mipmapped format. This must
		// be done here because the texture has just been rendered to.
		do := func(t *gfx.Texture) {
			if t == nil || !t.MinFilter.Mipmapped() {
				return
			}
			n := t.NativeTexture.(*nativeTexture)
			gl.BindTexture(gl.TEXTURE_2D, n.id)
			gl.GenerateMipmap(gl.TEXTURE_2D)
		}
		do(r.cfg.Color)
		do(r.cfg.Depth)
		do(r.cfg.Stencil)
		gl.BindTexture(gl.TEXTURE_2D, 0)
	})
}

// Implements gfx.Downloadable interface.
func (r *rttCanvas) Download(rect image.Rectangle, complete chan image.Image) {
	r.r.hookedDownload(rect, complete, r.rttBegin, r.rttEnd)
}

func (r *rttCanvas) rttBegin() {
	r.r.rttCanvas = r

	// Bind the framebuffer object.
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo)
}

func (r *rttCanvas) rttEnd() {
	r.r.rttCanvas = nil

	// Unbind the framebuffer object.
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

var (
	errFramebufferUndefined                   = errors.New("GL_FRAMEBUFFER_UNDEFINED")
	errFramebufferIncompleteAttachment        = errors.New("GL_FRAMEBUFFER_INCOMPLETE_ATTACHMENT")
	errFramebufferIncompleteMissingAttachment = errors.New("GL_FRAMEBUFFER_INCOMPLETE_MISSING_ATTACHMENT")
	errFramebufferIncompleteDrawBuffer        = errors.New("GL_FRAMEBUFFER_INCOMPLETE_DRAW_BUFFER")
	errFramebufferIncompleteReadBuffer        = errors.New("GL_FRAMEBUFFER_INCOMPLETE_READ_BUFFER")
	errFramebufferUnsupported                 = errors.New("GL_FRAMEBUFFER_UNSUPPORTED")
	errFramebufferIncompleteMultisample       = errors.New("GL_FRAMEBUFFER_INCOMPLETE_MULTISAMPLE")
	errFramebufferIncompleteLayerTargets      = errors.New("GL_FRAMEBUFFER_INCOMPLETE_LAYER_TARGETS")
)

func checkFramebufferError(target uint32) error {
	err := gl.CheckFramebufferStatus(target)
	switch err {
	case gl.FRAMEBUFFER_COMPLETE:
		return nil
	case gl.FRAMEBUFFER_UNDEFINED:
		return errFramebufferUndefined
	case gl.FRAMEBUFFER_INCOMPLETE_ATTACHMENT:
		return errFramebufferIncompleteAttachment
	case gl.FRAMEBUFFER_INCOMPLETE_MISSING_ATTACHMENT:
		return errFramebufferIncompleteMissingAttachment
	case gl.FRAMEBUFFER_INCOMPLETE_DRAW_BUFFER:
		return errFramebufferIncompleteDrawBuffer
	case gl.FRAMEBUFFER_INCOMPLETE_READ_BUFFER:
		return errFramebufferIncompleteReadBuffer
	case gl.FRAMEBUFFER_UNSUPPORTED:
		return errFramebufferUnsupported
	case gl.FRAMEBUFFER_INCOMPLETE_MULTISAMPLE:
		return errFramebufferIncompleteMultisample
		// TODO(slimsag): determine if this is needed
		//case gl.FRAMEBUFFER_INCOMPLETE_LAYER_TARGETS:
		//	return errFramebufferIncompleteLayerTargets
	}
	return fmt.Errorf("unkown framebuffer error (%d)", err)
}

// RenderToTexture implements the gfx.Renderer interface.
func (r *Renderer) RenderToTexture(cfg gfx.RTTConfig) gfx.Canvas {
	if !cfg.Valid() {
		panic("RenderToTexture(): Configuration is invalid!")
	}

	if !r.glArbFramebufferObject {
		// We don't have GL_ARB_framebuffer_object extension, we can't do this
		// at all.
		return nil
	}

	// Find OpenGL versions of formats.
	colorFormat, ok := r.rttTexFormats[cfg.ColorFormat]
	if cfg.ColorFormat != gfx.ZeroTexFormat && !ok {
		return nil
	}
	depthFormat, ok := r.rttDSFormats[cfg.DepthFormat]
	if cfg.DepthFormat != gfx.ZeroDSFormat && !ok {
		return nil
	}
	stencilFormat, ok := r.rttDSFormats[cfg.StencilFormat]
	if cfg.StencilFormat != gfx.ZeroDSFormat && !ok {
		return nil
	}

	// Lock textures.
	if cfg.Color != nil {
		cfg.Color.Lock()
	}
	if cfg.Depth != nil {
		cfg.Depth.Lock()
	}
	if cfg.Stencil != nil {
		cfg.Stencil.Lock()
	}

	// Choose correct bounds.
	bounds := cfg.Bounds
	if bounds.Empty() {
		bounds = r.baseCanvas.Bounds()
	}

	// Create the RTT canvas.
	cr, cg, cb, ca := cfg.ColorFormat.Bits()
	canvas := &rttCanvas{
		baseCanvas: &baseCanvas{
			msaa: true,
			precision: gfx.Precision{
				RedBits: cr, GreenBits: cg, BlueBits: cb, AlphaBits: ca,
				DepthBits:   cfg.DepthFormat.DepthBits(),
				StencilBits: cfg.StencilFormat.StencilBits(),
			},
			bounds: bounds,
		},
		r:   r,
		cfg: cfg,
	}

	var (
		nTexColor, nTexDepth, nTexStencil *nativeTexture
		fbError                           error
	)
	r.RenderExec <- func() bool {
		width := int32(bounds.Dx())
		height := int32(bounds.Dy())

		// Create the FBO.
		gl.GenFramebuffers(1, &canvas.fbo)
		//gl.Execute()
		gl.BindFramebuffer(gl.FRAMEBUFFER, canvas.fbo)

		// Create an OpenGL render buffer for each nil cfg texture. This allows
		// the driver a chance to optimize storage for e.g. a depth buffer when
		// you don't intend to use it as a texture.
		samples := int32(cfg.Samples)
		if cfg.Color == nil && cfg.ColorFormat != gfx.ZeroTexFormat {
			// We do not want a color texture, but we do want a color buffer.
			gl.GenRenderbuffers(1, &canvas.rbColor)
			gl.BindRenderbuffer(gl.RENDERBUFFER, canvas.rbColor)
			gl.RenderbufferStorageMultisample(gl.RENDERBUFFER, samples, uint32(colorFormat), width, height)
			gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.RENDERBUFFER, canvas.rbColor)
		}
		dsCombined := cfg.DepthFormat == cfg.StencilFormat && cfg.DepthFormat.IsCombined()
		if cfg.Depth == nil && cfg.Stencil == nil && dsCombined {
			// We do not want a depth or stencil texture, but we do want a
			// combined depth/stencil buffer.
			gl.GenRenderbuffers(1, &canvas.rbDepthAndStencil)
			gl.BindRenderbuffer(gl.RENDERBUFFER, canvas.rbDepthAndStencil)
			gl.RenderbufferStorageMultisample(gl.RENDERBUFFER, samples, uint32(depthFormat), width, height)
			gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, canvas.rbDepthAndStencil)
			gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.STENCIL_ATTACHMENT, gl.RENDERBUFFER, canvas.rbDepthAndStencil)
		} else {
			if cfg.Depth == nil && cfg.DepthFormat != gfx.ZeroDSFormat {
				// We do not want a depth texture, but we do want a depth buffer.
				gl.GenRenderbuffers(1, &canvas.rbDepth)
				gl.BindRenderbuffer(gl.RENDERBUFFER, canvas.rbDepth)
				gl.RenderbufferStorageMultisample(gl.RENDERBUFFER, samples, uint32(depthFormat), width, height)
				gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, canvas.rbDepth)
			}
			if cfg.Stencil == nil && cfg.StencilFormat != gfx.ZeroDSFormat {
				// We do not want a stencil texture, but we do want a stencil buffer.
				gl.GenRenderbuffers(1, &canvas.rbStencil)
				gl.BindRenderbuffer(gl.RENDERBUFFER, canvas.rbStencil)
				gl.RenderbufferStorageMultisample(gl.RENDERBUFFER, samples, uint32(stencilFormat), width, height)
				gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.STENCIL_ATTACHMENT, gl.RENDERBUFFER, canvas.rbStencil)
			}
		}

		// Create an OpenGL texture for every non-nil cfg texture.
		if cfg.Color != nil && cfg.ColorFormat != gfx.ZeroTexFormat {
			// We want a color texture, not a color buffer.
			nTexColor = newNativeTexture(r, colorFormat, bounds.Dx(), bounds.Dy())
			gl.TexImage2D(gl.TEXTURE_2D, 0, colorFormat, width, height, 0, gl.BGRA, gl.UNSIGNED_BYTE, nil)
			gl.GenerateMipmap(gl.TEXTURE_2D)
			gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, nTexColor.id, 0)
		}
		// Only non-combined depth/stencil formats can render into a texture.
		if !dsCombined {
			if cfg.Depth != nil && cfg.DepthFormat != gfx.ZeroDSFormat {
				// We want a depth texture, not a depth buffer.
				nTexDepth = newNativeTexture(r, depthFormat, bounds.Dx(), bounds.Dy())
				gl.TexImage2D(gl.TEXTURE_2D, 0, depthFormat, width, height, 0, gl.DEPTH_COMPONENT, gl.UNSIGNED_BYTE, nil)
				gl.GenerateMipmap(gl.TEXTURE_2D)
				gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, nTexDepth.id, 0)
			}
		}

		// Check for errors.
		fbError = checkFramebufferError(gl.FRAMEBUFFER)

		// Unbind textures, render buffers, and the FBO.
		gl.BindTexture(gl.TEXTURE_2D, 0)
		gl.BindRenderbuffer(gl.RENDERBUFFER, 0)
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

		// Signal render completion.
		r.renderComplete <- struct{}{}
		return false // No frame was rendered.
	}
	<-r.renderComplete

	if fbError != nil {
		if fbError == errFramebufferUnsupported {
			// Ideally this shouldn't happen, but it could under e.g. strange
			// drivers not supporting a combination of 'supported' formats.
			cfg.Color.Unlock()
			cfg.Depth.Unlock()
			cfg.Stencil.Unlock()
			return nil
		}
		panic(fbError)
	}

	// Finish textures (mark as loaded, clear data slices, unlock).
	finishTexture := func(t *gfx.Texture, dsFmt *gfx.DSFormat, native *nativeTexture) {
		if t == nil {
			return
		}
		if native == nil {
			t.Unlock()
			return
		}
		canvas.textureCount.count++
		// Attach a finalizer to the texture that will later free it.
		runtime.SetFinalizer(native, finalizeRTTTexture)
		native.rttCanvas = canvas
		native.destroyHandler = finalizeRTTTexture
		t.NativeTexture = native
		t.Bounds = bounds
		t.Loaded = true
		t.ClearData()
		t.Unlock()
	}
	finishTexture(cfg.Color, nil, nTexColor)
	finishTexture(cfg.Depth, &cfg.DepthFormat, nTexDepth)
	finishTexture(cfg.Stencil, &cfg.StencilFormat, nTexStencil)

	// OpenGL makes no guarantee about the data existing in the texture until
	// we actually draw something, so clear everything now.
	canvas.Clear(image.Rect(0, 0, 0, 0), gfx.Color{0, 0, 0, 1})
	canvas.ClearDepth(image.Rect(0, 0, 0, 0), 1.0)
	canvas.ClearStencil(image.Rect(0, 0, 0, 0), 0)

	return canvas
}
