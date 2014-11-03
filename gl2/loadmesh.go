// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"reflect"
	"runtime"
	"unsafe"

	"azul3d.org/gfx.v1"
	"azul3d.org/gfx/gl2.v2/internal/gl"
)

type nativeAttrib struct {
	size int32    // 1, 2, 3, 4 - parameter to VertexAttribPointer
	rows uint32   // e.g. 1 for vec[2,3,4], 3 for mat3, 4 for mat4.
	vbos []uint32 // length 1 for []gfx.Vec3, literal len() for [][]gfx.Vec3
}

// nativeMesh is stored inside the *Mesh.Native interface and stores vertex
// buffer object ID's.
type nativeMesh struct {
	indices                     uint32
	vertices                    uint32
	texCoords                   []uint32
	attribs                     map[string]*nativeAttrib
	verticesCount, indicesCount int32
	r                           *Renderer
}

func finalizeMesh(n *nativeMesh) {
	n.r.meshesToFree.Lock()
	n.r.meshesToFree.slice = append(n.r.meshesToFree.slice, n)
	n.r.meshesToFree.Unlock()
}

// Implements gfx.Destroyable interface.
func (n *nativeMesh) Destroy() {
	finalizeMesh(n)
}

func (r *Renderer) createVBO() (vboId uint32) {
	// Generate new VBO.
	gl.GenBuffers(1, &vboId)
	//gl.Execute()
	return
}

func (r *Renderer) updateVBO(usageHint int32, dataSize uintptr, dataLength int, data unsafe.Pointer, vboId uint32) {
	// Bind the VBO now.
	gl.BindBuffer(gl.ARRAY_BUFFER, vboId)

	// Fill the VBO with the data.
	gl.BufferData(
		gl.ARRAY_BUFFER,
		int(dataSize*uintptr(dataLength)),
		data,
		uint32(usageHint),
	)
	//gl.Execute()
}

func (r *Renderer) deleteVBO(vboId *uint32) {
	// Delete the VBO.
	if *vboId == 0 {
		return
	}
	gl.DeleteBuffers(1, vboId)
	//gl.Execute()
	*vboId = 0 // Just for safety.
}

// attribSize returns the number of rows, and the size of each row measured in
// 32-bit elements:
//  rows == 1 == float32, gfx.TexCoord, gfx.Vec3, gfx.Vec4, gfx.Color
//  rows == 4 == gfx.Mat4
//
//  size == 1 == float32
//  size == 2 == gfx.TexCoord
//  size == 3 == gfx.Vec3
//  size == 4 == gfx.Vec4, gfx.Color, gfx.Mat4
// ok == false is returned if x is not one of the above types.
func attribSize(x interface{}) (rows uint32, size int32, ok bool) {
	switch x.(type) {
	case float32:
		return 1, 1, true
	case gfx.TexCoord:
		return 1, 2, true
	case gfx.Vec3:
		return 1, 3, true
	case gfx.Vec4, gfx.Color:
		return 1, 4, true
	case gfx.Mat4:
		return 4, 4, true
	}
	return 0, 0, false
}

func (r *Renderer) updateCustomAttribVBO(usageHint int32, name string, attrib gfx.VertexAttrib, n *nativeAttrib) {
	v := reflect.ValueOf(attrib.Data)

	// If it's not a slice, or it's length is zero, then it is invalid.
	if v.Kind() != reflect.Slice || v.Len() == 0 {
		r.logf("VertexAttrib (%q) not of type slice or length is zero\n", name)
		return
	}

	// Are we sending an array of per-vertex attributes or not?
	vIndexZero := v.Index(0)
	isArray := vIndexZero.Kind() == reflect.Slice

	// Do we even have a valid data type? attribSize() will tell us if we do.
	var ok bool
	if isArray {
		n.rows, n.size, ok = attribSize(vIndexZero.Index(0).Interface())
	} else {
		n.rows, n.size, ok = attribSize(vIndexZero.Interface())
	}
	if !ok {
		// Invalid data type.
		r.logf("VertexAttrib (%q) has invalid underlying data type\n", name)
		return
	}

	// Generate vertex buffer objects, if we need to.
	if len(n.vbos) == 0 {
		// Determine the number of VBO's we need to create. For example if we
		// have:
		//  var x [][]float32
		//  numVBO := len(x[0])
		// otherwise if we have:
		//  var x []float32
		//  numVBO := 1
		numVBO := 1
		if isArray {
			numVBO = vIndexZero.Len()
		}

		// Generate them.
		n.vbos = make([]uint32, numVBO)
		gl.GenBuffers(int32(numVBO), &n.vbos[0])
		//gl.Execute()
	}

	// Update VBO's now.
	if isArray {
		for i := 0; i < v.Len(); i++ {
			data := unsafe.Pointer(v.Index(i).Index(0).UnsafeAddr())
			r.updateVBO(
				usageHint,
				uintptr(n.size*4),
				vIndexZero.Len(),
				data,
				n.vbos[i],
			)
		}
	} else {
		data := unsafe.Pointer(vIndexZero.UnsafeAddr())
		r.updateVBO(
			usageHint,
			uintptr(n.size*4),
			v.Len(),
			data,
			n.vbos[0],
		)
	}
}

func (r *Renderer) freeMeshes() {
	// Lock the list.
	r.meshesToFree.Lock()

	// Free the meshes.
	for _, native := range r.meshesToFree.slice {
		// Delete single VBO's.
		gl.DeleteBuffers(1, &native.indices)
		gl.DeleteBuffers(1, &native.vertices)

		// Delete texture coords buffers.
		if len(native.texCoords) > 0 {
			gl.DeleteBuffers(int32(len(native.texCoords)), &native.texCoords[0])
		}

		// Delete custom attribute buffers.
		for _, attrib := range native.attribs {
			gl.DeleteBuffers(int32(len(attrib.vbos)), &attrib.vbos[0])
		}

		// Flush and execute OpenGL commands.
		gl.Flush()
		//gl.Execute()
	}

	// Slice to zero, and unlock.
	r.meshesToFree.slice = r.meshesToFree.slice[:0]
	r.meshesToFree.Unlock()
}

// Implements gfx.Renderer interface.
func (r *Renderer) LoadMesh(m *gfx.Mesh, done chan *gfx.Mesh) {
	// Lock the mesh until we are done loading it.
	m.Lock()
	if m.Loaded && !m.HasChanged() {
		// Mesh is already loaded and has not changed, signal completion and
		// return after unlocking.
		m.Unlock()
		select {
		case done <- m:
		default:
		}
		return
	}

	f := func() bool {
		// Find the native mesh, creating a new one if none exists.
		var native *nativeMesh
		if !m.Loaded {
			native = &nativeMesh{
				r:       r,
				attribs: make(map[string]*nativeAttrib),
			}
		} else {
			native = m.NativeMesh.(*nativeMesh)
		}

		// Determine usage hint.
		usageHint := int32(gl.STATIC_DRAW)
		if m.Dynamic {
			usageHint = gl.DYNAMIC_DRAW
		}

		// Update Indices VBO.
		if !m.Loaded || m.IndicesChanged {
			if len(m.Indices) == 0 {
				// Delete indices VBO.
				r.deleteVBO(&native.indices)
			} else {
				if native.indices == 0 {
					// Create indices VBO.
					native.indices = r.createVBO()
				}
				// Update indices VBO.
				r.updateVBO(
					usageHint,
					unsafe.Sizeof(m.Indices[0]),
					len(m.Indices),
					unsafe.Pointer(&m.Indices[0]),
					native.indices,
				)
				native.indicesCount = int32(len(m.Indices))
			}
			m.IndicesChanged = false
		}

		// Update Vertices VBO.
		if !m.Loaded || m.VerticesChanged {
			if len(m.Vertices) == 0 {
				// Delete vertices VBO.
				r.deleteVBO(&native.vertices)
				native.verticesCount = 0
			} else {
				if native.vertices == 0 {
					// Create vertices VBO.
					native.vertices = r.createVBO()
				}
				// Update vertices VBO.
				r.updateVBO(
					usageHint,
					unsafe.Sizeof(m.Vertices[0]),
					len(m.Vertices),
					unsafe.Pointer(&m.Vertices[0]),
					native.vertices,
				)
				native.verticesCount = int32(len(m.Vertices))
			}
			m.VerticesChanged = false
		}

		allAttribs := make(map[string]gfx.VertexAttrib, len(m.Attribs))
		for k, s := range m.Attribs {
			allAttribs[k] = s
		}
		if m.Colors != nil {
			allAttribs["Color"] = gfx.VertexAttrib{
				Data:    m.Colors,
				Changed: m.ColorsChanged,
			}
		}
		if m.Bary != nil {
			allAttribs["Bary"] = gfx.VertexAttrib{
				Data:    m.Bary,
				Changed: m.BaryChanged,
			}
		}

		// Any texture coordinate sets that were removed should have their
		// VBO's deleted.
		deletedMax := len(m.TexCoords)
		if deletedMax > len(native.texCoords) {
			deletedMax = len(native.texCoords)
		}
		deleted := native.texCoords[:deletedMax]
		native.texCoords = native.texCoords[:deletedMax]
		for _, vbo := range deleted {
			r.deleteVBO(&vbo)
		}

		// Any texture coordinate sets that were added should have VBO's
		// created.
		added := m.TexCoords[len(native.texCoords):]
		toUpdate := m.TexCoords
		for _, set := range added {
			vbo := r.createVBO()
			native.texCoords = append(native.texCoords, vbo)

			// Update the VBO.
			r.updateVBO(
				usageHint,
				unsafe.Sizeof(set.Slice[0]),
				len(set.Slice),
				unsafe.Pointer(&set.Slice[0]),
				vbo,
			)
		}

		// And finally, any texture coordinate sets that were changed need to
		// have their VBO's updated.
		for index, set := range toUpdate {
			if set.Changed {
				// Update the VBO.
				r.updateVBO(
					usageHint,
					unsafe.Sizeof(set.Slice[0]),
					len(set.Slice),
					unsafe.Pointer(&set.Slice[0]),
					native.texCoords[index],
				)
				set.Changed = false
			}
		}

		// Any custom attributes that were removed should have their VBO's
		// deleted.
		for name, attrib := range native.attribs {
			_, exists := allAttribs[name]
			if exists {
				// It still exists.
				continue
			}
			for _, vbo := range attrib.vbos {
				r.deleteVBO(&vbo)
			}
			delete(native.attribs, name)
		}

		// Any custom attributes that were added should have VBO's created.
		for name, attrib := range allAttribs {
			_, exists := native.attribs[name]
			if exists {
				// It already has a VBO.
				continue
			}

			// Update the custom attribute's VBO.
			nAttrib := new(nativeAttrib)
			native.attribs[name] = nAttrib
			r.updateCustomAttribVBO(
				usageHint,
				name,
				attrib,
				nAttrib,
			)
		}

		// And finally, any custom attributes that were changed need to have
		// their VBO's updated.
		for name, attrib := range allAttribs {
			if attrib.Changed {
				// Update the custom attribute's VBO.
				nAttrib := native.attribs[name]
				r.updateCustomAttribVBO(
					usageHint,
					name,
					attrib,
					nAttrib,
				)
				attrib.Changed = false
			}
		}

		// Ensure no buffer is active when we leave (so that OpenGL state is untouched).
		gl.BindBuffer(gl.ARRAY_BUFFER, 0)

		// Flush and execute OpenGL commands.
		gl.Flush()
		//gl.Execute()

		// If the mesh is not loaded, then we need to assign the native mesh
		// and create a finalizer to free the native mesh later.
		if !m.Loaded {
			// Assign the native mesh.
			m.NativeMesh = native

			// Attach a finalizer to the mesh that will later free it.
			runtime.SetFinalizer(native, finalizeMesh)
		}

		// Set the mesh to loaded, clear any data slices if they are not wanted.
		m.Loaded = true
		m.ClearData()

		// Unlock, signal completion, and return.
		m.Unlock()
		select {
		case done <- m:
		default:
		}
		return false // no frame rendered.
	}

	select {
	case r.RenderExec <- f:
	default:
		go func() {
			r.RenderExec <- f
		}()
	}
}
