// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// +build 386 amd64

package window

import (
	"azul3d.org/keyboard.v1"
	"azul3d.org/mouse.v1"

	glfw "azul3d.org/native/glfw.v3"
)

func convertMouseAction(a glfw.Action) mouse.State {
	switch a {
	case glfw.Press:
		return mouse.Down
	case glfw.Release:
		return mouse.Up
	default:
		panic("invalid key action")
	}
}

func convertKeyAction(a glfw.Action) keyboard.State {
	switch a {
	case glfw.Press:
		return keyboard.Down
	case glfw.Release:
		return keyboard.Up
	default:
		panic("invalid key action")
	}
}

func convertMouseButton(b glfw.MouseButton) mouse.Button {
	switch b {
	case glfw.MouseButton1:
		return mouse.One
	case glfw.MouseButton2:
		return mouse.Two
	case glfw.MouseButton3:
		return mouse.Three
	case glfw.MouseButton4:
		return mouse.Four
	case glfw.MouseButton5:
		return mouse.Five
	case glfw.MouseButton6:
		return mouse.Six
	case glfw.MouseButton7:
		return mouse.Seven
	case glfw.MouseButton8:
		return mouse.Eight
	default:
		panic("unhandled mouse button")
	}
}

func convertKey(k glfw.Key) keyboard.Key {
	switch k {
	// 0-10 keys.
	case glfw.Key0:
		return keyboard.Zero
	case glfw.Key1:
		return keyboard.One
	case glfw.Key2:
		return keyboard.Two
	case glfw.Key3:
		return keyboard.Three
	case glfw.Key4:
		return keyboard.Four
	case glfw.Key5:
		return keyboard.Five
	case glfw.Key6:
		return keyboard.Six
	case glfw.Key7:
		return keyboard.Seven
	case glfw.Key8:
		return keyboard.Eight
	case glfw.Key9:
		return keyboard.Nine

	// A-Z keys.
	case glfw.KeyA:
		return keyboard.A
	case glfw.KeyB:
		return keyboard.B
	case glfw.KeyC:
		return keyboard.C
	case glfw.KeyD:
		return keyboard.D
	case glfw.KeyE:
		return keyboard.E
	case glfw.KeyF:
		return keyboard.F
	case glfw.KeyG:
		return keyboard.G
	case glfw.KeyH:
		return keyboard.H
	case glfw.KeyI:
		return keyboard.I
	case glfw.KeyJ:
		return keyboard.J
	case glfw.KeyK:
		return keyboard.K
	case glfw.KeyL:
		return keyboard.L
	case glfw.KeyM:
		return keyboard.M
	case glfw.KeyN:
		return keyboard.N
	case glfw.KeyO:
		return keyboard.O
	case glfw.KeyP:
		return keyboard.P
	case glfw.KeyQ:
		return keyboard.Q
	case glfw.KeyR:
		return keyboard.R
	case glfw.KeyS:
		return keyboard.S
	case glfw.KeyT:
		return keyboard.T
	case glfw.KeyU:
		return keyboard.U
	case glfw.KeyV:
		return keyboard.V
	case glfw.KeyW:
		return keyboard.W
	case glfw.KeyX:
		return keyboard.X
	case glfw.KeyY:
		return keyboard.Y
	case glfw.KeyZ:
		return keyboard.Z

	// F1-F25 keys.
	case glfw.KeyF1:
		return keyboard.F1
	case glfw.KeyF2:
		return keyboard.F2
	case glfw.KeyF3:
		return keyboard.F3
	case glfw.KeyF4:
		return keyboard.F4
	case glfw.KeyF5:
		return keyboard.F5
	case glfw.KeyF6:
		return keyboard.F6
	case glfw.KeyF7:
		return keyboard.F7
	case glfw.KeyF8:
		return keyboard.F8
	case glfw.KeyF9:
		return keyboard.F9
	case glfw.KeyF10:
		return keyboard.F10
	case glfw.KeyF11:
		return keyboard.F11
	case glfw.KeyF12:
		return keyboard.F12
	case glfw.KeyF13:
		return keyboard.F13
	case glfw.KeyF14:
		return keyboard.F14
	case glfw.KeyF15:
		return keyboard.F15
	case glfw.KeyF16:
		return keyboard.F16
	case glfw.KeyF17:
		return keyboard.F17
	case glfw.KeyF18:
		return keyboard.F18
	case glfw.KeyF19:
		return keyboard.F19
	case glfw.KeyF20:
		return keyboard.F20
	case glfw.KeyF21:
		return keyboard.F21
	case glfw.KeyF22:
		return keyboard.F22
	case glfw.KeyF23:
		return keyboard.F23
	case glfw.KeyF24:
		return keyboard.F24
	case glfw.KeyF25:
		return keyboard.F25

	// Numpad keys.
	case glfw.KeyKp0:
		return keyboard.NumZero
	case glfw.KeyKp1:
		return keyboard.NumOne
	case glfw.KeyKp2:
		return keyboard.NumTwo
	case glfw.KeyKp3:
		return keyboard.NumThree
	case glfw.KeyKp4:
		return keyboard.NumFour
	case glfw.KeyKp5:
		return keyboard.NumFive
	case glfw.KeyKp6:
		return keyboard.NumSix
	case glfw.KeyKp7:
		return keyboard.NumSeven
	case glfw.KeyKp8:
		return keyboard.NumEight
	case glfw.KeyKp9:
		return keyboard.NumNine
	case glfw.KeyKpDecimal:
		return keyboard.NumDecimal
	case glfw.KeyKpDivide:
		return keyboard.NumDivide
	case glfw.KeyKpMultiply:
		return keyboard.NumMultiply
	case glfw.KeyKpSubtract:
		return keyboard.NumSubtract
	case glfw.KeyKpAdd:
		return keyboard.NumAdd
	case glfw.KeyKpEnter:
		return keyboard.NumEnter
	case glfw.KeyNumLock:
		return keyboard.NumLock

	// Lefties.
	case glfw.KeyLeftBracket:
		return keyboard.LeftBracket
	case glfw.KeyLeftShift:
		return keyboard.LeftShift
	case glfw.KeyLeftControl:
		return keyboard.LeftCtrl
	case glfw.KeyLeftAlt:
		return keyboard.LeftAlt
	case glfw.KeyLeftSuper:
		return keyboard.LeftSuper

	// Righties.
	case glfw.KeyRightBracket:
		return keyboard.RightBracket
	case glfw.KeyRightShift:
		return keyboard.RightShift
	case glfw.KeyRightControl:
		return keyboard.RightCtrl
	case glfw.KeyRightAlt:
		return keyboard.RightAlt
	case glfw.KeyRightSuper:
		return keyboard.RightSuper

	// Arrow keys.
	case glfw.KeyLeft:
		return keyboard.ArrowLeft
	case glfw.KeyRight:
		return keyboard.ArrowRight
	case glfw.KeyDown:
		return keyboard.ArrowDown
	case glfw.KeyUp:
		return keyboard.ArrowUp

	// General keys.
	case glfw.KeyUnknown:
		return keyboard.Invalid
	case glfw.KeySpace:
		return keyboard.Space
	case glfw.KeyApostrophe:
		return keyboard.Apostrophe
	case glfw.KeyComma:
		return keyboard.Comma
	case glfw.KeyMinus:
		return keyboard.Dash
	case glfw.KeyPeriod:
		return keyboard.Period
	case glfw.KeySlash:
		return keyboard.ForwardSlash
	case glfw.KeyBackslash:
		return keyboard.BackSlash
	case glfw.KeySemicolon:
		return keyboard.Semicolon
	case glfw.KeyEqual:
		return keyboard.Equals
	case glfw.KeyEscape:
		return keyboard.Escape
	case glfw.KeyEnter:
		return keyboard.Enter
	case glfw.KeyTab:
		return keyboard.Tab
	case glfw.KeyBackspace:
		return keyboard.Backspace
	case glfw.KeyInsert:
		return keyboard.Insert
	case glfw.KeyDelete:
		return keyboard.Delete
	case glfw.KeyPageUp:
		return keyboard.PageUp
	case glfw.KeyPageDown:
		return keyboard.PageDown
	case glfw.KeyHome:
		return keyboard.Home
	case glfw.KeyEnd:
		return keyboard.End
	case glfw.KeyCapsLock:
		return keyboard.CapsLock
	case glfw.KeyScrollLock:
		return keyboard.ScrollLock
	case glfw.KeyPrintScreen:
		return keyboard.PrintScreen
	case glfw.KeyPause:
		return keyboard.Pause
	case glfw.KeyMenu:
		return keyboard.Applications
	case glfw.KeyGraveAccent:
		return keyboard.Tilde

	// TODO(slimsag): find the proper values for these.
	case glfw.KeyWorld1:
		return keyboard.Invalid
	case glfw.KeyWorld2:
		return keyboard.Invalid
	case glfw.KeyKpEqual:
		return keyboard.Invalid

	default:
		panic("unhandled key")
	}
}
