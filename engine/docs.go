/*
	opengl renderer engine

	TODO:
		fix:
		* add missing tests
		* documentation

		short-term features:
		* unify shader materials, move buffer binding/attribute enabling into material struct
		* distance field fonts
		* move lights from hardcoded shader to objects
		* fog/skycube

		later:
		* load managers that handle allocation/deallocation
		* scene object loading/unloading, curret scene



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
