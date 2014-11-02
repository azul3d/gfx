// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import "testing"

var meshAppendTests = []struct {
	name                                           string
	a, b, want                                     []Vec3
	aIndices, bIndices, wantIndices                []uint32
	aColors, bColors, wantColors                   []Color
	indicesChanged, verticesChanged, colorsChanged bool
}{
	{
		name:            "mesh.Append(mesh)",
		a:               []Vec3{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}},
		b:               []Vec3{{8, 7, 6}, {5, 4, 3}, {2, 1, 0}},
		want:            []Vec3{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, {8, 7, 6}, {5, 4, 3}, {2, 1, 0}},
		verticesChanged: true,
	}, {
		name:            "indexedMesh.Append(indexedMesh)",
		a:               []Vec3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}},
		b:               []Vec3{{3, 3, 3}, {4, 4, 4}, {5, 5, 5}},
		want:            []Vec3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}, {3, 3, 3}, {4, 4, 4}, {5, 5, 5}},
		aIndices:        []uint32{0, 1, 2},
		bIndices:        []uint32{2, 2, 1},
		wantIndices:     []uint32{0, 1, 2, 5, 5, 4},
		indicesChanged:  true,
		verticesChanged: true,
	}, {
		name:            "mesh.Append(indexedMesh)",
		a:               []Vec3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}},
		b:               []Vec3{{3, 3, 3}, {4, 4, 4}, {5, 5, 5}},
		want:            []Vec3{{0, 0, 0}, {1, 1, 1}, {2, 2, 2}, {5, 5, 5}, {4, 4, 4}, {3, 3, 3}},
		bIndices:        []uint32{2, 1, 0},
		verticesChanged: true,
	}, {
		name:            "indexedMesh.Append(mesh)",
		a:               []Vec3{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}},
		b:               []Vec3{{8, 7, 6}, {5, 4, 3}, {2, 1, 0}},
		want:            []Vec3{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, {8, 7, 6}, {5, 4, 3}, {2, 1, 0}},
		aIndices:        []uint32{2, 1, 0},
		wantIndices:     []uint32{2, 1, 0, 3, 4, 5},
		indicesChanged:  true,
		verticesChanged: true,
	}, {
		name:            "mesh.Append(mesh) // Colors Copy",
		a:               []Vec3{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}},
		b:               []Vec3{{8, 7, 6}, {5, 4, 3}, {2, 1, 0}},
		want:            []Vec3{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, {8, 7, 6}, {5, 4, 3}, {2, 1, 0}},
		aColors:         []Color{{R: 0}, {R: 1}, {R: 2}},
		bColors:         []Color{{R: 3}, {R: 4}, {R: 5}},
		wantColors:      []Color{{R: 0}, {R: 1}, {R: 2}, {R: 3}, {R: 4}, {R: 5}},
		verticesChanged: true,
		colorsChanged:   true,
	}, {
		name:            "mesh.Append(mesh) // Colors Loss A",
		a:               []Vec3{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}},
		b:               []Vec3{{8, 7, 6}, {5, 4, 3}, {2, 1, 0}},
		want:            []Vec3{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, {8, 7, 6}, {5, 4, 3}, {2, 1, 0}},
		bColors:         []Color{{R: 0}, {R: 1}, {R: 2}},
		verticesChanged: true,
		colorsChanged:   true,
	}, {
		name:            "mesh.Append(mesh) // Colors Loss B",
		a:               []Vec3{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}},
		b:               []Vec3{{8, 7, 6}, {5, 4, 3}, {2, 1, 0}},
		want:            []Vec3{{0, 1, 2}, {3, 4, 5}, {6, 7, 8}, {8, 7, 6}, {5, 4, 3}, {2, 1, 0}},
		aColors:         []Color{{R: 0}, {R: 1}, {R: 2}},
		verticesChanged: true,
		colorsChanged:   true,
	},
}

func TestMeshAppend(t *testing.T) {
	for caseNumber, tst := range meshAppendTests {
		// Create the meshes.
		a := NewMesh()
		a.Vertices = tst.a
		a.Indices = tst.aIndices
		a.Colors = tst.aColors
		a.Attribs = map[string]VertexAttrib{
			"Colors": {Data: tst.aColors},
		}

		b := NewMesh()
		b.Vertices = tst.b
		b.Indices = tst.bIndices
		b.Colors = tst.bColors
		b.Attribs = map[string]VertexAttrib{
			"Colors": {Data: tst.bColors},
		}

		// Append mesh b to mesh a.
		a.Append(b)

		// Validate the vertices slices.
		if a.VerticesChanged != tst.verticesChanged {
			t.Logf("Case %d: %s\n", caseNumber, tst.name)
			t.Log("Got VerticesChanged:", a.VerticesChanged)
			t.Fatal("Want VerticesChanged:", tst.verticesChanged)
		}
		if len(tst.want) != len(a.Vertices) {
			t.Logf("Case %d: %s\n", caseNumber, tst.name)
			t.Fatal("got", len(tst.want), "vertices, want", len(a.Vertices), "vertices")
		}
		for i, v := range tst.want {
			if a.Vertices[i] != v {
				t.Logf("Case %d: %s\n", caseNumber, tst.name)
				t.Log("got Vertices: ", a.Vertices)
				t.Fatal("want Vertices:", tst.want)
			}
		}

		// Validate the indices slices.
		if a.IndicesChanged != tst.indicesChanged {
			t.Logf("Case %d: %s\n", caseNumber, tst.name)
			t.Log("Got IndicesChanged:", a.IndicesChanged)
			t.Fatal("Want IndicesChanged:", tst.indicesChanged)
		}
		if len(tst.wantIndices) != len(a.Indices) {
			t.Logf("Case %d: %s\n", caseNumber, tst.name)
			t.Fatal("got", len(a.Indices), "indices, want", len(tst.wantIndices), "indices")
		}
		for i, v := range tst.wantIndices {
			if a.Indices[i] != v {
				t.Logf("Case %d: %s\n", caseNumber, tst.name)
				t.Log("got Indices: ", a.Indices)
				t.Fatal("want Indices:", tst.wantIndices)
			}
		}

		// Validate the Colors slices.
		if a.ColorsChanged != tst.colorsChanged {
			t.Logf("Case %d: %s\n", caseNumber, tst.name)
			t.Log("Got ColorsChanged:", a.ColorsChanged)
			t.Fatal("Want ColorsChanged:", tst.colorsChanged)
		}
		if len(tst.wantColors) != len(a.Colors) {
			t.Logf("Case %d: %s\n", caseNumber, tst.name)
			t.Fatal("got", len(a.Colors), "Colors, want", len(tst.wantColors), "Colors")
		}
		for i, v := range tst.wantColors {
			if a.Colors[i] != v {
				t.Logf("Case %d: %s\n", caseNumber, tst.name)
				t.Log("got Colors: ", a.Colors)
				t.Fatal("want Colors:", tst.wantColors)
			}
		}

		// Validate the custom per-vertex-attribute.
		var gotColors []Color
		va := a.Attribs["Colors"]
		if va.Data != nil {
			gotColors = va.Data.([]Color)
		}
		if va.Changed != tst.colorsChanged {
			t.Logf("Case %d: %s\n", caseNumber, tst.name)
			t.Log("Got Attrib Changed:", va.Changed)
			t.Fatal("Want Attrib Changed:", tst.colorsChanged)
		}
		if len(tst.wantColors) != len(gotColors) {
			t.Logf("Case %d: %s\n", caseNumber, tst.name)
			t.Fatal("got", len(gotColors), "Attrib Colors, want", len(tst.wantColors), "Attrib Colors")
		}
		for i, v := range tst.wantColors {
			if gotColors[i] != v {
				t.Logf("Case %d: %s\n", caseNumber, tst.name)
				t.Log("got Attrib Colors: ", gotColors)
				t.Fatal("want Attrib Colors:", tst.wantColors)
			}
		}
	}
}

func benchmarkMeshAppend(b *testing.B, n int, prealloc bool) {
	// Create the meshes.
	n /= 3
	a := NewMesh()
	a.Vertices = make([]Vec3, n)
	a.Colors = make([]Color, n)
	a.Attribs = map[string]VertexAttrib{
		"Colors": {Data: make([]Color, n)},
	}
	result := NewMesh()
	b.ResetTimer()

	if prealloc {
		for i := 0; i < b.N; i++ {
			result.Append(a)
			result.Reset()
		}
	} else {
		for i := 0; i < b.N; i++ {
			result = NewMesh()
			result.Append(a)
		}
	}
}

func BenchmarkMeshAppend100Prealloc(b *testing.B) {
	benchmarkMeshAppend(b, 100, true)
}

func BenchmarkMeshAppend100Dumb(b *testing.B) {
	benchmarkMeshAppend(b, 100, false)
}

func BenchmarkMeshAppend1kPrealloc(b *testing.B) {
	benchmarkMeshAppend(b, 1000, true)
}

func BenchmarkMeshAppend1kDumb(b *testing.B) {
	benchmarkMeshAppend(b, 1000, false)
}

func BenchmarkMeshAppend4kPrealloc(b *testing.B) {
	benchmarkMeshAppend(b, 16000, true)
}

func BenchmarkMeshAppend4kDumb(b *testing.B) {
	benchmarkMeshAppend(b, 16000, false)
}
