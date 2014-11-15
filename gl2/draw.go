// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"fmt"
	"image"

	"azul3d.org/gfx.v2"
	"azul3d.org/gfx.v2/gl2/internal/gl"
	"azul3d.org/lmath.v1"
)

var (
	// Get an matrix which will translate our matrix from ZUpRight to YUpRight
	zUpRightToYUpRight = lmath.CoordSysZUpRight.ConvertMat4(lmath.CoordSysYUpRight)
)

var (
	textureNames  = make([]string, 32)
	texCoordNames = make([]string, 32)
)

func init() {
	for i := 0; i < len(textureNames); i++ {
		textureNames[i] = fmt.Sprintf("Texture%d", i)
	}

	for i := 0; i < len(texCoordNames); i++ {
		texCoordNames[i] = fmt.Sprintf("TexCoord%d", i)
	}
}

func textureName(i int) string {
	if i < len(textureNames) {
		return textureNames[i]
	}
	n := fmt.Sprintf("Texture%d", i)
	textureNames = append(textureNames, n)
	return n
}

func texCoordName(i int) string {
	if i < len(texCoordNames) {
		return texCoordNames[i]
	}
	n := fmt.Sprintf("TexCoord%d", i)
	texCoordNames = append(texCoordNames, n)
	return n
}

// Used as the *gfx.Object.NativeObject interface value.
type nativeObject struct {
	// The graphics object's last-known transform, if they are not equal then
	// the matrices must be recalculated.
	Transform lmath.Mat4

	// The last-known camera transform and projection.
	CameraTransform lmath.Mat4
	Projection      gfx.Mat4

	// Cached pre-calculated matrices to feed into shaders, this way we don't
	// recalculate matrices every single frame but instead only when they
	// actually change.
	model, view, projection, mvp gfx.Mat4

	// The pending occlusion query ID.
	pendingQuery uint32

	// The sample count of the object the last time it was drawn.
	sampleCount int
}

// Implements the gfx.NativeObject interface.
func (n nativeObject) SampleCount() int {
	return n.sampleCount
}

// Implements the gfx.Destroyable interface.
func (n nativeObject) Destroy() {}

func (n nativeObject) needRebuild(o *gfx.Object, c *gfx.Camera) bool {
	if o.Transform.Mat4() != n.Transform {
		return true
	}
	if c.Object.Transform.Mat4() != n.CameraTransform {
		return true
	}
	if c.Projection != n.Projection {
		return true
	}
	return false
}

func (n nativeObject) rebuild(o *gfx.Object, c *gfx.Camera) nativeObject {
	objMat := o.Transform.Mat4()
	n.Transform = objMat

	// The "Model" matrix is the Object's transformation matrix, we feed it
	// directly in.
	n.model = gfx.ConvertMat4(objMat)

	// The "View" matrix is the coordinate system conversion, multiplied
	// against the camera object's transformation matrix
	view := zUpRightToYUpRight
	if c != nil {
		// Apply inverse of camera object transformation.
		camInverse, _ := c.Object.Transform.Mat4().Inverse()
		view = camInverse.Mul(view)
	}
	n.view = gfx.ConvertMat4(view)

	// The "Projection" matrix is the camera's projection matrix.
	projection := lmath.Mat4Identity
	if c != nil {
		projection = c.Projection.Mat4()
	}
	n.projection = gfx.ConvertMat4(projection)

	// The "MVP" matrix is Model * View * Projection matrix.
	mvp := objMat
	mvp = mvp.Mul(view)
	mvp = mvp.Mul(projection)
	n.mvp = gfx.ConvertMat4(mvp)
	return n
}

func (r *Renderer) hookedDraw(rect image.Rectangle, o *gfx.Object, c *gfx.Camera, pre, post func()) {
	// Make the implicit o.Bounds() call required by gfx.Canvas so that the
	// object has a chance to calculate a bounding box before it's data slices
	// are set to nil.
	o.Bounds()

	lock := func() {
		o.Lock()
		if c != nil {
			c.Lock()
		}
	}

	unlock := func() {
		o.Unlock()
		if c != nil {
			c.Unlock()
		}
	}

	// Lock the object until we are completely done drawing it.
	lock()

	var (
		shaderLoaded   chan *gfx.Shader
		meshesLoaded   []chan *gfx.Mesh
		texturesLoaded []chan *gfx.Texture
	)

	// Begin loading shader.
	if o.Shader == nil {
		// Can't draw.
		unlock()
		r.logf("Draw(): object has a nil shader\n")
		return
	}
	o.Shader.RLock()
	shaderNeedLoad := !o.Shader.Loaded
	shaderHasError := len(o.Shader.Error) > 0
	o.Shader.RUnlock()
	if shaderHasError {
		// Can't draw.
		unlock()
		return
	}
	if shaderNeedLoad {
		shaderLoaded = make(chan *gfx.Shader, 1)
		r.LoadShader(o.Shader, shaderLoaded)
	}

	// Begin loading meshes.
	if len(o.Meshes) == 0 {
		// Can't draw.
		unlock()
		r.logf("Draw(): object has no meshes\n")
		return
	}

	for _, m := range o.Meshes {
		m.RLock()
		meshNeedLoad := !m.Loaded || m.HasChanged()
		meshEmpty := !m.Loaded && len(m.Vertices) == 0
		m.RUnlock()
		if meshEmpty {
			// Can't draw.
			unlock()
			r.logf("Draw(): mesh is not loaded and has no vertices\n")
			return
		}
		if meshNeedLoad {
			ch := make(chan *gfx.Mesh, 1)
			r.LoadMesh(m, ch)
			meshesLoaded = append(meshesLoaded, ch)
		}
	}

	// Begin loading textures.
	for _, t := range o.Textures {
		t.RLock()
		texNeedLoad := !t.Loaded
		t.RUnlock()
		if texNeedLoad {
			ch := make(chan *gfx.Texture, 1)
			r.LoadTexture(t, ch)
			texturesLoaded = append(texturesLoaded, ch)
		}
	}

	// Wait for shader, meshes, and textures to finish loading.
	if shaderLoaded != nil {
		<-shaderLoaded
	}
	for _, load := range meshesLoaded {
		<-load
	}
	for _, load := range texturesLoaded {
		<-load
	}

	// Check if the now-loaded shader might have errors.
	o.Shader.RLock()
	shaderHasError = len(o.Shader.Error) > 0
	o.Shader.RUnlock()
	if shaderHasError {
		// Can't draw.
		unlock()
		return
	}

	// Must set at least an empty native object before Draw() returns.
	o.NativeObject = nativeObject{}

	// Ask the render loop to perform drawing.
	r.RenderExec <- func() bool {
		if pre != nil {
			pre()
		}

		// Set global GL state.
		r.setGlobalState()

		// Update the scissor region (effects drawing).
		r.performScissor(rect)

		var ns *nativeShader
		if o.NativeShader != nil {
			ns = o.NativeShader.(*nativeShader)
		}

		// Use the object's state.
		r.useState(ns, o, c)

		// Draw each mesh.
		for _, m := range o.Meshes {
			r.drawMesh(ns, m)
		}

		// Clear the object's state.
		r.clearState(ns, o)

		// Unlock the object now that we are done drawing it.
		unlock()

		// Yield for occlusion query results, if any are available.
		r.queryYield()

		if post != nil {
			post()
		}
		return false
	}
}

func (r *Renderer) findAttribLocation(native *nativeShader, name string) (uint32, bool) {
	location, ok := native.attribLookup[name]
	if ok {
		return uint32(location), true
	}
	location = gl.GetAttribLocation(native.program, glStr(name))
	if location < 0 {
		return 0, false
	}
	return uint32(location), true
}

func (r *Renderer) findUniformLocation(native *nativeShader, name string) int32 {
	location, ok := native.uniformLookup[name]
	if ok {
		return location
	}
	location = gl.GetUniformLocation(native.program, glStr(name))
	if location < 0 {
		// Just for sanity.
		return -1
	}
	return location
}

type texSlot int32

func (r *Renderer) updateUniform(native *nativeShader, name string, value interface{}) {
	location := r.findUniformLocation(native, name)
	if location == -1 {
		// The uniform is not used by the shader program and should just be
		// dropped.
		return
	}

	switch v := value.(type) {
	case texSlot:
		// Special case: Texture input uniform.
		gl.Uniform1i(location, int32(v))

	case bool:
		var intBool int32
		if v {
			intBool = 1
		}
		gl.Uniform1iv(location, 1, &intBool)

	case float32:
		gl.Uniform1fv(location, 1, &v)

	case []float32:
		if len(v) > 0 {
			gl.Uniform1fv(location, int32(len(v)), &v[0])
		}

	case gfx.TexCoord:
		gl.Uniform2fv(location, 1, &v.U)

	case []gfx.TexCoord:
		if len(v) > 0 {
			gl.Uniform2fv(location, int32(len(v)), &v[0].U)
		}

	case gfx.Vec3:
		gl.Uniform3fv(location, 1, &v.X)

	case []gfx.Vec3:
		if len(v) > 0 {
			gl.Uniform3fv(location, int32(len(v)), &v[0].X)
		}

	case gfx.Vec4:
		gl.Uniform4fv(location, 1, &v.X)

	case []gfx.Vec4:
		if len(v) > 0 {
			gl.Uniform4fv(location, int32(len(v)), &v[0].X)
		}

	case gfx.Color:
		gl.Uniform4fv(location, 1, &v.R)

	case []gfx.Color:
		if len(v) > 0 {
			gl.Uniform4fv(location, int32(len(v)), &v[0].R)
		}

	case gfx.Mat4:
		gl.UniformMatrix4fv(location, 1, false, &v[0][0])

	case []gfx.Mat4:
		if len(v) > 0 {
			gl.UniformMatrix4fv(location, int32(len(v)), false, &v[0][0][0])
		}

	default:
		// We don't know of the type at all, ignore it.
	}
}

func (r *Renderer) beginQuery(o *gfx.Object, n nativeObject) nativeObject {
	if r.glArbOcclusionQuery && o.OcclusionTest {
		gl.GenQueries(1, &n.pendingQuery)
		//gl.Execute()
		gl.BeginQuery(gl.SAMPLES_PASSED, n.pendingQuery)
		//gl.Execute()

		// Add the pending query.
		r.pending.Lock()
		r.pending.queries = append(r.pending.queries, pendingQuery{n.pendingQuery, o})
		r.pending.Unlock()
	}
	return n
}

func (r *Renderer) endQuery(o *gfx.Object, n nativeObject) nativeObject {
	if r.glArbOcclusionQuery && o.OcclusionTest {
		gl.EndQuery(gl.SAMPLES_PASSED)
		//gl.Execute()
	}
	return n
}

func (r *Renderer) useState(ns *nativeShader, obj *gfx.Object, c *gfx.Camera) {
	// Use object state.
	r.stateColorWrite([4]bool{obj.WriteRed, obj.WriteGreen, obj.WriteBlue, obj.WriteAlpha})
	r.stateDithering(obj.Dithering)
	r.stateStencilTest(obj.StencilTest)
	r.stateStencilOp(obj.StencilFront, obj.StencilBack)
	r.stateStencilFunc(obj.StencilFront, obj.StencilBack)
	r.stateStencilMask(obj.StencilFront.WriteMask, obj.StencilBack.WriteMask)
	r.stateDepthFunc(obj.DepthCmp)
	r.stateDepthTest(obj.DepthTest)
	r.stateDepthWrite(obj.DepthWrite)
	r.stateFaceCulling(obj.FaceCulling)

	// Begin using the shader.
	shader := obj.Shader
	if r.lastShader != shader {
		r.lastShader = shader

		r.stateProgram(ns.program)

		// Update shader inputs.
		for name := range shader.Inputs {
			value := shader.Inputs[name]
			r.updateUniform(ns, name, value)
		}
	}

	// Consider rebuilding the object's cached matrices, if needed.
	nativeObj := obj.NativeObject.(nativeObject)
	if nativeObj.needRebuild(obj, c) {
		// Rebuild cached matrices.
		nativeObj = nativeObj.rebuild(obj, c)
	}
	obj.NativeObject = nativeObj

	// Add the matrix inputs for the object.
	r.updateUniform(ns, "Model", nativeObj.model)
	r.updateUniform(ns, "View", nativeObj.view)
	r.updateUniform(ns, "Projection", nativeObj.projection)
	r.updateUniform(ns, "MVP", nativeObj.mvp)

	// Set alpha mode.
	r.stateAlphaToCoverage(&r.gpuInfo, obj.AlphaMode == gfx.AlphaToCoverage)
	r.stateBlend(obj.AlphaMode == gfx.AlphaBlend)
	if obj.AlphaMode == gfx.AlphaBlend {
		r.stateBlendColor(obj.Blend.Color)
		r.stateBlendFuncSeparate(obj.Blend)
		r.stateBlendEquationSeparate(obj.Blend)
	}

	switch obj.AlphaMode {
	case gfx.NoAlpha, gfx.AlphaBlend:
		r.updateUniform(ns, "BinaryAlpha", false)

	case gfx.BinaryAlpha, gfx.AlphaToCoverage:
		r.updateUniform(ns, "BinaryAlpha", true)
	}

	// Bind each texture.
	for i, t := range obj.Textures {
		// Ensure there are no feedback loops if we are rendering to a texture.
		if r.rttCanvas != nil {
			cfg := r.rttCanvas.cfg
			color := cfg.Color.NativeTexture
			depth := cfg.Color.NativeTexture
			stencil := cfg.Color.NativeTexture
			native := t.NativeTexture
			if native != nil && (native == color || native == depth || native == stencil) {
				panic("Feedback Loop - Object cannot use the texture that is being drawn to.")
			}
		}

		nt := t.NativeTexture.(*nativeTexture)

		gl.ActiveTexture(gl.TEXTURE0 + uint32(i))
		gl.BindTexture(gl.TEXTURE_2D, nt.id)

		// Load wrap mode.
		uWrap := convertWrap(t.WrapU)
		vWrap := convertWrap(t.WrapV)
		if t.WrapU == gfx.BorderColor || t.WrapV == gfx.BorderColor {
			// We must specify the actual border color then.
			gl.TexParameterfv(gl.TEXTURE_2D, gl.TEXTURE_BORDER_COLOR, &t.BorderColor.R)
			//gl.Execute()
		}
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, uWrap)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, vWrap)

		// Load filter.
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, convertFilter(t.MinFilter))
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, convertFilter(t.MagFilter))

		// If we do not want mipmapping, turn it off. Note that only the
		// minification filter can be mipmapped (mag filter can never be).
		if t.MinFilter.Mipmapped() {
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_BASE_LEVEL, 0)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAX_LEVEL, 1000)
		} else {
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_BASE_LEVEL, 0)
			gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAX_LEVEL, 0)
		}

		// Add uniform input.
		r.updateUniform(ns, textureName(i), texSlot(i))
	}

	// Begin occlusion query.
	obj.NativeObject = r.beginQuery(obj, nativeObj)
}

func (r *Renderer) clearState(ns *nativeShader, obj *gfx.Object) {
	// End occlusion query.
	obj.NativeObject = r.endQuery(obj, obj.NativeObject.(nativeObject))

	// Use no texture.
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.ActiveTexture(gl.TEXTURE0)
}

func (r *Renderer) drawMesh(ns *nativeShader, m *gfx.Mesh) {
	// Grab the native mesh.
	native := m.NativeMesh.(*nativeMesh)

	// Use vertices data.
	location, ok := r.findAttribLocation(ns, "Vertex")
	if ok {
		gl.BindBuffer(gl.ARRAY_BUFFER, native.vertices)
		gl.EnableVertexAttribArray(location)
		defer gl.DisableVertexAttribArray(location)
		gl.VertexAttribPointer(location, 3, gl.FLOAT, false, 0, nil)
	}

	// Use each texture coordinate set data.
	for index, texCoords := range native.texCoords {
		name := texCoordName(index)
		location, ok = r.findAttribLocation(ns, name)
		if ok {
			gl.BindBuffer(gl.ARRAY_BUFFER, texCoords)
			gl.EnableVertexAttribArray(location)
			defer gl.DisableVertexAttribArray(location)
			gl.VertexAttribPointer(location, 2, gl.FLOAT, false, 0, nil)
		}
	}

	// Use each custom vertex data set.
	for name, attrib := range native.attribs {
		for i, vbo := range attrib.vbos {
			// Determine name.
			indexName := name
			if len(attrib.vbos) > 1 {
				indexName = fmt.Sprintf("%s%d", name, i)
			}

			// Find input location.
			location, ok = r.findAttribLocation(ns, indexName)
			if !ok {
				continue
			}

			// Bind the buffer, send each row.
			gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
			for row := uint32(0); row < attrib.rows; row++ {
				l := location + row
				gl.EnableVertexAttribArray(l)
				defer gl.DisableVertexAttribArray(l)
				gl.VertexAttribPointer(l, attrib.size, gl.FLOAT, false, 0, nil)
			}
		}
	}

	if native.indicesCount > 0 {
		// Draw indexed mesh.
		gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, native.indices)
		gl.DrawElements(gl.TRIANGLES, native.indicesCount, gl.UNSIGNED_INT, nil)
	} else {
		// Draw regular mesh.
		gl.DrawArrays(gl.TRIANGLES, 0, native.verticesCount)
	}

	// Unbind buffer to avoid carrying OpenGL state.
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
}
