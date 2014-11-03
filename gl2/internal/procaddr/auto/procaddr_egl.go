// +build egl,!noauto

package auto

import "azul3d.org/gfx/gl2.v2/internal/procaddr/egl"

func init() {
	GetProcAddress = egl.GetProcAddress
}
