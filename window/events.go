// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package window

import (
	"fmt"
	"time"
)

// Event represents an event of some sort. The only requirement is that the
// event specify the point in time at which it happened.
//
// Using a type assertion or a type switch you can determine the actualy type
// of event which contains much more information:
//  select ev := event.(type){
//  case *keyboard.StateEvent:
//      fmt.Println("The keyboard button", ev.Key, "is now", ev.State)
//      // example: "The keyboard button keyboard.A is now keyboard.Down"
//  }
type Event interface {
	Time() time.Time
}

// Close is sent when the user requests that the application window be closed,
// using the exit button or a quick-key combination like Alt + F4, etc.
type Close struct {
	T time.Time
}

// String returns a string representation of this event.
func (ev Close) String() string {
	return fmt.Sprintf("Close(Time=%v)", ev.T)
}

// Time implements the Event interface.
func (ev Close) Time() time.Time {
	return ev.T
}

// Damaged is sent when the window's client area has been damaged and the window
// needs to be redrawn.
type Damaged struct {
	T time.Time
}

// String returns a string representation of this event.
func (ev Damaged) String() string {
	return fmt.Sprintf("Damaged(Time=%v)", ev.T)
}

// Time implements the Event interface.
func (ev Damaged) Time() time.Time {
	return ev.T
}

// CursorMoved is sent when the user has moved the mouse cursor.
type CursorMoved struct {
	// Position of cursor relative to the upper-left corner of the window.
	X, Y float64

	// Whether or not the event's X and Y values are actually relative delta
	// values (e.g. for a FPS style camera).
	Delta bool

	T time.Time
}

// String returns a string representation of this event.
func (ev CursorMoved) String() string {
	return fmt.Sprintf("CursorMoved(X=%f, Y=%f, Time=%v)", ev.X, ev.Y, ev.T)
}

// Time implements the Event interface.
func (ev CursorMoved) Time() time.Time {
	return ev.T
}

// CursorEnter is an event where the user moved the mouse cursor inside of the
// window.
type CursorEnter struct {
	T time.Time
}

// String returns a string representation of this event.
func (ev CursorEnter) String() string {
	return fmt.Sprintf("CursorEnter(Time=%v)", ev.T)
}

// Time implements the Event interface.
func (ev CursorEnter) Time() time.Time {
	return ev.T
}

// CursorExit is an event where the user moved the mouse cursor outside of the
// window.
type CursorExit struct {
	T time.Time
}

// String returns a string representation of this event.
func (ev CursorExit) String() string {
	return fmt.Sprintf("CursorExit(Time=%v)", ev.T)
}

// Time implements the Event interface.
func (ev CursorExit) Time() time.Time {
	return ev.T
}

// Minimized is an event where the user minimized the window.
type Minimized struct {
	T time.Time
}

// String returns a string representation of this event.
func (ev Minimized) String() string {
	return fmt.Sprintf("Minimized(Time=%v)", ev.T)
}

// Time implements the Event interface.
func (ev Minimized) Time() time.Time {
	return ev.T
}

// Restored is an event where the user restored (un-minimized) the window.
type Restored struct {
	T time.Time
}

// String returns a string representation of this event.
func (ev Restored) String() string {
	return fmt.Sprintf("Restored(Time=%v)", ev.T)
}

// Time implements the Event interface.
func (ev Restored) Time() time.Time {
	return ev.T
}

// GainedFocus is an event where the window has gained focus.
type GainedFocus struct {
	T time.Time
}

// String returns a string representation of this event.
func (ev GainedFocus) String() string {
	return fmt.Sprintf("GainedFocus(Time=%v)", ev.T)
}

// Time implements the Event interface.
func (ev GainedFocus) Time() time.Time {
	return ev.T
}

// LostFocus is an event where the window has lost focus.
type LostFocus struct {
	T time.Time
}

// String returns a string representation of this event.
func (ev LostFocus) String() string {
	return fmt.Sprintf("LostFocus(Time=%v)", ev.T)
}

// Time implements the Event interface.
func (ev LostFocus) Time() time.Time {
	return ev.T
}

// Moved is an event where the user changed the position of the window.
type Moved struct {
	// Position of the window's client area, relative to the top-left of the
	// screen.
	X, Y int

	T time.Time
}

// String returns a string representation of this event.
func (ev Moved) String() string {
	return fmt.Sprintf("Moved(X=%v, Y=%v, Time=%v)", ev.X, ev.Y, ev.T)
}

// Time implements the Event interface.
func (ev Moved) Time() time.Time {
	return ev.T
}

// Resized is an event where the user changed the size of the window.
type Resized struct {
	// Size of the window in screen coordinates.
	Width, Height int

	T time.Time
}

// String returns a string representation of this event.
func (ev Resized) String() string {
	return fmt.Sprintf("Resized(Width=%v, Height=%v, Time=%v)", ev.Width, ev.Height, ev.T)
}

// Time implements the Event interface.
func (ev Resized) Time() time.Time {
	return ev.T
}

// FramebufferResized is an event where the framebuffer of the window has been
// resized.
type FramebufferResized struct {
	// Size of the framebuffer in pixels.
	Width, Height int

	T time.Time
}

// String returns a string representation of this event.
func (ev FramebufferResized) String() string {
	return fmt.Sprintf("FramebufferResized(Width=%v, Height=%v, Time=%v)", ev.Width, ev.Height, ev.T)
}

// Time implements the Event interface.
func (ev FramebufferResized) Time() time.Time {
	return ev.T
}

// ItemsDropped is an event where the user dropped an item (or multiple items)
// onto the window.
type ItemsDropped struct {
	Items []string

	T time.Time
}

// String returns a string representation of this event.
func (ev ItemsDropped) String() string {
	return fmt.Sprintf("ItemsDropped(Items=%v, Time=%v)", ev.Items, ev.T)
}

// Time implements the Event interface.
func (ev ItemsDropped) Time() time.Time {
	return ev.T
}
