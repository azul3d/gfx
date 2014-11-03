// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package window is the easiest way to open a window and render graphics.
//
// The window package effectively provides an easy cross-platform way to create
// and configure a window, as well as manipulate it and receive user input from
// it efficiently.
//
// The window package lays the groundwork for developing applications that are
// truly cross-platform (i.e. desktop and mobile applications) using nearly the
// exact same API in the future. As such, the window package is mostly just an
// abstraction layer for GLFW on desktop operating systems.
//
// If you truly need features not found in this package then you might be best
// using GLFW directly if you intend to just write desktop applications. Some
// features are not supported by this package intentionally, like multiple
// windows because mobile devices don't have them.
//
// The goal of the window package is not to provide a one solution fits all, it
// is instead to abstract away the tedious parts of writing cross-device
// applications in the future.
//
// The window package is also extremely simple to use:
//  func gfxLoop(w window.Window, r gfx.Renderer) {
//      // Initialization here.
//      for {
//          // Render loop here.
//      }
//  }
//
//  func main() {
//      window.Run(gfxLoop, nil)
//  }
package window
