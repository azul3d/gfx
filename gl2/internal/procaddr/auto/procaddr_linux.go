// +build !egl,!noauto

package auto

import "azul3d.org/gfx.v2/gl2/internal/procaddr/glx"

func init() {
	GetProcAddress = glx.GetProcAddress
}
