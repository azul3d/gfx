// Copyright 2014 The Azul3D Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gl2 provides an OpenGL 2 based graphics renderer.
//
// The behavior of the renderer is fully defined in the gfx package, and as
// such the following documentation only makes note of strictly OpenGL related
// caveats (like initialization, etc).
//
// Feedback Loops
//
// When performing render-to-texture (RTT), feedback loops are explicitly
// prohibited.
//
// This means that the renderer will panic if you attempt to draw
// an object to a RTT canvas when the object uses the literal RTT texture in
// itself.
//
// That is, rendering an object with a texture that of final render
// destination will panic. Such recursive drawing is prohibited by OpenGL, and
// as such is now allowed.
//
// Mipmapping
//
// The gfx package allows turning on and off mipmapping of a loaded texture
// dynamically by setting it's minification filter to a mipmapped or
// non-mipmapped one. This OpenGL renderer reflects this behavior, but with a
// small caveat:
//
// Mipmapping can only be toggled post-load in the event that the texture was
// first loaded with a mipmapped filter, or else the request is ignored and the
// texture filter is not changed.
//
// This is done for performance reasons, turning on mipmapping post-load time
// would require a full texture reload (and having it on by default would use
// more memory due to mipmaps always being generated).
//
// Uniforms
//
// A gfx.Shader will have all of it's inputs (from the Shader.Inputs map)
// mapped by name to uniforms within the fragment and vertex shader programs.
//
// If you do not intend to utilize a uniform value, then omit it in your GLSL
// code and it will not be fed into the GLSL program.
//
// The default uniforms are:
//
//  uniform mat4 Model;       -> Model matrix from gfx.Object.Transform
//  uniform mat4 View;        -> View matrix from gfx.Camera.Transform
//  uniform mat4 Projection;  -> Projection matrix from gfx.Camera.Projection
//  uniform mat4 MVP;         -> Premultiplied Model/View/Projection matrix.
//  uniform bool BinaryAlpha; -> See below.
//
// BinaryAlpha is a boolean uniform value that informs the shader of the chosen
// alpha transparency mode of an object. It is set to true if the gfx.Object
// being drawn has a gfx.State.AlphaMode of gfx.BinaryAlpha or if the alpha
// mode is gfx.AlphaToCoverage but the GPU does not support it.
//
// Vertex Attributes
//
// A mesh will have all of it's attributes (from the Mesh.Attribs map) mapped
// by name to attributes within the GLSL fragment and vertex shader programs.
//
// Like uniforms, if you do not intend to utilize a attribute value, then omit
// it in your GLSL code and it will simply not be fed into the GLSL program.
//
// The default vertex attributes are:
//
//  attribute vec3 Vertex;      -> from gfx.Mesh.Vertices and gfx.Mesh.Indices
//  attribute vec4 Color;       -> from gfx.Mesh.Colors
//  attribute vec3 Bary;        -> from gfx.Mesh.Bary
//  attribute vec2 TexCoord[N]; -> [N] is the nth index of gfx.Mesh.TexCoords
//
// Uniform And Attribute Types
//
// In both the case of uniforms as well as attributes, data types from the gfx
// package are mapped directly to their GLSL equivilent:
//
//  gfx.Vec4     -> vec4
//  gfx.Vec3     -> vec3
//  gfx.Color    -> vec4 (GLSL does not have a dedicated color type)
//  gfx.TexCoord -> vec2 (GLSL does not have a dedicated texture coordinate type)
//
// Slices are mapped directly to GLSL arrays, which can be fixed or dynamically
// sized, standard GLSL restrictions apply (such as a lack of dynamic indexing
// on dynamically sized arrays, etc).
package gl2
