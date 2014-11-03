// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import "testing"

var meshStateTests = []struct {
	name              string
	indices           []uint32
	vertices, normals []Vec3
	texCoords         []TexCoordSet
	attribs           map[string]VertexAttrib
	MeshState
}{
	{
		name:     "standard",
		indices:  make([]uint32, 3),
		vertices: make([]Vec3, 3),
		normals:  make([]Vec3, 3),
		MeshState: MeshState{
			Indices:  true,
			Vertices: true,
			Normals:  true,
		},
	}, {
		name: "texture coordinates",
		texCoords: []TexCoordSet{
			TexCoordSet{Slice: make([]TexCoord, 3)},
			TexCoordSet{},
			TexCoordSet{Slice: make([]TexCoord, 3)},
		},
		MeshState: MeshState{
			TexCoords: []bool{true, false, true},
		},
	}, {
		name: "attribs",
		attribs: map[string]VertexAttrib{
			"indices":  {Data: make([]uint32, 3)},
			"vertices": {Data: make([]Vec3, 3)},
			"normals":  {Data: make([]Vec3, 3)},
		},
		MeshState: MeshState{
			Attribs: map[string]bool{
				"indices":  true,
				"vertices": true,
				"normals":  true,
			},
		},
	},
}

// Tests querying of a mesh state from an actual mesh.
func TestMeshState(t *testing.T) {
	for _, tst := range meshStateTests {
		// Create mesh.
		m := NewMesh()
		m.Indices = tst.indices
		m.Vertices = tst.vertices
		m.Normals = tst.normals
		m.TexCoords = tst.texCoords
		m.Attribs = tst.attribs

		// Query state.
		s := new(MeshState)
		m.State(s)

		// The states should be equal at this point, verify it.
		if !s.Equals(&tst.MeshState) {
			t.Log(tst.name, "expected equal states")
			t.Logf("1: %+v\n", s)
			t.Fatalf("2: %+v\n", &tst.MeshState)
		}
	}
}

var meshStateEqualTests = []struct {
	name   string
	a, b   *MeshState
	expect bool
}{
	{
		name: "standard",
		a: &MeshState{
			Indices: true,
			Colors:  true,
		},
		b: &MeshState{
			Indices: true,
			Colors:  true,
		},
		expect: true,
	}, {
		name: "standard (unequal)",
		a: &MeshState{
			Indices: true,
			Colors:  true,
		},
		b: &MeshState{
			Indices: true,
			Colors:  false,
		},
		expect: false,
	}, {
		name: "texture coordinates",
		a: &MeshState{
			TexCoords: []bool{true, false, true},
			Colors:    true,
		},
		b: &MeshState{
			TexCoords: []bool{true, false, true},
			Colors:    true,
		},
		expect: true,
	}, {
		name: "texture coordinates (unequal lengths)",
		a: &MeshState{
			TexCoords: []bool{true, false, true},
		},
		b: &MeshState{
			TexCoords: []bool{true, false},
		},
		expect: false,
	}, {
		name: "attribs",
		a: &MeshState{
			Attribs: map[string]bool{"a": true, "b": false, "c": true},
			Colors:  true,
		},
		b: &MeshState{
			Attribs: map[string]bool{"a": true, "b": false, "c": true},
			Colors:  true,
		},
		expect: true,
	}, {
		name: "attribs (unequal lengths)",
		a: &MeshState{
			Attribs: map[string]bool{"a": true, "b": false, "c": true},
		},
		b: &MeshState{
			Attribs: map[string]bool{"a": true, "b": false},
		},
		expect: false,
	},
}

// Tests the ability to compare two mesh states for equality.
func TestMeshStateEquals(t *testing.T) {
	for _, tst := range meshStateEqualTests {
		got := tst.a.Equals(tst.b)
		if got != tst.expect {
			t.Log(tst.name)
			t.Logf("State A: %+v\n", tst.a)
			t.Logf("State B: %+v\n", tst.b)
			t.Fatalf("expected equal=%t, got equal=%t\n", tst.expect, got)
		}
	}
}

var meshStateDiffTests = []struct {
	name         string
	anyDiff      bool
	a, b, expect *MeshState
}{
	{
		name:    "no diff",
		anyDiff: false,
		a: &MeshState{
			Indices: true,
			Colors:  true,
		},
		b: &MeshState{
			Indices: true,
			Colors:  true,
		},
		expect: &MeshState{
			Indices: false,
			Colors:  false,
		},
	}, {
		name:    "colors diff",
		anyDiff: true,
		a: &MeshState{
			Indices: true,
			Colors:  false,
		},
		b: &MeshState{
			Indices: true,
			Colors:  true,
		},
		expect: &MeshState{
			Indices: false,
			Colors:  true,
		},
	},
}

// Tests the ability to find the difference of two mesh states.
func TestMeshStateDiff(t *testing.T) {
	diff := new(MeshState)
	for _, tst := range meshStateDiffTests {
		anyDiff := diff.Diff(tst.a, tst.b)
		if anyDiff != tst.anyDiff {
			t.Log(tst.name)
			t.Logf("State A:  %+v\n", tst.a)
			t.Logf("State B:  %+v\n", tst.b)
			t.Logf("Diff:     %+v\n", diff)
			t.Fatalf("Got anyDiff=%t, want anyDiff=%t\n", anyDiff, tst.anyDiff)
		}
		if !diff.Equals(tst.expect) {
			t.Log(tst.name)
			t.Logf("State A:  %+v\n", tst.a)
			t.Logf("State B:  %+v\n", tst.b)
			t.Logf("Got:      %+v\n", diff)
			t.Fatalf("Expected: %+v\n", tst.expect)
		}
	}
}

// Benchmarks the cost of querying a mesh's state.
func BenchmarkMeshState(b *testing.B) {
	var (
		a     = NewMesh()
		state = new(MeshState)
	)
	for n := 0; n < b.N; n++ {
		a.State(state)
	}
}

// Benchmarks the cost of testing of two mesh states are equal.
func BenchmarkMeshStateEqual(b *testing.B) {
	var (
		a      = NewMesh()
		aState = new(MeshState)
		bState = new(MeshState)
	)

	// Load the mesh's state into aState and bState.
	a.RLock()
	a.State(aState)
	a.State(bState)
	a.RUnlock()

	for n := 0; n < b.N; n++ {
		aState.Equals(bState)
	}
}

// Benchmarks the cost of finding the difference between two mesh states.
func BenchmarkMeshStateDiff(b *testing.B) {
	var (
		a         = NewMesh()
		diffState = new(MeshState)
		aState    = new(MeshState)
		bState    = new(MeshState)
		anyDiff   bool
	)

	// Load the mesh's state into aState and bState.
	a.RLock()
	a.State(aState)
	a.State(bState)
	a.RUnlock()

	for n := 0; n < b.N; n++ {
		anyDiff = diffState.Diff(aState, bState)
	}
	_ = anyDiff
}
