// Package auto automatically selects an appropriate GetProcAddressFunc based
// on the build environment. The default configuration is:
// - Windows -> WGL
// - OS X (Darwin) -> CGL
// - Linux -> GLX
// To control the selection use build tags:
// - egl: Use EGL
// - noauto: Disable automatic selection
package auto

import "azul3d.org/gfx/gl2.v2/internal/procaddr"

// Automatically selected GetProcAddressFunc.
var GetProcAddress procaddr.GetProcAddressFunc
