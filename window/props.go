// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package window

import (
	"fmt"
	"sync"

	"azul3d.org/gfx.v2"
)

// Props represents window properties. Properties are safe for use concurrently
// from multiple goroutines.
type Props struct {
	l                                                 sync.RWMutex
	title                                             string
	width, height, fbWidth, fbHeight, x, y            int
	cursorX, cursorY                                  float64
	fullscreen, shouldClose, visible, decorated       bool
	minimized, focused, vsync, resizable, alwaysOnTop bool
	cursorGrabbed                                     bool
	precision                                         gfx.Precision
}

// String returns a string like:
//  "Window(Title="Hello World!", Fullscreen=false)"
func (p *Props) String() string {
	p.l.RLock()
	str := fmt.Sprintf("Window(Title=%q, Fullscreen=%q)", p.title, p.fullscreen)
	p.l.RUnlock()
	return str
}

// SetTitle sets the title of the window. The backend will replace the first
// string in the title matching "{FPS}" with the actual frames per second.
//
// For example, a title "Hello World - {FPS}" would end up as:
//  "Hello world - 60FPS"
func (p *Props) SetTitle(title string) {
	p.l.Lock()
	p.title = title
	p.l.Unlock()
}

// Title returns the title of the window, as previously set via SetTitle.
func (p *Props) Title() string {
	p.l.RLock()
	title := p.title
	p.l.RUnlock()
	return title
}

// SetFullscreen sets whether or not the window is fullscreen.
func (p *Props) SetFullscreen(fullscreen bool) {
	p.l.Lock()
	p.fullscreen = fullscreen
	p.l.Unlock()
}

// Fullscreen tells whether or not the window is fullscreen.
func (p *Props) Fullscreen() bool {
	p.l.RLock()
	fullscreen := p.fullscreen
	p.l.RUnlock()
	return fullscreen
}

// SetFramebufferSize sets the size of the framebuffer in pixels. Each value is
// clamped to at least a value of 1.
//
// Only the window owner should ever set the framebuffer size.
func (p *Props) SetFramebufferSize(width, height int) {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	p.l.Lock()
	p.fbWidth = width
	p.fbHeight = height
	p.l.Unlock()
}

// FramebufferSize returns the size of the framebuffer in pixels.
func (p *Props) FramebufferSize() (width, height int) {
	p.l.RLock()
	width = p.fbWidth
	height = p.fbHeight
	p.l.RUnlock()
	return
}

// SetSize sets the size of the window in screen coordinates. Each value is
// clamped to at least a value of 1.
func (p *Props) SetSize(width, height int) {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	p.l.Lock()
	p.width = width
	p.height = height
	p.l.Unlock()
}

// Size returns the size of the window in screen coordinates.
func (p *Props) Size() (width, height int) {
	p.l.RLock()
	width = p.width
	height = p.height
	p.l.RUnlock()
	return
}

// SetPos sets the position of the upper-left corner of the client area of the
// window in screen coordinates.
//
// A special value of x=-1, y=-1 means to center the window on the screen.
func (p *Props) SetPos(x, y int) {
	p.l.Lock()
	p.x = x
	p.y = y
	p.l.Unlock()
}

// Pos returns the position of the upper-left corner of the client area of the
// window in screen coordinates.
func (p *Props) Pos() (x, y int) {
	p.l.RLock()
	x = p.x
	y = p.y
	p.l.RUnlock()
	return
}

// SetCursorPos sets the position of the mouse cursor.
//
// A special value of x=-1.0, y=-1.0 means to not move the mouse cursor at all.
func (p *Props) SetCursorPos(x, y float64) {
	p.l.Lock()
	p.cursorX = x
	p.cursorY = y
	p.l.Unlock()
}

// CursorPos returns the position of the mouse cursor.
func (p *Props) CursorPos() (x, y float64) {
	p.l.RLock()
	x = p.cursorX
	y = p.cursorY
	p.l.RUnlock()
	return
}

// SetShouldClose sets whether the window should close or not when the user
// tries to close the window.
func (p *Props) SetShouldClose(shouldClose bool) {
	p.l.Lock()
	p.shouldClose = shouldClose
	p.l.Unlock()
}

// ShouldClose tells if the window will close or not when the user tries to
// close the window.
func (p *Props) ShouldClose() bool {
	p.l.RLock()
	shouldClose := p.shouldClose
	p.l.RUnlock()
	return shouldClose
}

// SetVisible sets whether or not the window is visible or hidden.
func (p *Props) SetVisible(visible bool) {
	p.l.Lock()
	p.visible = visible
	p.l.Unlock()
}

// Visible tells whether or not the window is visible or hidden.
func (p *Props) Visible() bool {
	p.l.RLock()
	visible := p.visible
	p.l.RUnlock()
	return visible
}

// SetMinimized sets whether or not the window is minimized.
func (p *Props) SetMinimized(minimized bool) {
	p.l.Lock()
	p.minimized = minimized
	p.l.Unlock()
}

// Minimized tells whether or not the window is minimized.
func (p *Props) Minimized() bool {
	p.l.RLock()
	minimized := p.minimized
	p.l.RUnlock()
	return minimized
}

// SetVSync turns on or off vertical refresh rate synchronization (vsync).
func (p *Props) SetVSync(vsync bool) {
	p.l.Lock()
	p.vsync = vsync
	p.l.Unlock()
}

// VSync tells if vertical refresh rate synchronization (vsync) is on or off.
func (p *Props) VSync() bool {
	p.l.RLock()
	vsync := p.vsync
	p.l.RUnlock()
	return vsync
}

// SetFocused sets whether or not the window has focus.
func (p *Props) SetFocused(focused bool) {
	p.l.Lock()
	p.focused = focused
	p.l.Unlock()
}

// Focused tells whether or not the window has focus.
func (p *Props) Focused() bool {
	p.l.RLock()
	focused := p.focused
	p.l.RUnlock()
	return focused
}

// SetResizable sets whether or not the window can be resized.
func (p *Props) SetResizable(resizable bool) {
	p.l.Lock()
	p.resizable = resizable
	p.l.Unlock()
}

// Resizable tells whether or not the window can be resized.
func (p *Props) Resizable() bool {
	p.l.RLock()
	resizable := p.resizable
	p.l.RUnlock()
	return resizable
}

// SetDecorated sets whether or not the window has it's decorations shown.
func (p *Props) SetDecorated(decorated bool) {
	p.l.Lock()
	p.decorated = decorated
	p.l.Unlock()
}

// Decorated tells whether or not the window has it's decorations shown.
func (p *Props) Decorated() bool {
	p.l.RLock()
	decorated := p.decorated
	p.l.RUnlock()
	return decorated
}

// SetAlwaysOnTop sets whether or not the window is always on top of other
// windows.
func (p *Props) SetAlwaysOnTop(alwaysOnTop bool) {
	p.l.Lock()
	p.alwaysOnTop = alwaysOnTop
	p.l.Unlock()
}

// AlwaysOnTop tells whether or not the window is set to be always on top of
// other windows.
func (p *Props) AlwaysOnTop() bool {
	p.l.RLock()
	alwaysOnTop := p.alwaysOnTop
	p.l.RUnlock()
	return alwaysOnTop
}

// SetCursorGrabbed sets whether or not the cursor should be grabbed. If the
// cursor is grabbed, it is hidden from sight and cannot leave the window.
//
// When grabbed, a window generates CursorMoved events with Delta=true, this
// is useful for e.g. FPS style cameras.
func (p *Props) SetCursorGrabbed(grabbed bool) {
	p.l.Lock()
	p.cursorGrabbed = grabbed
	p.l.Unlock()
}

// CursorGrabbed returns whether or not the cursor is grabbed.
func (p *Props) CursorGrabbed() bool {
	p.l.RLock()
	grabbed := p.cursorGrabbed
	p.l.RUnlock()
	return grabbed
}

// SetPrecision sets the framebuffer precision to be requested when the window
// is created.
//
// Requesting a specific framebuffer precision is simply that -- a request. No
// guarantee is made that you will receive that precision, as it is completely
// hardware dependent.
//
// To check what framebuffer precision you actually receive, look at the
// precision of the renderer's canvas:
//
//  r.Canvas.Precision()
//
func (p *Props) SetPrecision(precision gfx.Precision) {
	p.l.Lock()
	p.precision = precision
	p.l.Unlock()
}

// Precision returns the framebuffer precision that is to be requested when the
// window is created.
//
// As mentioned by the SetPrecision documentation, the precision returned by
// this function is the one you request -- not strictly the precision that you
// will receive (as that is hardware dependent).
//
// To check what framebuffer precision you actually receive, look at the
// precision of the renderer's canvas:
//
//  r.Canvas.Precision()
//
func (p *Props) Precision() gfx.Precision {
	p.l.RLock()
	precision := p.precision
	p.l.RUnlock()
	return precision
}

// NewProps returns a new initialized set of window properties. The default
// values for each property are as follows:
//  Title: "Azul3D - {FPS}"
//  Size: 800x450
//  Pos: -1, -1 (centered on screen)
//  CursorPos: -1.0, -1.0 (current position)
//  ShouldClose: true
//  Visible: true
//  Minimized: false
//  Fullscreen: false
//  Focused: true
//  VSync: true
//  Resizable: true
//  Decorated: true
//  AlwaysOnTop: false
//  CursorGrabbed: false
//  FramebufferSize: 1x1 (set via window owner)
//  Precision: gfx.Precision{
//      RedBits: 8, GreenBits: 8, BlueBits: 8, AlphaBits: 0,
//      DepthBits: 24,
//      StencilBits: 0,
//      Samples: 2,
//  }
func NewProps() *Props {
	return &Props{
		title:         "Azul3D - {FPS}",
		width:         800,
		height:        450,
		fbWidth:       1,
		fbHeight:      1,
		x:             -1,
		y:             -1,
		cursorX:       -1.0,
		cursorY:       -1.0,
		shouldClose:   true,
		visible:       true,
		minimized:     false,
		fullscreen:    false,
		focused:       true,
		vsync:         true,
		resizable:     true,
		decorated:     true,
		alwaysOnTop:   false,
		cursorGrabbed: false,
		precision: gfx.Precision{
			RedBits: 8, GreenBits: 8, BlueBits: 8, AlphaBits: 8,
			DepthBits: 24,
			Samples:   2,
		},
	}
}

// DefaultProps is the default set of window properties. You may modify them as
// you see fit.
//
// They are used in place of nil properties (e.g. see the Run function).
var DefaultProps = NewProps()
