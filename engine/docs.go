/*
	opengl renderer engine

	TODO:
		fix:
		* add missing tests
		* documentation

		needed features, in order:
		* unify shader materials, move buffer binding/attribute enabling into material struct
		* distance field fonts
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
