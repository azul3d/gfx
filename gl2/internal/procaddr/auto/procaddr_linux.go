// +build !egl,!noauto

package auto

import "azul3d.org/gfx/gl2.v2/internal/procaddr/glx"

func init() {
	GetProcAddress = glx.GetProcAddress
}
