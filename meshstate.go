// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gfx

import (
	"reflect"
)

// intMax returns the largest integer.
func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// intMin returns the smallest integer.
func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MeshState represents the state that a mesh is in: specifically what data
// slices it has present (i.e. does it have vertex colors, normals, etc).
//
// A mesh state can be used to determine what the difference is between two
// meshes:
//
//  // TODO(reader): properly RLock/RUnlock the two meshes.
//  var stateA, stateB gfx.MeshState
//  meshA.State(&stateA) // Query mesh A's state.
//  meshB.State(&stateB) // Query mesh B's state.
//
//  // Calculate the difference between the two states.
//  anyDiff := stateA.Diff(stateB)
//  if anyDiff {
//      // The two mesh states have a difference of some sort (i.e. one has
//      // vertex colors and the other does not). Appending them together would
//      // result in some sort of data loss (describes by the stateA receiver
//      // struct)
//
//      // We can check directly by looking at the results of the diff:
//      if stateA.Normals {
//          // Now we're certain that one mesh has normals and the other does
//          // not. Appending the two meshes together would result in the loss
//          // of the normals.
//      }
//  }
//
// It can also be used to directly tell if two meshes could be appended
// together without losing any data from either mesh:
//
//  // TODO(reader): properly RLock/RUnlock the two meshes.
//  var stateA, stateB gfx.MeshState
//  meshA.State(&stateA) // Query mesh A's state.
//  meshB.State(&stateB) // Query mesh B's state.
//
//  // Check the two states for equality.
//  if meshA.State().Equals(meshB.State()) {
//      // The meshes can be appended together with no loss of information.
//  }
//
// If your only intent is to query if a mesh has a given data slice, you can
// check it yourself directly yourself:
//
//  hasNormals := len(mesh.Normals) > 0
//
type MeshState struct {
	// Whether or not indices, vertices, etc are present in the mesh.
	Indices, Vertices, Colors, Normals, Bary bool

	// How many texture coordinate sets are present in the mesh. The boolean
	// value signifies whether the len(texCoord.Slice) > 0 or not.
	TexCoords []bool

	// A map of attribute names to boolean values signifying whether the
	// len(attrib.Data) > 0 or not.
	Attribs map[string]bool
}

// Equals tells if this mesh state is equal to the other one.
func (s *MeshState) Equals(other *MeshState) bool {
	if s.Indices != other.Indices {
		return false
	}
	if s.Vertices != other.Vertices {
		return false
	}
	if s.Colors != other.Colors {
		return false
	}
	if s.Normals != other.Normals {
		return false
	}
	if s.Bary != other.Bary {
		return false
	}
	if len(s.TexCoords) > 0 && len(other.TexCoords) > 0 {
		if len(s.TexCoords) != len(other.TexCoords) {
			return false
		}
		for i, hasData := range s.TexCoords {
			if hasData != other.TexCoords[i] {
				return false
			}
		}
	}
	if len(s.Attribs) > len(other.Attribs) {
		if len(s.Attribs) != len(other.Attribs) {
			return false
		}
		if (s.Attribs == nil) != (other.Attribs == nil) {
			return false
		}
		for name, hasData := range s.Attribs {
			if hasData != other.Attribs[name] {
				return false
			}
		}
	}
	return true
}

// Diff calculates the difference between the mesh states a and b, the result
// is stored inside of s.
//
// If there is any difference between the two, true is returned.
func (s *MeshState) Diff(a, b *MeshState) bool {
	// Compare standard data slices.
	s.Indices = a.Indices != b.Indices
	s.Vertices = a.Vertices != b.Vertices
	s.Colors = a.Colors != b.Colors
	s.Normals = a.Normals != b.Normals
	s.Bary = a.Bary != b.Bary

	// Generate the diff boolean.
	diff := s.Indices || s.Vertices || s.Colors || s.Normals || s.Bary

	// Only compare texture coordinates if we have them.
	if len(a.TexCoords) > 0 && len(b.TexCoords) > 0 {
		// If the lengths are not the same: something is different.
		if len(a.TexCoords) != len(b.TexCoords) {
			diff = true
		}

		// Ensure we have enough space for all of the texture coordinate
		// comparisons.
		maxTexCoords := intMax(len(a.TexCoords), len(b.TexCoords))
		if len(s.TexCoords) < maxTexCoords {
			s.TexCoords = make([]bool, maxTexCoords)
		}

		// First, we compare texture coordinates that we know both have valid
		// indices for (so we don't panic with index out of bounds).
		minTexCoords := intMin(len(a.TexCoords), len(b.TexCoords))
		for i, tc := range a.TexCoords[:minTexCoords] {
			tcDiff := tc == b.TexCoords[i]
			s.TexCoords[i] = tcDiff
			diff = diff || tcDiff
		}

		// Second, for any texture coordinate the other does not have, it is
		// obviously different.
		if maxTexCoords > minTexCoords {
			diff = true
			for i := minTexCoords; i < maxTexCoords; i++ {
				s.TexCoords[i] = true
			}
		}
	}

	// Only compare attributes if we have them.
	if len(a.Attribs) > 0 && len(b.Attribs) > 0 {
		// If the lengths are not the same: something is different.
		if len(a.Attribs) != len(b.Attribs) {
			diff = true
		}

		// Initialize the map, if needed.
		maxAttribs := intMax(len(a.Attribs), len(b.Attribs))
		if s.Attribs == nil && maxAttribs > 0 {
			s.Attribs = make(map[string]bool, maxAttribs)
		}

		// Compare vertex attributes, once for each map.
		for name, attr := range a.Attribs {
			bAttr, ok := b.Attribs[name]
			if !ok {
				diff = true
				s.Attribs[name] = true
				continue
			}
			attrDiff := attr == bAttr
			diff = diff || attrDiff
			s.Attribs[name] = attrDiff
		}

		// It's possible that attributes exist in a.Attribs that do not exist in
		// b.Attribs, so find those now.
		for name := range a.Attribs {
			_, ok := b.Attribs[name]
			if !ok {
				diff = true
				s.Attribs[name] = true
			}
		}
	}
	return diff
}

// State fills the given MeshState structure with information regarding the
// current state of this mesh. For example:
//
//  var state = &MeshState{}
//  mesh.State(&state) // Fill the state structure with the mesh's state.
//
// The read lock of the mesh must be held for this method to operate safely.
func (m *Mesh) State(s *MeshState) {
	s.Indices = len(m.Indices) > 0
	s.Vertices = len(m.Vertices) > 0
	s.Colors = len(m.Colors) > 0
	s.Normals = len(m.Normals) > 0
	s.Bary = len(m.Bary) > 0
	if len(m.TexCoords) > 0 {
		s.TexCoords = make([]bool, len(m.TexCoords))
		for i, tcs := range m.TexCoords {
			s.TexCoords[i] = len(tcs.Slice) > 0
		}
	}
	if len(m.Attribs) > 0 {
		s.Attribs = make(map[string]bool, len(m.Attribs))
		for name, attr := range m.Attribs {
			s.Attribs[name] = reflect.ValueOf(attr.Data).Len() > 0
		}
	}
}
