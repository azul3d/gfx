// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package window

import (
	"math"
	"runtime"

	"azul3d.org/gfx.v1"
	"azul3d.org/keyboard.v1"
	"azul3d.org/mouse.v1"
)

// EventMask is a bitmask of event types. They can be combined, for instance:
//  mask := GenericEvents
//  mask |= MouseEvents
//  mask |= KeyboardEvents
// would select generic, mouse, and keyboard events.
type EventMask uint32

const (
	// Event mask matching no events at all.
	NoEvents EventMask = 0

	// Each event mask below matches it's corresponding event (defined in this
	// package) without the `Events` suffix.
	CloseEvents              EventMask = 1 << 0
	DamagedEvents            EventMask = 1 << 1
	CursorMovedEvents        EventMask = 1 << 2
	CursorEnterEvents        EventMask = 1 << 3
	CursorExitEvents         EventMask = 1 << 4
	MinimizedEvents          EventMask = 1 << 5
	RestoredEvents           EventMask = 1 << 6
	GainedFocusEvents        EventMask = 1 << 7
	LostFocusEvents          EventMask = 1 << 8
	MovedEvents              EventMask = 1 << 9
	ResizedEvents            EventMask = 1 << 10
	FramebufferResizedEvents EventMask = 1 << 11
	ItemsDroppedEvents       EventMask = 1 << 12

	// Event mask for the mouse.Event event type.
	MouseEvents EventMask = 1 << 13

	// Event mask for the mouse.Scrolled event type.
	MouseScrolledEvents EventMask = 1 << 14

	// Event mask for the keyboard.TypedEvent event type.
	KeyboardTypedEvents EventMask = 1 << 15

	// Event mask for the keyboard.StateEvent event type.
	KeyboardStateEvents EventMask = 1 << 16

	// Event mask matching all possible events.
	AllEvents = EventMask(math.MaxUint32)
)

// Window represents a single window that graphics can be rendered to. The
// window is safe for use concurrently from multiple goroutines.
type Window interface {
	// Props returns the window's properties.
	Props() *Props

	// Request makes a request to use a new set of properties, p. It is then
	// reccomended to make changes to the window using something like:
	//  props := window.Props()
	//  props.SetTitle("Hello World!")
	//  props.SetSize(640, 480)
	//  window.Request(props)
	//
	// Interpretation of the given properties is left strictly up to the
	// platform dependant implementation (for instance, on Android you cannot
	// set the window's size, so instead a request for this is simply ignored.
	Request(p *Props)

	// Keyboard returns a keyboard watcher for the window. It can be used to
	// tell if certain keyboard buttons are currently held down, for instance:
	//
	//  if w.Keyboard().Down(keyboard.W) {
	//      fmt.Println("The W key is currently held down")
	//  }
	Keyboard() *keyboard.Watcher

	// Mouse returns a mouse watcher for the window. It can be used to tell if
	// certain mouse buttons are currently held down, for instance:
	//
	//  if w.Mouse().Down(mouse.Left) {
	//      fmt.Println("The left mouse button is currently held down")
	//  }
	Mouse() *mouse.Watcher

	// SetClipboard sets the clipboard string.
	SetClipboard(clipboard string)

	// Clipboard returns the clipboard string.
	Clipboard() string

	// Notify causes the window to relay window events to ch based on the event
	// mask.
	//
	// The special event mask NoEvents causes the window to stop relaying any
	// events to the given channel. You should always perform this action when
	// you are done using the event channel.
	//
	// The window will not block sending events to ch: the caller must ensure
	// that ch has a sufficient amount of buffer space to keep up with the
	// event rate.
	//
	// If you only expect to receive a single event like Close then a buffer
	// size of one is acceptable.
	//
	// You are allowed to make multiple calls to this method with the same
	// channel, if you do then the same event will be sent over the channel
	// multiple times. When you no longer want the channel to receive events
	// then call this function again with NoEvents:
	//  w.Notify(ch, NoEvents)
	//
	// Multiple calls to Events with different channels works as you would
	// expect, each channel receives a copy of the events independant of other
	// ones.
	//
	// Warning: Many people use high-precision mice, some which can reach well
	// above 1000hz, so for cursor events a generous buffer size (256, etc) is
	// highly recommended.
	//
	// Warning: Depending on the operating system, window manager, etc, some of
	// the events may never be sent or may only be sent sporiadically, so plan
	// for this.
	Notify(ch chan<- Event, m EventMask)

	// Close closes the window, and causes Run() to return.
	Close()
}

// Run opens a window with the given properties and runs the given graphics
// loop in a separate goroutine.
//
// Interpretation of the given properties is left strictly up to the platform
// dependant implementation (for instance, on Android you cannot set the
// window's size so it is simply ignored).
//
// Requesting a specific framebuffer configuration via Props.SetPrecision is
// just a request. You may be given some other configuration (most likely one
// closest to it). You can check what you received by looking at:
//  r.Canvas.Precision()
//
// If the properties are nil, DefaultProps is used instead.
func Run(gfxLoop func(w Window, r gfx.Renderer), p *Props) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if p == nil {
		p = NewProps()
	}
	if gfxLoop == nil {
		panic("window: nil graphics loop function!")
	}
	doRun(gfxLoop, p)
}
