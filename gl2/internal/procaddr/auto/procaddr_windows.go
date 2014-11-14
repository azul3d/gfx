// +build !egl,!noauto

package auto

import "azul3d.org/gfx/gl2.v2/internal/procaddr/wgl"

func init() {
	GetProcAddress = wgl.GetProcAddress
}
