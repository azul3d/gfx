// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gl2

import (
	"errors"
	"fmt"
	"image"
	"io"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"azul3d.org/clock.v1"
	"azul3d.org/gfx.v2"
	"azul3d.org/gfx.v2/gl2/internal/gl"
)

// Used when attempting to create an OpenGL 2.0 renderer in a lesser OpenGL context.
var ErrInvalidVersion = errors.New("invalid OpenGL version; must be at least OpenGL 2.0")

type pendingQuery struct {
	// The ID of the pending occlusion query.
	id uint32

	// The object of the pending occlusion query.
	o *gfx.Object
}

// Renderer is an OpenGL 2 based graphics renderer, it runs independant of the
// window management library being used (GLFW, SDL, Chippy, QML, etc).
//
// The renderer primarily uses two independant OpenGL contexts, one is used for
// rendering and one is used for managing resources like meshes, textures, and
// shaders which allows for asynchronous loading (although it is also possible
// to use only a single OpenGL context for windowing libraries that do not
// support multiple, but this will inheritly disable asynchronous loading).
type Renderer struct {
	*baseCanvas

	// Render execution channels.
	RenderExec chan func() bool

	// Whether or not the existing graphics state should be kept between
	// frames. If set to true before rendering a frame the renderer will ask
	// OpenGL for the existing state, the frame will be rendered, and the old
	// OpenGL state restored. This is particularly useful when the renderer
	// must interoperate with other renderers (e.g. QT5).
	keepState bool

	// The graphics clock.
	clock *clock.Clock

	// GPU limitations.
	gpuInfo gfx.GPUInfo

	// Whether or not certain extensions we use are present or not.
	glArbMultisample, glArbFramebufferObject, glArbOcclusionQuery bool

	// Number of multisampling samples, buffers.
	samples, sampleBuffers int32

	// List of OpenGL texture compression format identifiers.
	compressedTextureFormats []int32

	// A channel which will have one empty struct inside it in the event that
	// a finalizer for a mesh, texture, etc has ran and something needs to be
	// free'd.
	wantFree chan struct{}

	// List of native meshes to free at next frame.
	meshesToFree struct {
		sync.RWMutex
		slice []*nativeMesh
	}

	// List of native shaders to free at next frame.
	shadersToFree struct {
		sync.RWMutex
		slice []*nativeShader
	}

	// List of native texture id's to free at next frame.
	texturesToFree struct {
		sync.RWMutex
		slice []uint32
	}

	// List of native FBO id's to free at next frame.
	fbosToFree struct {
		sync.RWMutex
		slice []uint32
	}

	// List of native render buffer id's to free at next frame.
	renderbuffersToFree struct {
		sync.RWMutex
		slice []uint32
	}

	*graphicsState
	prevGraphicsState *graphicsState
	lastShader        *gfx.Shader

	// Structure used to manage the debug output stream.
	debug struct {
		sync.RWMutex
		W io.Writer
	}

	// Structure used to manage pending occlusion queries.
	pending struct {
		sync.Mutex
		queries []pendingQuery
	}

	// RTT format lookups (from gfx formats to GL ones).
	rttTexFormats map[gfx.TexFormat]int32
	rttDSFormats  map[gfx.DSFormat]int32

	// If non-nil, then we are currently rendering to a texture. It is only
	// touched inside RenderExec.
	*rttCanvas

	// Whether or not our global OpenGL state has been set for this frame.
	stateSetForFrame bool

	// Channel to wait for a Render() call to finish.
	renderComplete chan struct{}
}

// Clock implements the gfx.Renderer interface.
func (r *Renderer) Clock() *clock.Clock {
	return r.clock
}

// Short methods that just call the hooked methods (hooked methods are used in
// rtt.go file for render to texture things).

// Clear implements the gfx.Canvas interface.
func (r *Renderer) Clear(rect image.Rectangle, bg gfx.Color) {
	r.hookedClear(rect, bg, nil, nil)
}

// ClearDepth implements the gfx.Canvas interface.
func (r *Renderer) ClearDepth(rect image.Rectangle, depth float64) {
	r.hookedClearDepth(rect, depth, nil, nil)
}

// ClearStencil implements the gfx.Canvas interface.
func (r *Renderer) ClearStencil(rect image.Rectangle, stencil int) {
	r.hookedClearStencil(rect, stencil, nil, nil)
}

// Draw implements the gfx.Canvas interface.
func (r *Renderer) Draw(rect image.Rectangle, o *gfx.Object, c *gfx.Camera) {
	r.hookedDraw(rect, o, c, nil, nil)
}

// QueryWait implements the gfx.Canvas interface.
func (r *Renderer) QueryWait() {
	r.hookedQueryWait(nil, nil)
}

// Render implements the gfx.Canvas interface.
func (r *Renderer) Render() {
	r.hookedRender(nil, nil)
}

// Implements gfx.Canvas interface.
func (r *Renderer) hookedClear(rect image.Rectangle, bg gfx.Color, pre, post func()) {
	r.RenderExec <- func() bool {
		if pre != nil {
			pre()
		}
		r.performClear(rect, bg)
		r.queryYield()
		if post != nil {
			post()
		}
		return false
	}
}

// Implements gfx.Canvas interface.
func (r *Renderer) hookedClearDepth(rect image.Rectangle, depth float64, pre, post func()) {
	r.RenderExec <- func() bool {
		if pre != nil {
			pre()
		}
		r.performClearDepth(rect, depth)
		r.queryYield()
		if post != nil {
			post()
		}
		return false
	}
}

// Implements gfx.Canvas interface.
func (r *Renderer) hookedClearStencil(rect image.Rectangle, stencil int, pre, post func()) {
	r.RenderExec <- func() bool {
		if pre != nil {
			pre()
		}
		r.performClearStencil(rect, stencil)
		r.queryYield()
		if post != nil {
			post()
		}
		return false
	}
}

func (r *Renderer) hookedQueryWait(pre, post func()) {
	// Ask the render channel to wait for query results now.
	r.RenderExec <- func() bool {
		if pre != nil {
			pre()
		}

		// Flush and execute any pending OpenGL commands.
		gl.Flush()
		//gl.Execute()

		// Wait for occlusion query results to come in.
		r.queryWait()

		if post != nil {
			post()
		}

		// signal render completion.
		r.renderComplete <- struct{}{}
		return false
	}
	<-r.renderComplete
}

func (r *Renderer) hookedRender(pre, post func()) {
	// Ask the render channel to render things now.
	r.RenderExec <- func() bool {
		// If any finalizers have ran and actually want us to free something,
		// then we perform this operation now.
		r.freeMeshes()
		r.freeShaders()
		r.freeTextures()
		r.freeFBOs()
		r.freeRenderbuffers()

		if pre != nil {
			pre()
		}

		// Execute all pending operations.
		for i := 0; i < len(r.RenderExec); i++ {
			f := <-r.RenderExec
			f()
		}

		// Flush and execute any pending OpenGL commands.
		gl.Flush()
		//gl.Execute()

		// Wait for occlusion query results to come in.
		r.queryWait()

		if post != nil {
			post()
		}

		if r.rttCanvas != nil {
			// We are rendering to a texture. We do not need to clear global
			// state, tick the clock, or return true (frame rendered).

			// We do still need to signal render completion.
			r.renderComplete <- struct{}{}
			return false
		}

		// Clear our OpenGL state now.
		r.clearGlobalState()

		// Tick the clock.
		r.clock.Tick()

		// signal render completion.
		r.renderComplete <- struct{}{}
		return true
	}
	<-r.renderComplete
}

// Tries to receive pending occlusion query results, returns immediately if
// none are available yet. Returns the number of queries still pending.
func (r *Renderer) queryYield() int {
	if !r.glArbOcclusionQuery {
		return 0
	}
	r.pending.Lock()
	var (
		available, result int32
	)
	for queryIndex, query := range r.pending.queries {
		gl.GetQueryObjectiv(query.id, gl.QUERY_RESULT_AVAILABLE, &available)
		//gl.Execute()
		if available == gl.TRUE {
			// Get the result then.
			gl.GetQueryObjectiv(query.id, gl.QUERY_RESULT, &result)

			// Delete the query.
			gl.DeleteQueries(1, &query.id)
			//gl.Execute()

			// Update object's sample count.
			nativeObj := query.o.NativeObject.(nativeObject)
			nativeObj.sampleCount = int(result)
			query.o.NativeObject = nativeObj

			// Remove from pending slice.
			r.pending.queries = append(r.pending.queries[:queryIndex], r.pending.queries[queryIndex+1:]...)
		}
	}
	length := len(r.pending.queries)
	r.pending.Unlock()
	return length
}

// Blocks until all pending occlusion query results are received.
func (r *Renderer) queryWait() {
	if !r.glArbOcclusionQuery {
		return
	}

	// We have no choice except to busy-wait until the results come: OpenGL
	// doesn't provide a blocking mechanism for waiting for query results but
	// at least we can runtime.Gosched() other goroutines.
	for i := 0; r.queryYield() > 0; i++ {
		// Only runtime.Gosched() every 16th iteration to avoid bogging down
		// rendering.
		if i != 0 && (i%16) == 0 {
			runtime.Gosched()
		}
	}
}

// GPUInfo implements the gfx.Renderer interface.
func (r *Renderer) GPUInfo() gfx.GPUInfo {
	return r.gpuInfo
}

// Effectively just calls stateScissor(), but passes in the proper bounds
// according to whether or not we are rendering to an rttCanvas or not.
func (r *Renderer) performScissor(rect image.Rectangle) {
	if r.rttCanvas != nil {
		r.stateScissor(r.rttCanvas.Bounds(), rect)
	} else {
		r.stateScissor(r.Bounds(), rect)
	}
}

func (r *Renderer) performClear(rect image.Rectangle, bg gfx.Color) {
	r.setGlobalState()

	// Color write mask effects the glClear call below.
	r.stateColorWrite([4]bool{true, true, true, true})

	// Perform clearing.
	r.performScissor(rect)
	r.stateClearColor(bg)
	gl.Clear(uint32(gl.COLOR_BUFFER_BIT))
}

func (r *Renderer) performClearDepth(rect image.Rectangle, depth float64) {
	r.setGlobalState()

	// Depth write mask effects the glClear call below.
	r.stateDepthWrite(true)

	// Perform clearing.
	r.performScissor(rect)
	r.stateClearDepth(depth)
	gl.Clear(uint32(gl.DEPTH_BUFFER_BIT))
}

func (r *Renderer) performClearStencil(rect image.Rectangle, stencil int) {
	r.setGlobalState()

	// Stencil mask effects the glClear call below.
	r.stateStencilMask(0xFFFF, 0xFFFF)

	// Perform clearing.
	r.performScissor(rect)
	r.stateClearStencil(stencil)
	gl.Clear(uint32(gl.STENCIL_BUFFER_BIT))
}

// UpdateBounds updates the effective bounding rectangle of this renderer. It
// must be called whenever the OpenGL canvas size should change (e.g. on window
// resize).
func (r *Renderer) UpdateBounds(bounds image.Rectangle) {
	r.baseCanvas.setBounds(bounds)
}

func (r *Renderer) setGlobalState() {
	if !r.stateSetForFrame {
		r.stateSetForFrame = true

		if r.keepState {
			// We want to maintain state between frames for cooperation with
			// another renderer. Store the existing graphics state now so that
			// we can restore it after the frame is rendered.
			r.prevGraphicsState = queryExistingState(&r.gpuInfo, r.Bounds())

			// Since the existing state is also not what we think it is, we
			// must update our state now.
			cpy := *r.prevGraphicsState
			r.graphicsState = &cpy
		}

		// Update viewport bounds.
		bounds := r.baseCanvas.Bounds()
		gl.Viewport(0, 0, int32(bounds.Dx()), int32(bounds.Dy()))

		// Enable scissor testing.
		gl.Enable(gl.SCISSOR_TEST)

		// Enable multisampling, if available and wanted.
		if r.glArbMultisample {
			if r.baseCanvas.MSAA() {
				gl.Enable(gl.MULTISAMPLE)
			}
		}
	}
}

func (r *Renderer) clearGlobalState() {
	if r.stateSetForFrame {
		r.stateSetForFrame = false

		// Clear last used state.
		oldState := defaultGraphicsState
		if r.keepState {
			// We want to maintain state between frames for cooperation with
			// another renderer. Use the graphics state that was active when
			// the frame started then.
			oldState = r.prevGraphicsState
		}
		r.graphicsState.load(&r.gpuInfo, r.Bounds(), oldState)

		// Reset last shader so that uniforms are loaded again next frame.
		r.lastShader = nil

		// Disable scissor testing.
		gl.Disable(gl.SCISSOR_TEST)

		// Disable multisampling, if available.
		if r.glArbMultisample {
			gl.Disable(gl.MULTISAMPLE)
		}
	}
}

func (r *Renderer) logf(format string, args ...interface{}) {
	// Log the error.
	r.debug.RLock()
	if r.debug.W != nil {
		fmt.Fprintf(r.debug.W, format, args...)
	}
	r.debug.RUnlock()
}

// SetDebugOutput sets the writer, w, to write debug output to. It will mostly
// contain just shader debug information, but other information may be written
// in the future as well.
func (r *Renderer) SetDebugOutput(w io.Writer) {
	r.debug.RLock()
	r.debug.W = w
	r.debug.RUnlock()
}

func parseVersionString(ver string) (major, minor, release int, vendor string) {
	if len(ver) == 0 {
		// Version string must not be empty
		return
	}

	// According to http://www.opengl.org/sdk/docs/man/xhtml/glGetString.xml
	//
	// the string returned may be 'major.minor' or 'major.minor.release'
	// and may be following by a space and any vendor specific information.

	// First locate a proper version string without vendor specific
	// information.
	var (
		versionString string
		err           error
	)
	if strings.Contains(ver, " ") {
		// It must have vendor information
		split := strings.Split(ver, " ")
		if len(split) > 0 || len(split[0]) > 0 {
			// Everything looks good.
			versionString = split[0]
		} else {
			// Something must be wrong with their vendor string.
			return
		}

		// Store the vendor version information.
		vendor = ver[len(versionString):]
	} else {
		// No vendor information.
		versionString = ver
	}

	// We have a proper version string now without vendor information.
	dots := strings.Count(versionString, ".")
	if dots == 1 {
		// It's a 'major.minor' style string
		versions := strings.Split(versionString, ".")
		if len(versions) == 2 {
			major, err = strconv.Atoi(versions[0])
			if err != nil {
				return
			}

			minor, err = strconv.Atoi(versions[1])
			if err != nil {
				return
			}

		} else {
			return
		}

	} else if dots == 2 {
		// It's a 'major.minor.release' style string
		versions := strings.Split(versionString, ".")
		if len(versions) == 3 {
			major, err = strconv.Atoi(versions[0])
			if err != nil {
				return
			}

			minor, err = strconv.Atoi(versions[1])
			if err != nil {
				return
			}

			release, err = strconv.Atoi(versions[2])
			if err != nil {
				return
			}
		} else {
			return
		}
	}
	return
}

func queryVersion() (major, minor, release int, vendorVersion string) {
	versionString := gl.GoStr(gl.GetString(gl.VERSION))
	return parseVersionString(versionString)
}

func queryShaderVersion() (major, minor, release int, vendorVersion string) {
	versionString := gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION))
	return parseVersionString(versionString)
}

func queryExtensions() map[string]bool {
	// Initialize extensions map
	var (
		extensions = make(map[string]bool)
		extString  = gl.GoStr(gl.GetString(gl.EXTENSIONS))
	)
	if len(extString) > 0 {
		split := strings.Split(extString, " ")
		for _, ext := range split {
			if len(ext) > 0 {
				extensions[ext] = true
			}
		}
	}
	return extensions
}

func extension(name string, extensions map[string]bool) bool {
	_, ok := extensions[name]
	return ok
}

func glStr(s string) *int8 {
	return gl.Str(s + "\x00")
}

// New returns a new OpenGL 2 based graphics renderer. If any error is returned
// then a nil renderer is also returned. This function must be called only when
// an OpenGL 2 context is active.
//
// keepState specifies whether or not the existing graphics state should be
// maintained between frames. If set to true then before rendering a frame the
// graphics state will be saved, the frame rendered, and the old graphics state
// restored again. This is particularly useful when the renderer must cooperate
// with another renderer (e.g. QT5). Do not turn it on needlessly though as it
// does come with a performance cost.
func New(keepState bool) (*Renderer, error) {
	r := &Renderer{
		baseCanvas: &baseCanvas{
			msaa: true,
		},
		RenderExec:     make(chan func() bool, 1024),
		keepState:      keepState,
		renderComplete: make(chan struct{}, 8),
		wantFree:       make(chan struct{}, 1),
		clock:          clock.New(),
	}

	// Initialize OpenGL.
	err := gl.Init()
	if err != nil {
		return nil, fmt.Errorf("OpenGL Error: %v", err)
	}

	// Note: we don't need r.gl.Lock() here because no other goroutines
	// can be using r.ctx yet since we haven't returned from New().

	// Find the renderer's precision.
	var redBits, greenBits, blueBits, alphaBits, depthBits, stencilBits int32
	gl.GetIntegerv(gl.RED_BITS, &redBits)
	gl.GetIntegerv(gl.GREEN_BITS, &greenBits)
	gl.GetIntegerv(gl.BLUE_BITS, &blueBits)
	gl.GetIntegerv(gl.ALPHA_BITS, &alphaBits)
	gl.GetIntegerv(gl.DEPTH_BITS, &depthBits)
	gl.GetIntegerv(gl.STENCIL_BITS, &stencilBits)
	//gl.Execute()

	r.precision.RedBits = uint8(redBits)
	r.precision.GreenBits = uint8(greenBits)
	r.precision.BlueBits = uint8(blueBits)
	r.precision.AlphaBits = uint8(alphaBits)
	r.precision.DepthBits = uint8(depthBits)
	r.precision.StencilBits = uint8(stencilBits)

	exts := queryExtensions()
	extsStr := make([]string, len(exts))
	ei := 0
	for s := range exts {
		extsStr[ei] = s
		ei++
	}

	// Query whether we have the GL_ARB_framebuffer_object extension.
	r.glArbFramebufferObject = extension("GL_ARB_framebuffer_object", exts)

	// Query whether we have the GL_ARB_occlusion_query extension.
	r.glArbOcclusionQuery = extension("GL_ARB_occlusion_query", exts)

	// Query whether we have the GL_ARB_multisample extension.
	r.glArbMultisample = extension("GL_ARB_multisample", exts)
	if r.glArbMultisample {
		// Query the number of samples and sample buffers we have, if any.
		gl.GetIntegerv(gl.SAMPLES, &r.samples)
		gl.GetIntegerv(gl.SAMPLE_BUFFERS, &r.sampleBuffers)
		//gl.Execute() // Needed because glGetIntegerv must execute now.
		r.precision.Samples = int(r.samples)
	}

	// Store GPU info.
	var maxTextureSize, maxVaryingFloats, maxVertexInputs, maxFragmentInputs, occlusionQueryBits int32
	gl.GetIntegerv(gl.MAX_TEXTURE_SIZE, &maxTextureSize)
	gl.GetIntegerv(gl.MAX_VARYING_FLOATS, &maxVaryingFloats)
	gl.GetIntegerv(gl.MAX_VERTEX_UNIFORM_COMPONENTS, &maxVertexInputs)
	gl.GetIntegerv(gl.MAX_FRAGMENT_UNIFORM_COMPONENTS, &maxFragmentInputs)
	if r.glArbOcclusionQuery {
		gl.GetQueryiv(gl.SAMPLES_PASSED, gl.QUERY_COUNTER_BITS, &occlusionQueryBits)
	}
	//gl.Execute()

	// Collect GPU information.
	r.gpuInfo.MaxTextureSize = int(maxTextureSize)
	r.gpuInfo.GLSLMaxVaryingFloats = int(maxVaryingFloats)
	r.gpuInfo.GLSLMaxVertexInputs = int(maxVertexInputs)
	r.gpuInfo.GLSLMaxFragmentInputs = int(maxFragmentInputs)
	r.gpuInfo.GLExtensions = extsStr
	r.gpuInfo.AlphaToCoverage = r.glArbMultisample && r.samples > 0 && r.sampleBuffers > 0
	r.gpuInfo.Name = gl.GoStr(gl.GetString(gl.RENDERER))
	r.gpuInfo.Vendor = gl.GoStr(gl.GetString(gl.VENDOR))
	r.gpuInfo.GLMajor, r.gpuInfo.GLMinor, _, _ = queryVersion()
	r.gpuInfo.GLSLMajor, r.gpuInfo.GLSLMinor, _, _ = queryShaderVersion()
	r.gpuInfo.OcclusionQuery = r.glArbOcclusionQuery && occlusionQueryBits > 0
	r.gpuInfo.OcclusionQueryBits = int(occlusionQueryBits)
	r.gpuInfo.NPOT = extension("GL_ARB_texture_non_power_of_two", exts)
	if r.glArbFramebufferObject {
		// See http://www.opengl.org/wiki/Image_Format for more information.
		//
		// TODO:
		//  GL_DEPTH32F_STENCIL8 and GL_DEPTH_COMPONENT32F via Texture.Format
		//      option. (does it require an extension check with GL 2.0?)
		//  GL_STENCIL_INDEX8 (looks like 4.3+ GL hardware)
		//  GL_RGBA16F, GL_RGBA32F via Texture.Format
		//  Compressed formats (DXT ?)
		//  sRGB formats
		//
		//  GL_RGB16, GL_RGBA16

		r.rttTexFormats = make(map[gfx.TexFormat]int32, 16)
		r.rttDSFormats = make(map[gfx.DSFormat]int32, 16)

		// Formats below are guaranteed to be supported in OpenGL 2.x hardware:
		fmts := r.gpuInfo.RTTFormats

		// Color formats.
		fmts.ColorFormats = append(fmts.ColorFormats, []gfx.TexFormat{
			gfx.RGB,
			gfx.RGBA,
		}...)
		for _, cf := range fmts.ColorFormats {
			r.rttTexFormats[cf] = convertTexFormat(cf)
		}

		// Depth formats.
		fmts.DepthFormats = append(fmts.DepthFormats, []gfx.DSFormat{
			gfx.Depth16,
			gfx.Depth24,
			gfx.Depth32,
			gfx.Depth24AndStencil8,
		}...)
		r.rttDSFormats[gfx.Depth16] = gl.DEPTH_COMPONENT16
		r.rttDSFormats[gfx.Depth24] = gl.DEPTH_COMPONENT24
		r.rttDSFormats[gfx.Depth32] = gl.DEPTH_COMPONENT32

		// Stencil formats.
		fmts.StencilFormats = append(fmts.StencilFormats, []gfx.DSFormat{
			gfx.Depth24AndStencil8,
		}...)
		r.rttDSFormats[gfx.Depth24AndStencil8] = gl.DEPTH24_STENCIL8

		// Sample counts.
		// TODO: Beware integer texture formats -- MSAA can at max be
		//       GL_MAX_INTEGER_SAMPLES with those.
		var maxSamples int32
		gl.GetIntegerv(gl.MAX_SAMPLES, &maxSamples)
		//gl.Execute()
		for i := 0; i < int(maxSamples); i++ {
			fmts.Samples = append(fmts.Samples, i)
		}

		r.gpuInfo.RTTFormats = fmts
	}

	// Grab the current renderer bounds (opengl viewport).
	var viewport [4]int32
	gl.GetIntegerv(gl.VIEWPORT, &viewport[0])
	//gl.Execute()
	r.baseCanvas.bounds = image.Rect(0, 0, int(viewport[2]), int(viewport[3]))

	if keepState {
		// Load the existing graphics state.
		r.graphicsState = queryExistingState(&r.gpuInfo, r.baseCanvas.bounds)
	} else {
		r.graphicsState = defaultGraphicsState
	}

	// Update scissor rectangle.
	r.stateScissor(r.baseCanvas.bounds, r.baseCanvas.bounds)

	// Grab the number of texture compression formats.
	var numFormats int32
	gl.GetIntegerv(gl.NUM_COMPRESSED_TEXTURE_FORMATS, &numFormats)
	//gl.Execute() // Needed because glGetIntegerv must execute now.

	// Store the slice of texture compression formats.
	if numFormats > 0 {
		r.compressedTextureFormats = make([]int32, numFormats)
		gl.GetIntegerv(gl.COMPRESSED_TEXTURE_FORMATS, &r.compressedTextureFormats[0])
		//gl.Execute() // Needed because glGetIntegerv must execute now.
	}
	return r, nil
}
