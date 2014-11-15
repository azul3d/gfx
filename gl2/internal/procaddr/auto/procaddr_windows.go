// +build !egl,!noauto

package auto

import "azul3d.org/gfx.v2/gl2/internal/procaddr/wgl"

func init() {
	GetProcAddress = wgl.GetProcAddress
}
