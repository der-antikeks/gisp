/*
	opengl renderer engine

	TODO:
		fix:
		* add missing tests
		* documentation
		* rendertarget
		* distance field fonts

		needed features, in order:
		* billboards
		* move lights from hardcoded shader to objects
		* fog/skycube
		* load managers that handle allocation/deallocation
		* scene object loading/unloading, current scene, scene preload



	MaterialLoader
	GeometryLoader
	TextureLoader

	scene (object)
		mesh (renderable object)
			3d matrix (transformed by parent)
			material
				program
				textures
					diffuse, light, bump, normal
			geometry
				vertices, normals, uvs
		lights (object)
		fog

	camera (object)
*/

package engine
