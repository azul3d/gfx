# Azul3D - gfx

This repository hosts Azul3D's graphics packages. For examples please see the `example` directory.

| Package | Description |
|---------|-------------|
| [azul3d.org/gfx.v2](https://azul3d.org/gfx.v2) | *Generic Go interface to GPU rendering techniques.* |
| [azul3d.org/gfx.v2/window](https://azul3d.org/gfx.v2/window) | *Easiest way to create a window and render graphics to it.* |
| [azul3d.org/gfx.v2/gl2](https://azul3d.org/gfx.v2/gl2) | *OpenGL 2 based graphics renderer.* |

## Version 2

##### [gfx.v2](https://azul3d.org/gfx.v2) package:

* Added `Mesh.Append` method to append two meshes together (see [#21](https://github.com/azul3d/gfx/issues/21)).
* Added `MeshState` type to check if two meshes can append together perfectly (see [#21](https://github.com/azul3d/gfx/issues/21)).
* `TexCoord` and `Color` are now valid types for use in the `Shader.Input` map and as data to `VertexAttrib` (see [#23](https://github.com/azul3d/gfx/issues/23)).
* Added a convenience `Mesh.Normals` slice for storing the normals of a mesh (see [#11](https://github.com/azul3d/gfx/issues/11)).

##### [gfx.v2/window](https://azul3d.org/gfx.v2/window) package:

* Moved to this repository as a sub-package (see [old repository](https://github.com/azul3d/gfx-window) and [issue 33](https://github.com/azul3d/issues/issues/33)).

##### [gfx.v2/gl2](https://azul3d.org/gfx.v2/gl2) package:

* Moved to this repository as a sub-package (see [old repository](https://github.com/azul3d/gfx-gl2) and [issue 33](https://github.com/azul3d/issues/issues/33)).
* Renderer now uses just one OpenGL context (see [#24](https://github.com/azul3d/gfx/issues/24)).
* Improved package documentation ([view commit](https://github.com/azul3d/gfx-gl2/commit/493f72dbb36547e394f2d4995ee7d74dbf7b86d4)).

## Version 1.0.1

##### [gfx.v1](https://azul3d.org/gfx.v1) changes:

* Fixed a bug causing Transforms to be constantly recalculated (see [#16](https://github.com/azul3d/gfx/issues/16)).

## Version 1

##### [gfx.v1](https://azul3d.org/gfx.v1) changes:

* Initial implementation.
