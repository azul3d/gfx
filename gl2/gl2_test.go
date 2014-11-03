// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"image"
	"math/rand"
	"testing"

	"azul3d.org/gfx.v1"
)

func TestRendererInterface(t *testing.T) {
	var r *Renderer
	_ = gfx.Renderer(r)
}

func TestRectConversion(t *testing.T) {
	bounds := image.Rect(0, 0, 100, 100)
	s := image.Rect(20, 20, 50, 50)
	x, y, w, h := convertRect(s, bounds)
	if x != 20 || y != 50 {
		t.Log(x, y, w, h)
		t.Fail()
	}
	if w != 30 || h != 30 {
		t.Log(x, y, w, h)
		t.Fail()
	}

	tmp := unconvertRect(bounds, x, y, int32(w), int32(h))
	if tmp != s {
		t.Log("got", tmp)
		t.Log("want", s)
		t.Fail()
	}
}

func TestRectConversionDual(t *testing.T) {
	rand.Seed(3000)
	bounds := image.Rect(0, 0, 100, 100)
	for i := 0; i < 50; i++ {
		x0 := rand.Intn(600)
		y0 := rand.Intn(400)
		x1 := x0 + rand.Intn(40)
		y1 := y0 + rand.Intn(80)
		s := image.Rect(x0, y0, x1, y1)
		x, y, w, h := convertRect(s, bounds)
		tmp := unconvertRect(bounds, x, y, int32(w), int32(h))
		if tmp != s {
			t.Log("got", tmp)
			t.Log("want", s)
			t.Fail()
			return
		}
	}
}
