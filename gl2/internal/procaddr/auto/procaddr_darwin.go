// +build !egl,!noauto

package auto

import "azul3d.org/gfx.v2/gl2/internal/procaddr/darwin"

func init() {
	GetProcAddress = darwin.GetProcAddress
}
