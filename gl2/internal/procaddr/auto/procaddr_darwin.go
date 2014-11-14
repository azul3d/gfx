// +build !egl,!noauto

package auto

import "azul3d.org/gfx/gl2.v2/internal/procaddr/darwin"

func init() {
	GetProcAddress = darwin.GetProcAddress
}
