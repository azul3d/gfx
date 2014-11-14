// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build 386 amd64

package window

import (
	"fmt"
	"image"
	"log"
	"math"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"azul3d.org/gfx.v1"
	"azul3d.org/gfx/gl2.v2"
	"azul3d.org/keyboard.v1"
	"azul3d.org/mouse.v1"
	glfw "azul3d.org/native/glfw.v3.1"
)

// TODO(slimsag): rebuild window when fullscreen/precision changes.

// intBool returns 0 or 1 depending on b.
func intBool(b bool) int {
	if b {
		return 1
	}
	return 0
}

type notifier struct {
	EventMask
	ch chan<- Event
}

// glfwWindow implements the Window interface using a GLFW backend.
type glfwWindow struct {
	sync.RWMutex
	props, last                                        *Props
	mouse                                              *mouse.Watcher
	keyboard                                           *keyboard.Watcher
	renderer                                           *gl2.Renderer
	window, assets                                     *glfw.Window
	monitor                                            *glfw.Monitor
	shutdown                                           chan bool
	runOnMain                                          chan func()
	lastCursorX, lastCursorY                           float64
	extWGLEXTSwapControlTear, extGLXEXTSwapControlTear bool
	notifiers                                          []notifier
	closed                                             bool
}

// Implements the Window interface.
func (w *glfwWindow) Props() *Props {
	w.RLock()
	props := w.props
	w.RUnlock()
	return props
}

// Implements the Window interface.
func (w *glfwWindow) Request(p *Props) {
	w.runOnMain <- func() {
		w.useProps(p, false)
	}
}

// Implements the Window interface.
func (w *glfwWindow) Keyboard() *keyboard.Watcher {
	w.RLock()
	keyboard := w.keyboard
	w.RUnlock()
	return keyboard
}

// Implements the Window interface.
func (w *glfwWindow) Mouse() *mouse.Watcher {
	w.RLock()
	mouse := w.mouse
	w.RUnlock()
	return mouse
}

// Implements the Window interface.
func (w *glfwWindow) SetClipboard(clipboard string) {
	w.runOnMain <- func() {
		w.Lock()
		w.window.SetClipboardString(clipboard)
		w.Unlock()
	}
}

// Implements the Window interface.
func (w *glfwWindow) Clipboard() string {
	w.RLock()
	var str string
	w.waitFor(func() {
		str, _ = w.window.GetClipboardString()
	})
	w.RUnlock()
	return str
}

// Implements the Window interface.
func (w *glfwWindow) Close() {
	w.shutdown <- true
}

// Implements the Window interface.
func (w *glfwWindow) Notify(ch chan<- Event, m EventMask) {
	w.Lock()
	if m == NoEvents {
		w.deleteNotifiers(ch)
	} else {
		w.notifiers = append(w.notifiers, notifier{m, ch})
	}
	w.Unlock()
}

// searches for the event notifier associated with ch, returns it's slice index
// or -1.
//
// w.Lock must be held for it to operate safely.
func (w *glfwWindow) findNotifier(ch chan<- Event) int {
	for index, ev := range w.notifiers {
		if ev.ch == ch {
			return index
		}
	}
	return -1
}

// deletes all notifiers associated with ch.
func (w *glfwWindow) deleteNotifiers(ch chan<- Event) {
	s := w.notifiers
	idx := w.findNotifier(ch)
	for idx != -1 {
		s = append(s[:idx], s[idx+1:]...)
		idx = w.findNotifier(ch)
	}
	w.notifiers = s
}

// waitFor runs f on the main thread and waits for the function to complete.
func (w *glfwWindow) waitFor(f func()) {
	done := make(chan bool, 1)
	w.runOnMain <- func() {
		f()
		done <- true
	}
	<-done
}

// Update window title and accounts for "{FPS}" strings.
func (w *glfwWindow) updateTitle() {
	fps := fmt.Sprintf("%dFPS", int(math.Ceil(w.renderer.Clock().FrameRate())))
	title := strings.Replace(w.props.Title(), "{FPS}", fps, 1)
	w.window.SetTitle(title)
}

func (w *glfwWindow) useProps(p *Props, force bool) {
	w.Lock()
	defer w.Unlock()
	w.props = p

	// Runs f without the currently held lock. Because some functions cause an
	// event to be generated, calling the event callback and causing a deadlock.
	withoutLock := func(f func()) {
		w.Unlock()
		f()
		w.Lock()
	}
	win := w.window

	// Set each property, only if it differs from the last known value for that
	// property.

	w.updateTitle()

	// Window Size.
	width, height := w.props.Size()
	lastWidth, lastHeight := w.last.Size()
	if force || width != lastWidth || height != lastHeight {
		w.last.SetSize(width, height)
		withoutLock(func() {
			win.SetSize(width, height)
		})
	}

	// Window Position.
	x, y := w.props.Pos()
	lastX, lastY := w.last.Pos()
	if force || x != lastX || y != lastY {
		w.last.SetPos(x, y)
		if x == -1 && y == -1 {
			vm, err := w.monitor.GetVideoMode()
			if err == nil {
				x = (vm.Width / 2) - (width / 2)
				y = (vm.Height / 2) - (height / 2)
			}
		}
		withoutLock(func() {
			win.SetPosition(x, y)
		})
	}

	// Cursor Position.
	cursorX, cursorY := w.props.CursorPos()
	lastCursorX, lastCursorY := w.last.CursorPos()
	if force || cursorX != lastCursorX || cursorY != lastCursorY {
		w.last.SetCursorPos(cursorX, cursorY)
		if cursorX != -1 && cursorY != -1 {
			withoutLock(func() {
				win.SetCursorPosition(cursorX, cursorY)
			})
		}
	}

	// Window Visibility.
	visible := w.props.Visible()
	if force || w.last.Visible() != visible {
		w.last.SetVisible(visible)
		withoutLock(func() {
			if visible {
				win.Show()
			} else {
				win.Hide()
			}
		})
	}

	// Window Minimized.
	minimized := w.props.Minimized()
	if force || w.last.Minimized() != minimized {
		w.last.SetMinimized(minimized)
		withoutLock(func() {
			if minimized {
				win.Iconify()
			} else {
				win.Restore()
			}
		})
	}

	// Vertical sync mode.
	vsync := w.props.VSync()
	if force || w.last.VSync() != vsync {
		w.last.SetVSync(vsync)

		// Determine the swap interval and set it.
		var swapInterval int
		if vsync {
			// We want vsync on, we will use adaptive vsync if we have it, if
			// not we will use standard vsync.
			if w.extWGLEXTSwapControlTear || w.extGLXEXTSwapControlTear {
				// We can use adaptive vsync via a swap interval of -1.
				swapInterval = -1
			} else {
				// No adaptive vsync, use standard then.
				swapInterval = 1
			}
		}
		glfw.SwapInterval(swapInterval)
	}

	// The following cannot be changed via GLFW post window creation -- and
	// they are not deemed significant enough to warrant rebuilding the window.
	//
	// TODO(slimsag): consider these when rebuilding the window for Fullscreen
	// or Precision switches.
	//
	//  Focused
	//  Resizable
	//  Decorated
	//  AlwaysOnTop (via GLFW_FLOATING)

	// Cursor Mode.
	grabbed := w.props.CursorGrabbed()
	if force || w.last.CursorGrabbed() != grabbed {
		w.last.SetCursorGrabbed(grabbed)

		// Reset both last cursor values to the callback can identify the
		// large/fake delta.
		w.lastCursorX = math.Inf(-1)
		w.lastCursorY = math.Inf(-1)

		// Set input mode.
		withoutLock(func() {
			if grabbed {
				w.window.SetInputMode(glfw.Cursor, glfw.CursorDisabled)
			} else {
				w.window.SetInputMode(glfw.Cursor, glfw.CursorNormal)
			}
		})
	}
}

func (w *glfwWindow) sendEvent(ev Event, m EventMask) {
	w.RLock()
	for _, nf := range w.notifiers {
		if (nf.EventMask & m) != 0 {
			select {
			case nf.ch <- ev:
			default:
			}
		}
	}
	w.RUnlock()
}

// initCallbacks sets a callback handler for each GLFW window event.
func (w *glfwWindow) initCallbacks() {
	// Close event.
	w.window.SetCloseCallback(func(gw *glfw.Window) {
		// If they want us to close the window, then close the window.
		if w.Props().ShouldClose() {
			w.Close()

			// Return so we don't give people the idea that they can rely on
			// Close event below to cleanup things.
			return
		}
		w.sendEvent(Close{T: time.Now()}, CloseEvents)
	})

	// Damaged event.
	w.window.SetRefreshCallback(func(gw *glfw.Window) {
		w.sendEvent(Damaged{T: time.Now()}, DamagedEvents)
	})

	// Minimized and Restored events.
	w.window.SetIconifyCallback(func(gw *glfw.Window, iconify bool) {
		// Store the minimized/restored state.
		w.RLock()
		w.last.SetMinimized(iconify)
		w.props.SetMinimized(iconify)
		w.RUnlock()

		// Send the proper event.
		if iconify {
			w.sendEvent(Minimized{T: time.Now()}, MinimizedEvents)
			return
		}
		w.sendEvent(Restored{T: time.Now()}, RestoredEvents)
	})

	// FocusChanged event.
	w.window.SetFocusCallback(func(gw *glfw.Window, focused bool) {
		// Store the focused state.
		w.RLock()
		w.last.SetFocused(focused)
		w.props.SetFocused(focused)
		w.RUnlock()

		// Send the proper event.
		if focused {
			w.sendEvent(GainedFocus{T: time.Now()}, GainedFocusEvents)
			return
		}
		w.sendEvent(LostFocus{T: time.Now()}, LostFocusEvents)
	})

	// Moved event.
	w.window.SetPositionCallback(func(gw *glfw.Window, x, y int) {
		// Store the position state.
		w.RLock()
		w.last.SetPos(x, y)
		w.props.SetPos(x, y)
		w.RUnlock()
		w.sendEvent(Moved{X: x, Y: y, T: time.Now()}, MovedEvents)
	})

	// Resized event.
	w.window.SetSizeCallback(func(gw *glfw.Window, width, height int) {
		// Store the size state.
		w.RLock()
		w.last.SetSize(width, height)
		w.props.SetSize(width, height)
		w.RUnlock()
		w.sendEvent(Resized{
			Width:  width,
			Height: height,
			T:      time.Now(),
		}, ResizedEvents)
	})

	// FramebufferResized event.
	w.window.SetFramebufferSizeCallback(func(gw *glfw.Window, width, height int) {
		// Store the framebuffer size state.
		w.RLock()
		w.last.SetFramebufferSize(width, height)
		w.props.SetFramebufferSize(width, height)
		w.RUnlock()

		// Update renderer bounds.
		w.renderer.UpdateBounds(image.Rect(0, 0, width, height))

		// Send the event.
		w.sendEvent(FramebufferResized{
			Width:  width,
			Height: height,
			T:      time.Now(),
		}, FramebufferResizedEvents)
	})

	// Dropped event.
	w.window.SetDropCallback(func(gw *glfw.Window, items []string) {
		w.sendEvent(ItemsDropped{Items: items, T: time.Now()}, ItemsDroppedEvents)
	})

	// CursorMoved event.
	w.window.SetCursorPositionCallback(func(gw *glfw.Window, x, y float64) {
		// Store the cursor position state.
		w.RLock()
		grabbed := w.props.CursorGrabbed()
		if grabbed {
			// Store/swap last cursor values. Note: It's safe to modify
			// lastCursorX/Y with just w.RLock because they are only modified
			// in this callback on the main thread.
			lastX := w.lastCursorX
			lastY := w.lastCursorY
			w.lastCursorX = x
			w.lastCursorY = y

			// First cursor position callback since grab occured, avoid the
			// large/fake delta.
			if lastX == math.Inf(-1) && lastY == math.Inf(-1) {
				w.RUnlock()
				return
			}

			// Calculate cursor delta.
			x = x - lastX
			y = y - lastY
		} else {
			// Store cursor position.
			w.last.SetCursorPos(x, y)
			w.props.SetCursorPos(x, y)
		}
		w.RUnlock()

		// Send proper event.
		w.sendEvent(CursorMoved{
			X:     x,
			Y:     y,
			Delta: grabbed,
			T:     time.Now(),
		}, CursorMovedEvents)
	})

	// CursorEnter and CursorExit events.
	w.window.SetCursorEnterCallback(func(gw *glfw.Window, enter bool) {
		// TODO(slimsag): expose *within window* state, but not via Props.
		if enter {
			w.sendEvent(CursorEnter{T: time.Now()}, CursorEnterEvents)
			return
		}
		w.sendEvent(CursorExit{T: time.Now()}, CursorExitEvents)
	})

	// keyboard.TypedEvent
	w.window.SetCharacterCallback(func(gw *glfw.Window, r rune) {
		w.sendEvent(keyboard.TypedEvent{Rune: r, T: time.Now()}, KeyboardTypedEvents)
	})

	// keyboard.StateEvent
	w.window.SetKeyCallback(func(gw *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Repeat {
			return
		}

		// Convert GLFW event.
		k := convertKey(key)
		s := convertKeyAction(action)
		r := uint64(scancode)

		// Update keyboard watcher.
		w.keyboard.SetState(k, s)
		w.keyboard.SetRawState(r, s)

		// Send the event.
		w.sendEvent(keyboard.StateEvent{
			T:     time.Now(),
			Key:   convertKey(key),
			State: convertKeyAction(action),
			Raw:   uint64(scancode),
		}, KeyboardStateEvents)
	})

	// mouse.Event
	w.window.SetMouseButtonCallback(func(gw *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
		// Convert GLFW event.
		b := convertMouseButton(button)
		s := convertMouseAction(action)

		// Update mouse watcher.
		w.mouse.SetState(b, s)

		// Send the event.
		w.sendEvent(mouse.Event{
			T:      time.Now(),
			Button: b,
			State:  s,
		}, MouseEvents)
	})

	// mouse.Scrolled event.
	w.window.SetScrollCallback(func(gw *glfw.Window, x, y float64) {
		w.sendEvent(mouse.Scrolled{
			T: time.Now(),
			X: x,
			Y: y,
		}, MouseScrolledEvents)
	})
}

func doRun(gfxLoop func(w Window, r gfx.Renderer), p *Props) {
	// Initialize GLFW, and later on invoke Terminate.
	err := glfw.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer glfw.Terminate()

	// Create the asset context by creating a hidden window.
	glfw.WindowHint(glfw.Visible, 0)
	assets, err := glfw.CreateWindow(32, 32, "assets", nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Specify the primary monitor if we want fullscreen, store the monitor
	// regardless for centering the window.
	var targetMonitor, monitor *glfw.Monitor
	monitor, err = glfw.GetPrimaryMonitor()
	if err != nil {
		log.Fatal(err)
	}
	if p.Fullscreen() {
		targetMonitor = monitor
	}

	// Hint standard properties (note visibility is always false, we show the
	// window later after moving it).
	glfw.WindowHint(glfw.Visible, 0)
	//glfw.WindowHint(glfw.Focused, intBool(p.Focused()))
	//glfw.WindowHint(glfw.Iconified, intBool(p.Minimized()))
	glfw.WindowHint(glfw.Resizable, intBool(p.Resizable()))
	glfw.WindowHint(glfw.Decorated, intBool(p.Decorated()))
	glfw.WindowHint(glfw.AutoIconify, 1)
	glfw.WindowHint(glfw.Floating, intBool(p.AlwaysOnTop()))

	// Hint context properties.
	prec := p.Precision()
	glfw.WindowHint(glfw.RedBits, int(prec.RedBits))
	glfw.WindowHint(glfw.GreenBits, int(prec.GreenBits))
	glfw.WindowHint(glfw.BlueBits, int(prec.BlueBits))
	glfw.WindowHint(glfw.AlphaBits, int(prec.AlphaBits))
	glfw.WindowHint(glfw.DepthBits, int(prec.DepthBits))
	glfw.WindowHint(glfw.StencilBits, int(prec.StencilBits))
	glfw.WindowHint(glfw.Samples, prec.Samples)
	glfw.WindowHint(glfw.SRGBCapable, 1)

	// Create the render window.
	width, height := p.Size()
	window, err := glfw.CreateWindow(width, height, p.Title(), targetMonitor, assets)
	if err != nil {
		log.Fatal(err)
	}

	// OpenGL rendering context must be active.
	window.MakeContextCurrent()
	defer glfw.DetachCurrentContext()

	// Create the renderer.
	r, err := gl2.New(false)
	if err != nil {
		log.Fatal(err)
	}

	// Write renderer debug output (shader errors, etc) to stdout.
	r.SetDebugOutput(os.Stdout)

	// Channel to signal shutdown to the loader.
	loaderShutdown := make(chan bool, 1)

	// Spawn a goroutine to manage loading of resources.
	go func() {
		// All OpenGL related calls must occur in the same OS thread.
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		// OpenGL loading context must be active.
		assets.MakeContextCurrent()
		defer glfw.DetachCurrentContext()

		// Execute loader functions until shutdown.
		for {
			select {
			case <-loaderShutdown:
				return
			case fn := <-r.LoaderExec:
				fn()
			}
		}
	}()

	// Initialize window.
	w := &glfwWindow{
		props:     p,
		last:      NewProps(),
		mouse:     mouse.NewWatcher(),
		keyboard:  keyboard.NewWatcher(),
		renderer:  r,
		window:    window,
		assets:    assets,
		monitor:   monitor,
		shutdown:  make(chan bool, 1),
		runOnMain: make(chan func(), 32),
	}

	// Test for adaptive vsync extensions.
	w.extWGLEXTSwapControlTear = glfw.ExtensionSupported("WGL_EXT_swap_control_tear")
	w.extGLXEXTSwapControlTear = glfw.ExtensionSupported("GLX_EXT_swap_control_tear")

	// Setup callbacks and the window.
	w.initCallbacks()
	w.useProps(p, true)

	// Start the user-controlled graphics loop. If the graphics loop causes a
	// panic to occur we want to be sure to recover it and properly terminate
	// GLFW first -- since it e.g. restores the monitor resolution/gamma/etc.
	go func() {
		defer func() {
			if r := recover(); r != nil {
				glfw.Terminate()
				panic(r)
			}
		}()
		gfxLoop(w, r)
	}()

	// Enter the (main) rendering loop.
	updateFPS := time.Tick(1 * time.Second)
	for {
		select {
		case <-updateFPS:
			// Update title with FPS.
			w.Lock()
			w.updateTitle()
			w.Unlock()

		case <-w.shutdown:
			loaderShutdown <- true
			w.window.Destroy()
			return

		case fn := <-w.runOnMain:
			fn()

		case fn := <-r.RenderExec:
			if renderedFrame := fn(); renderedFrame {
				// Swap OpenGL buffers.
				window.SwapBuffers()

				// Poll for events.
				glfw.PollEvents()
			}
		}
	}
}
