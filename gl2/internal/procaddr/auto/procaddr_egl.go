// +build egl,!noauto

package auto

import "azul3d.org/gfx.v2/gl2/internal/procaddr/egl"

func init() {
	GetProcAddress = egl.GetProcAddress
}
