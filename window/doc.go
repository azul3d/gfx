// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package window is the easiest way to open a window and render graphics.
//
// Through this package you can easilly create windows that can be rendered to
// via the gfx package. A single API is used and works to target desktop,
// mobile and web platforms.
//
// Simple
//
// The package is extremely easy to use, simply declare a graphics loop and
// call the Run function:
//
//  // gfxLoop is our graphics loop, which runs in a separate goroutine.
//  func gfxLoop(w window.Window, r gfx.Renderer) {
//      // Initialization here.
//
//      // Enter our render loop:
//      for {
//          // TODO: clear and draw to the canvas, r.
//          r.Render()
//      }
//  }
//
//  func main() {
//      // Create a window and run our graphics loop.
//      window.Run(gfxLoop, nil)
//  }
//
// Events
//
// Windows can notify you of the events that they recieve over a channel, which
// allows for user input to be processed in any goroutine completely
// independent of rendering.
//
// A bitmask is used to efficiently select the events that you are interested
// in, rather than you receiving all of them and discarding the ones you don't
// want:
//
//  // Create our events channel with sufficient buffer size.
//  events := make(chan window.Event, 256)
//
//  // Get notified of just mouse and keyboard events.
//  w.Notify(events, window.MouseEvents|window.KeyboardTypedEvents)
//
//  // Less efficient:
//  // w.Notify(events, window.AllEvents)
//
// Properties
//
// Creating window properties (such as a window's title, position, size,
// fullscreen status, etc) is very easy:
//
//  props := window.NewProps()
//  props.SetTitle("Hello World!")
//  props.SetFullscreen(true)
//
// The backend is ultimately what interprets the window properties, for
// instance mobile devices have no concept of positioning windows -- so
// requests to do so are simply ignored.
//
// Both the Run and New functions exposed by this package take a second
// argument which can be a set of window properties (or nil), this lets you
// specify window properties when a window is created:
//
//  // Second argument is *Props, if nil it defaults to DefaultProps.
//  window.Run(gfxLoop, props)
//
//  // Or when creating multiple windows:
//  window.New(gfxLoop, props)
//
// Changing Properties
//
// You can request that window properties be changed after a window has already
// been created, for instance to toggle fullscreen at runtime:
//
//  // Get the current window properties.
//  props := myWindow.Props()
//
//  // Toggle the fullscreen status.
//  props.SetFullscreen(!props.Fullscreen())
//
//  // Request the new window properties.
//  myWindow.Request(props)
//
// FPS in Title
//
// If the window title contains a "{FPS}" string inside of it, it will be
// automatically updated with the application's current frame rate (frames per
// second):
//
//  // The title would look like: "My App! 60FPS".
//  props.SetTitle("My App! {FPS}")
//
// Multiple Windows
//
// The package allows creation of multiple windows on platforms that support it
// (Windows, Linux, OS X, and the web via seperate HTML elements). Only one
// window can be created on mobile devices such as Android.
//
// New can create a new window from any goroutine, for instance in the above
// example:
//
//  // gfxLoop is our graphics loop, which runs in a separate goroutine.
//  func gfxLoop(w window.Window, r gfx.Renderer) {
//      // Create a second window!
//      w2, r2, err := window.New(nil)
//      if err != nil {
//          log.Fatal(err)
//      }
//
//      // Spawn the same graphics loop, but rendering to the second window!
//      go gfxLoop(w2, r2)
//
//      // Enter our render loop:
//      for {
//          // TODO: clear and draw to the canvas, r.
//          r.Render()
//      }
//  }
//
//  func main() {
//      // Create a window and run our graphics loop.
//      window.Run(gfxLoop, nil)
//  }
//
// If you prefer not to use the simple Run function, you can use the New and
// MainLoop functions yourself. The only requirement is that New cannot
// complete unless MainLoop is already running.
//
// The following code works:
//
//  func main() {
//      go func() {
//          // New runs in a seperate goroutine, after MainLoop has started.
//          w, r, err := window.New(nil)
//          ... use w, r, handle err ...
//      }()
//      window.MainLoop()
//  }
//
// The following code does not work (a deadlock occurs):
//
//  func main() {
//      // Won't ever complete: the main loop isn't running yet!
//      w, r, err := window.New(nil)
//      ... use w, r, handle err ...
//      window.MainLoop()
//  }
//
// Main Thread
//
// The MainLoop function internally locks the OS thread for you. In simple
// terms the only requirement is that you call either MainLoop, or Run which
// calls MainLoop for you, from your programs main function.
//
// This is required because some GUI libraries (such as Cocoa on OS X) require
// that they be run on the main operating system thread; not just the same
// operating system thread, but actually the main one.
//
// The MainLoop function takes control of the main loop for you, which is
// helpful. But in some situations you might want control of it (e.g. to call
// certain GUI libraries).
//
// To send an arbitrary function to execute on the main loop, simply send it
// over the main loop channel exposed by this package:
//
//  window.MainLoopChan <- func() {
//      fmt.Println("On the main thread!")
//  }
//
// More complex situations can be handled as well, by implementing the (small)
// MainLoop function yourself.
//
// Because a channel is used, the main loop is said to be communicative rather
// than employing a busy-waiting scheme.
//
// Examples
//
// The examples subdirectory of the gfx package contains several examples
// which utilize this package. Please see:
//
// https://azul3d.org/gfx.v2#hdr-Examples
//
package window
