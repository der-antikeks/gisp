package game

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"log"
	m "math"
	"os"
	"sync"

	"github.com/der-antikeks/gisp/math"

	"github.com/go-gl/gl"
	"github.com/go-gl/glh"
)

/*
	geometry, material, texture, shader

	LoadGeometry(name)
*/
type AssetLoaderSystem struct {
	lock    sync.Mutex
	path    string
	context *GlContextSystem

	meshbuffers    map[string]*meshbuffer
	shaderPrograms map[string]*shaderprogram
	textures       map[string]*Texture
}

func NewAssetLoaderSystem(path string, context *GlContextSystem) *AssetLoaderSystem {
	s := &AssetLoaderSystem{
		path:    path,
		context: context,

		meshbuffers:    map[string]*meshbuffer{},
		shaderPrograms: map[string]*shaderprogram{},
		textures:       map[string]*Texture{},
	}

	return s
}

/*
	wavefront obj/mtl importer
	http://en.wikipedia.org/wiki/Wavefront_OBJ

	object format
	http://paulbourke.net/dataformats/obj/

	material format
	http://paulbourke.net/dataformats/mtl/

	obj
		o - named object (ignored)
			g - group of polygons

	mtl
		material
		...

var objCache = struct {
	sync.RWMutex
	geometry map[string]Geometry
	material map[string]Material
}{
	geometry: map[string]Geometry{},
	material: map[string]Material{},
}

func LoadObj(path string) (err error, found []string) {
	return nil, nil
}

func loadMtl(path string) (err error, found []string) {
	return nil, nil
}

*/

func (ls *AssetLoaderSystem) GetMeshBuffer(name string) *meshbuffer {
	ls.lock.Lock()
	defer ls.lock.Unlock()

	if mb, found := ls.meshbuffers[name]; found {
		return mb
	}

	mb := &meshbuffer{}

	switch name {
	default:
		log.Fatal("unknown geometry name: ", name)
	case "cube":
		// dimensions
		size := 2.0
		halfSize := size / 2.0

		// vertices
		a := math.Vector{halfSize, halfSize, halfSize}
		b := math.Vector{-halfSize, halfSize, halfSize}
		c := math.Vector{-halfSize, -halfSize, halfSize}
		d := math.Vector{halfSize, -halfSize, halfSize}
		e := math.Vector{halfSize, halfSize, -halfSize}
		f := math.Vector{halfSize, -halfSize, -halfSize}
		g := math.Vector{-halfSize, -halfSize, -halfSize}
		h := math.Vector{-halfSize, halfSize, -halfSize}

		// uvs
		tl := math.Vector{0, 1}
		tr := math.Vector{1, 1}
		bl := math.Vector{0, 0}
		br := math.Vector{1, 0}

		var normal math.Vector

		// front
		normal = math.Vector{0, 0, 1}
		mb.AddFace(
			Vertex{ // a
				position: a,
				normal:   normal,
				uv:       tr,
			}, Vertex{ // b
				position: b,
				normal:   normal,
				uv:       tl,
			}, Vertex{ // c
				position: c,
				normal:   normal,
				uv:       bl,
			})
		mb.AddFace(
			Vertex{
				position: c,
				normal:   normal,
				uv:       bl,
			}, Vertex{
				position: d,
				normal:   normal,
				uv:       br,
			}, Vertex{
				position: a,
				normal:   normal,
				uv:       tr,
			})

		// back
		normal = math.Vector{0, 0, -1}
		mb.AddFace(
			Vertex{
				position: e,
				normal:   normal,
				uv:       tl,
			}, Vertex{
				position: f,
				normal:   normal,
				uv:       bl,
			}, Vertex{
				position: g,
				normal:   normal,
				uv:       br,
			})
		mb.AddFace(
			Vertex{
				position: g,
				normal:   normal,
				uv:       br,
			}, Vertex{
				position: h,
				normal:   normal,
				uv:       tr,
			}, Vertex{
				position: e,
				normal:   normal,
				uv:       tl,
			})

		// top
		normal = math.Vector{0, 1, 0}
		mb.AddFace(
			Vertex{
				position: e,
				normal:   normal,
				uv:       tr,
			}, Vertex{
				position: h,
				normal:   normal,
				uv:       tl,
			}, Vertex{
				position: b,
				normal:   normal,
				uv:       bl,
			})
		mb.AddFace(
			Vertex{
				position: b,
				normal:   normal,
				uv:       bl,
			}, Vertex{
				position: a,
				normal:   normal,
				uv:       br,
			}, Vertex{
				position: e,
				normal:   normal,
				uv:       tr,
			})

		// bottom
		normal = math.Vector{0, -1, 0}
		mb.AddFace(
			Vertex{
				position: f,
				normal:   normal,
				uv:       br,
			}, Vertex{
				position: d,
				normal:   normal,
				uv:       tr,
			}, Vertex{
				position: c,
				normal:   normal,
				uv:       tl,
			})
		mb.AddFace(
			Vertex{
				position: c,
				normal:   normal,
				uv:       tl,
			}, Vertex{
				position: g,
				normal:   normal,
				uv:       bl,
			}, Vertex{
				position: f,
				normal:   normal,
				uv:       br,
			})

		// left
		normal = math.Vector{-1, 0, 0}
		mb.AddFace(
			Vertex{
				position: b,
				normal:   normal,
				uv:       tr,
			}, Vertex{
				position: h,
				normal:   normal,
				uv:       tl,
			}, Vertex{
				position: g,
				normal:   normal,
				uv:       bl,
			})
		mb.AddFace(
			Vertex{
				position: g,
				normal:   normal,
				uv:       bl,
			}, Vertex{
				position: c,
				normal:   normal,
				uv:       br,
			}, Vertex{
				position: b,
				normal:   normal,
				uv:       tr,
			})

		// right
		normal = math.Vector{1, 0, 0}
		mb.AddFace(
			Vertex{
				position: e,
				normal:   normal,
				uv:       tr,
			}, Vertex{
				position: a,
				normal:   normal,
				uv:       tl,
			}, Vertex{
				position: d,
				normal:   normal,
				uv:       bl,
			})
		mb.AddFace(
			Vertex{
				position: d,
				normal:   normal,
				uv:       bl,
			}, Vertex{
				position: f,
				normal:   normal,
				uv:       br,
			}, Vertex{
				position: e,
				normal:   normal,
				uv:       tr,
			})

	case "sphere":
		// dimensions
		radius := 2.0
		widthSegments, heightSegments := 100, 50

		// if widthSegments < 3 {widthSegments = 3}
		// if heightSegments < 2 {heightSegments = 2}

		phiStart, phiLength := 0.0, math.Pi*2
		thetaStart, thetaLength := 0.0, math.Pi

		var vertices, uvs [][]math.Vector

		for y := 0; y <= heightSegments; y++ {
			var verticesRow, uvsRow []math.Vector

			for x := 0; x <= widthSegments; x++ {
				u := float64(x) / float64(widthSegments)
				v := float64(y) / float64(heightSegments)

				vertex := math.Vector{
					-radius * m.Cos(phiStart+u*phiLength) * m.Sin(thetaStart+v*thetaLength),
					radius * m.Cos(thetaStart+v*thetaLength),
					radius * m.Sin(phiStart+u*phiLength) * m.Sin(thetaStart+v*thetaLength),
				}

				verticesRow = append(verticesRow, vertex)
				uvsRow = append(uvsRow, math.Vector{u, 1.0 - v})
			}

			vertices = append(vertices, verticesRow)
			uvs = append(uvs, uvsRow)
		}

		for y := 0; y < heightSegments; y++ {
			for x := 0; x < widthSegments; x++ {
				// vertex id
				v1 := vertices[y][x+1]
				v2 := vertices[y][x]
				v3 := vertices[y+1][x]
				v4 := vertices[y+1][x+1]

				// normals
				n1 := v1.Normalize()
				n2 := v2.Normalize()
				n3 := v3.Normalize()
				n4 := v4.Normalize()

				// uvs
				uv1 := uvs[y][x+1]
				uv2 := uvs[y][x]
				uv3 := uvs[y+1][x]
				uv4 := uvs[y+1][x+1]

				if m.Abs(v1[1]) == radius {
					mb.AddFace(
						Vertex{
							position: v1,
							normal:   n1,
							uv:       uv1,
						}, Vertex{
							position: v3,
							normal:   n3,
							uv:       uv3,
						}, Vertex{
							position: v4,
							normal:   n4,
							uv:       uv4,
						})
				} else if m.Abs(v3[1]) == radius {
					mb.AddFace(
						Vertex{
							position: v1,
							normal:   n1,
							uv:       uv1,
						}, Vertex{
							position: v2,
							normal:   n2,
							uv:       uv2,
						}, Vertex{
							position: v3,
							normal:   n3,
							uv:       uv3,
						})
				} else {
					mb.AddFace(
						Vertex{
							position: v1,
							normal:   n1,
							uv:       uv1,
						}, Vertex{
							position: v2,
							normal:   n2,
							uv:       uv2,
						}, Vertex{
							position: v4,
							normal:   n4,
							uv:       uv4,
						})
					mb.AddFace(
						Vertex{
							position: v2,
							normal:   n2,
							uv:       uv2,
						}, Vertex{
							position: v3,
							normal:   n3,
							uv:       uv3,
						}, Vertex{
							position: v4,
							normal:   n4,
							uv:       uv4,
						})
				}
			}
		}

	}

	mb.MergeVertices()
	mb.ComputeBoundary()
	mb.FaceCount = len(mb.Faces)
	ls.context.MainThread(func() {
		mb.Init()
	})

	ls.meshbuffers[name] = mb
	return mb
}

type Vertex struct {
	position math.Vector
	normal   math.Vector
	uv       math.Vector
}

func (v Vertex) Key(precision int) string {
	return fmt.Sprintf("%v_%v_%v_%v_%v_%v_%v_%v_%v_%v_%v",
		math.Round(v.position[0], precision),
		math.Round(v.position[1], precision),
		math.Round(v.position[2], precision),

		math.Round(v.normal[0], precision),
		math.Round(v.normal[1], precision),
		math.Round(v.normal[2], precision),

		math.Round(v.uv[0], precision),
		math.Round(v.uv[1], precision),
	)
}

type Face struct {
	A, B, C int
}

type meshbuffer struct {
	Vertices    []Vertex // TODO: remove unnecessary slices
	Faces       []Face
	initialized bool

	VertexArrayObject gl.VertexArray
	FaceBuffer        gl.Buffer
	PositionBuffer    gl.Buffer
	NormalBuffer      gl.Buffer
	UvBuffer          gl.Buffer

	Bounding  math.Boundary
	FaceCount int
}

func (g *meshbuffer) AddFace(a, b, c Vertex) {
	offset := len(g.Vertices)
	g.Vertices = append(g.Vertices, a, b, c)
	g.Faces = append(g.Faces, Face{offset, offset + 1, offset + 2})
}

func (g *meshbuffer) MergeVertices() {
	// search and mark duplicate vertices
	lookup := map[string]int{}
	unique := []Vertex{}
	changed := map[int]int{}

	for i, v := range g.Vertices {
		key := v.Key(4)

		if j, found := lookup[key]; !found {
			// new vertex
			lookup[key] = i
			unique = append(unique, v)
			changed[i] = len(unique) - 1
		} else {
			// duplicate vertex
			changed[i] = changed[j]
		}
	}

	// change faces
	cleaned := []Face{}

	for _, f := range g.Faces {
		a, b, c := changed[f.A], changed[f.B], changed[f.C]
		if a == b || b == c || c == a {
			// degenerated face, remove
			continue
		}

		nf := Face{a, b, c}
		cleaned = append(cleaned, nf)
	}

	// replace with cleaned
	g.Vertices = unique
	g.Faces = cleaned
}

func (g *meshbuffer) ComputeBoundary() {
	g.Bounding = math.NewBoundary()
	for _, v := range g.Vertices {
		g.Bounding.AddPoint(v.position)
	}
}

func (g *meshbuffer) Init() {
	if g.initialized {
		return
	}

	// init vertex buffers
	g.VertexArrayObject = gl.GenVertexArray() // vao
	g.FaceBuffer = gl.GenBuffer()             // ebo
	g.PositionBuffer = gl.GenBuffer()         // vbo's
	g.NormalBuffer = gl.GenBuffer()
	g.UvBuffer = gl.GenBuffer()

	g.VertexArrayObject.Bind()

	// init mesh buffers
	faceArray := make([]uint16, len(g.Faces)*3) // uint32 (4 byte) if points > 65535

	nvertices := len(g.Vertices)
	positionArray := make([]float32, nvertices*3)
	normalArray := make([]float32, nvertices*3)
	uvArray := make([]float32, nvertices*2)

	// copy values to buffers
	for i, v := range g.Vertices {
		// position
		positionArray[i*3] = float32(v.position[0])
		positionArray[i*3+1] = float32(v.position[1])
		positionArray[i*3+2] = float32(v.position[2])

		// normal
		normalArray[i*3] = float32(v.normal[0])
		normalArray[i*3+1] = float32(v.normal[1])
		normalArray[i*3+2] = float32(v.normal[2])

		// uv
		uvArray[i*2] = float32(v.uv[0])
		uvArray[i*2+1] = float32(v.uv[1])
	}

	for i, f := range g.Faces {
		faceArray[i*3] = uint16(f.A)
		faceArray[i*3+1] = uint16(f.B)
		faceArray[i*3+2] = uint16(f.C)
	}

	// set mesh buffers

	// position
	g.PositionBuffer.Bind(gl.ARRAY_BUFFER)
	size := len(positionArray) * int(glh.Sizeof(gl.FLOAT))              // float32 - gl.FLOAT, float64 - gl.DOUBLE
	gl.BufferData(gl.ARRAY_BUFFER, size, positionArray, gl.STATIC_DRAW) // gl.DYNAMIC_DRAW

	// normal
	g.NormalBuffer.Bind(gl.ARRAY_BUFFER)
	size = len(normalArray) * int(glh.Sizeof(gl.FLOAT))
	gl.BufferData(gl.ARRAY_BUFFER, size, normalArray, gl.STATIC_DRAW)

	// uv
	g.UvBuffer.Bind(gl.ARRAY_BUFFER)
	size = len(uvArray) * int(glh.Sizeof(gl.FLOAT))
	gl.BufferData(gl.ARRAY_BUFFER, size, uvArray, gl.STATIC_DRAW)

	// face
	g.FaceBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
	size = len(faceArray) * int(glh.Sizeof(gl.UNSIGNED_SHORT)) // gl.UNSIGNED_SHORT 2, gl.UNSIGNED_INT 4
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, size, faceArray, gl.STATIC_DRAW)

	g.initialized = true
}

func (g *meshbuffer) Cleanup() {
	if g.PositionBuffer != 0 {
		g.PositionBuffer.Delete()
	}

	if g.NormalBuffer != 0 {
		g.NormalBuffer.Delete()
	}

	if g.UvBuffer != 0 {
		g.UvBuffer.Delete()
	}

	if g.FaceBuffer != 0 {
		g.FaceBuffer.Delete()
	}

	if g.VertexArrayObject != 0 {
		g.VertexArrayObject.Delete()
	}
}

var shaderProgramLib = map[string]struct {
	Vertex, Fragment string
	Uniforms         map[string]interface{} // name, default value
	Attributes       map[string]uint        // name, size
}{
	"basic": {
		Vertex: `
				#version 330 core

				// Input vertex data, different for all executions of this shader.
				in vec3 vertexPosition;
				in vec3 vertexNormal;
				in vec2 vertexUV;

				// Values that stay constant for the whole mesh.
				uniform mat4 projectionMatrix;
				uniform mat4 viewMatrix;
				uniform mat4 modelMatrix;
				uniform mat4 modelViewMatrix;
				uniform mat3 normalMatrix;

				void main(){
					// Output position of the vertex
					//gl_Position = projectionMatrix * viewMatrix * modelMatrix * vec4(vertexPosition, 1.0);
					gl_Position = projectionMatrix * modelViewMatrix * vec4(vertexPosition, 1.0);
				}`,
		Fragment: `
				#version 330 core

				// Values that stay constant for the whole mesh.
				uniform vec3  diffuse;
				uniform float opacity;

				// Output data
				out vec4 fragmentColor;

				void main()
				{
					fragmentColor = vec4(diffuse, opacity);
				}`,
		Uniforms: map[string]interface{}{ // name, default value
			"projectionMatrix": nil, //[16]float32{}, // matrix.Float32()
			"viewMatrix":       nil, //[16]float32{},
			"modelMatrix":      nil, //[16]float32{},
			"modelViewMatrix":  nil, //[16]float32{},
			"normalMatrix":     nil, //[9]float32{}, // matrix.Matrix3Float32()

			"diffuse": math.Color{1, 1, 1},
			"opacity": 1.0,
		},
		Attributes: map[string]uint{ // name, size
			"vertexPosition": 3,
			"vertexNormal":   3,
			"vertexUV":       2,
		},
	},
	"phong": {
		Vertex: `
				#version 330 core

				// Input vertex data, different for all executions of this shader.
				in vec3 vertexPosition;
				in vec3 vertexNormal;
				in vec2 vertexUV;
				in vec2 vertexUV2;

				// Values that stay constant for the whole mesh.
				uniform mat4 projectionMatrix;
				uniform mat4 viewMatrix;
				uniform mat4 modelMatrix;
				uniform mat4 modelViewMatrix;
				uniform mat3 normalMatrix;

				// Output data, will be interpolated for each fragment.
				out vec2 UV;

				out vec3 Position; //Position_worldspace
				//out vec3 eyeDir;   //EyeDirection_cameraspace 
				out vec3 lightDir; //LightDirection_cameraspace 
				out vec3 Normal;   //Normal_cameraspace 

				void main(){
					// Output position of the vertex, clipspace
					gl_Position = projectionMatrix * modelViewMatrix * vec4(vertexPosition, 1.0);

					// Position of the vertex, worldspace
					//Position_worldspace
					Position = (modelMatrix * vec4(vertexPosition, 1.0)).xyz;

					// Direction from vertex to camera, cameraspace
					//EyeDirection_cameraspace 
					vec3 eyeDir = vec3(0.0, 0.0, 0.0) - (viewMatrix * modelMatrix * vec4(vertexPosition, 1.0)).xyz;

					// Direction from vertex to light, cameraspace
					//LightPosition_worldspace
					vec3 lightPosition = vec3(0.0, 0.0, 0.0);
					//LightDirection_cameraspace 
					lightDir = (viewMatrix * vec4(lightPosition, 1.0)).xyz + eyeDir;

					// Normal of the the vertex, cameraspace
					//Normal_cameraspace 
					Normal = (viewMatrix * modelMatrix * vec4(vertexNormal, 0.0)).xyz;
					
					// UV of the vertex
					UV = vertexUV;
				}`,
		Fragment: `
				#version 330 core

				// Interpolated values from the vertex shaders
				in vec2 UV;

				in vec3 Position; //Position_worldspace
				//in vec3 eyeDir;   //EyeDirection_cameraspace 
				in vec3 lightDir; //LightDirection_cameraspace 
				in vec3 Normal;   //Normal_cameraspace 

				// Values that stay constant for the whole mesh.
				uniform mat4 viewMatrix;
				uniform vec3 diffuse;
				uniform float opacity;
				uniform sampler2D diffuseMap;

				// Output data
				out vec4 fragmentColor;

				void main()
				{
					// Light properties
					vec3 lightColor = vec3(1.0, 1.0, 1.0);
					float lightPower = 50.0f;

					// Distance to the light
					//LightPosition_worldspace
					vec3 lightPosition = vec3(0.0, 0.0, 0.0);
					float distance = length(lightPosition - Position);

					// Normal of the computed fragment, in camera space
					vec3 n = normalize(Normal);

					// Direction of the light (from the fragment to the light)
					vec3 l = normalize(lightDir);

					// Cosine of the angle between the normal and the light direction
					float cosTheta = clamp(dot(n, l), 0, 1);
					
					// Cosine of the angle between the Eye vector and the Reflect vector,
					//float cosAlpha = clamp(dot(normalize(eyeDir), reflect(-l, n)), 0, 1);

					// Material properties
					vec3 materialDiffuseColor = texture(diffuseMap, UV).rgb;
					vec3 materialAmbientColor = vec3(0.5, 0.5, 0.5) * materialDiffuseColor;
					//vec3 materialSpecularColor = vec3(0.3, 0.3, 0.3);

					// Combine colors
					//fragmentColor = vec4(diffuse, opacity);
					//fragmentColor = fragmentColor * materialDiffuseColor;
					//fragmentColor = fragmentColor * vec4( Color, opacity );

					fragmentColor = vec4( 
							materialAmbientColor +
							materialDiffuseColor * lightColor * lightPower * cosTheta / (distance * distance) /*+
							materialSpecularColor * lightColor * lightPower * pow(cosAlpha, 5.0) / (distance * distance)*/,
						opacity);
				}`,
		Uniforms: map[string]interface{}{
			"projectionMatrix": nil, //[16]float32{}, // matrix.Float32()
			"viewMatrix":       nil, //[16]float32{},
			"modelMatrix":      nil, //[16]float32{},
			"modelViewMatrix":  nil, //[16]float32{},
			"normalMatrix":     nil, //[9]float32{}, // matrix.Matrix3Float32()

			"diffuseMap": nil, // texture
			"opacity":    1.0,
			"diffuse":    math.Color{1, 1, 1},

			"ambient":  math.Color{1, 1, 1},
			"emissive": math.Color{1, 1, 1},
			"specular": math.Color{1, 1, 1},
		},
		Attributes: map[string]uint{
			"vertexPosition": 3,
			"vertexNormal":   3,
			"vertexUV":       2,
		},
	},

	"flat": {
		Vertex: `
				#version 330 core

				//#define MAX_LIGHTS 10

				// Input vertex data, different for all executions of this shader
				in vec3 vertexPosition;			// modelspace
				in vec3 vertexNormal;
				in vec2 vertexUV;

				// Values that stay constant for the whole mesh
				uniform mat4 projectionMatrix;	// camera projection
				uniform mat4 viewMatrix;		// camera matrix
				uniform mat4 modelMatrix;		// model matrix
				uniform mat4 modelViewMatrix;	// view * model matrix
				uniform mat3 normalMatrix;		// model-view normal

				//struct LightInfo {
					uniform vec3 lightPosition; // worldspace
					uniform vec3 lightDiffuse;
					uniform float lightPower;
				//};
				//uniform int lightCount;
				//uniform LightInfo Lights[MAX_LIGHTS];

				uniform vec3 ambientColor;		// indirect light

				/*
				struct MaterialInfo {
					vec3 Ka; // Ambient reflectivity
					vec3 Kd; // Diffuse reflectivity
					vec3 Ks; // Specular reflectivity
					float Shininess; // Specular shininess factor
				};
				uniform MaterialInfo Material;
				*/

				// Output data, will be interpolated for each fragment
				 out vec3 lightColor;
				out vec2 UV;

				vec3 adsShading(vec4 position, vec3 norm /*, int idx*/)
				{
					vec4 lightPosCam = viewMatrix * vec4(lightPosition, 1.0); // cameraspace
					vec3 lightDir = normalize(vec3(lightPosCam - position));
					vec3 viewDir = normalize(-position.xyz);
					vec3 reflectDir = reflect(-lightDir, norm);
					float distance = length(lightPosition - (modelMatrix * vec4(vertexPosition,1)).xyz);

					// ambient, simulates indirect lighting
					vec3 amb = ambientColor * vec3(0.1, 0.1, 0.1);

					// diffuse, direct lightning
					float cosTheta = clamp(dot(norm, lightDir), 0.0, 1.0);
					vec3 diff = lightDiffuse * lightPower * cosTheta / (distance * distance);

					// specular, reflective highlight, like a mirror
					float cosAlpha = clamp(dot(viewDir, reflectDir), 0.0, 1.0);
					vec3 spec = vec3(0.3, 0.3, 0.3) * lightDiffuse * lightPower * pow(cosAlpha, 5.0) / (distance * distance);

					return amb + diff + spec;
				}

				void main(){
					// Get the position and normal in camera space
					//vec3 camNorm = normalize(normalMatrix * vertexNormal);
					vec3 camNorm = normalize((viewMatrix * modelMatrix * vec4(vertexNormal, 0.0)).xyz);
					vec4 camPosition = modelViewMatrix * vec4(vertexPosition, 1.0);
					
					// Evaluate the lighting equation
					lightColor = adsShading(camPosition, camNorm);
					/*
					lightColor = vec3(0.0);
					for (int idx = 0; idx < lightCount; idx++)
					{
						lightColor += adsShading(camPosition, camNorm, idx);
					}
					*/
					UV = vertexUV;

					// Output position of the vertex, clipspace
					gl_Position = projectionMatrix * modelViewMatrix * vec4(vertexPosition, 1.0);
				}`,
		Fragment: `
				#version 330 core

				// Interpolated values from the vertex shaders
				 in vec3 lightColor;
				in vec2 UV;

				// Values that stay constant for the whole mesh
				uniform float opacity;
				uniform sampler2D diffuseMap;

				// Output data
				out vec4 fragmentColor;

				void main()
				{
					vec3 materialColor = texture(diffuseMap, UV).rgb;
					fragmentColor = vec4(lightColor * materialColor, opacity);
				}`,
		Uniforms: map[string]interface{}{
			"projectionMatrix": nil, //[16]float32{}, // matrix.Float32()
			"viewMatrix":       nil, //[16]float32{},
			"modelMatrix":      nil, //[16]float32{},
			"modelViewMatrix":  nil, //[16]float32{},
			"normalMatrix":     nil, //[9]float32{}, // matrix.Matrix3Float32()

			"lightPosition": math.Vector{0, 0, 0},
			"lightDiffuse":  math.Color{1, 1, 1}, //
			"lightPower":    50.0,

			"ambientColor": math.Color{1, 1, 1},

			"diffuseMap": nil, // texture
			"opacity":    1.0,
		},
		Attributes: map[string]uint{
			"vertexPosition": 3,
			"vertexNormal":   3,
			"vertexUV":       2,
		},
	},
}

type shaderprogram struct {
	program gl.Program
	enabled bool

	uniforms map[string]struct {
		location gl.UniformLocation
		standard interface{}
	}
	attributes map[string]struct {
		location gl.AttribLocation
		size     uint
		enabled  bool
	}
}

func (ls *AssetLoaderSystem) GetShader(name string) *shaderprogram {
	ls.lock.Lock()
	defer ls.lock.Unlock()

	if s, found := ls.shaderPrograms[name]; found {
		return s
	}

	data, found := shaderProgramLib[name]
	if !found {
		log.Fatal("unknown shader program: ", name)
	}

	s := &shaderprogram{
		uniforms: map[string]struct {
			location gl.UniformLocation
			standard interface{}
		}{},
		attributes: map[string]struct {
			location gl.AttribLocation
			size     uint
			enabled  bool
		}{},
	}

	ls.context.MainThread(func() {
		s.program = gl.CreateProgram()

		// vertex shader
		vshader := gl.CreateShader(gl.VERTEX_SHADER)
		vshader.Source(data.Vertex)
		vshader.Compile()
		if vshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
			log.Fatalf("vertex shader error: %v", vshader.GetInfoLog())
		}
		defer vshader.Delete()

		// fragment shader
		fshader := gl.CreateShader(gl.FRAGMENT_SHADER)
		fshader.Source(data.Fragment)
		fshader.Compile()
		if fshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
			log.Fatalf("fragment shader error: %v", fshader.GetInfoLog())
		}
		defer fshader.Delete()

		// program
		s.program.AttachShader(vshader)
		s.program.AttachShader(fshader)
		s.program.Link()
		if s.program.Get(gl.LINK_STATUS) != gl.TRUE {
			log.Fatalf("linker error: %v", s.program.GetInfoLog())
		}

		// locations
		for n, v := range data.Uniforms {
			s.uniforms[n] = struct {
				location gl.UniformLocation
				standard interface{}
			}{
				location: s.program.GetUniformLocation(n),
				standard: v,
			}
		}

		for n, v := range data.Attributes {
			s.attributes[n] = struct {
				location gl.AttribLocation
				size     uint
				enabled  bool
			}{
				location: s.program.GetAttribLocation(n),
				size:     v,
				enabled:  false,
			}
		}
	})

	ls.shaderPrograms[name] = s
	return s
}

func (s *shaderprogram) DisableAttributes() {
	for n, a := range s.attributes {
		if a.enabled {
			a.location.DisableArray()
			a.enabled = false
			s.attributes[n] = a
		}
	}
}

func (s *shaderprogram) EnableAttribute(name string) {
	a, ok := s.attributes[name]
	if !ok {
		log.Fatal("unknown attribute: ", name)
	}

	if !a.enabled {
		a.location.EnableArray()
		a.enabled = true
		s.attributes[name] = a
	}

	a.location.AttribPointer(a.size, gl.FLOAT, false, 0, nil)
}

func (s *shaderprogram) Cleanup() {
	// TODO: do something
}

type Texture struct {
	File string // loading from entitymanager?

	buffer gl.Texture
	w, h   int
}

// TODO: binding/unbinding from rendersystem
// MainThread(func() {})

// bind texture in Texture Unit slot
func (t Texture) Bind(slot int) {
	t.buffer.Bind(gl.TEXTURE_2D)
	gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(slot))
}

func (t Texture) Unbind() {
	t.buffer.Unbind(gl.TEXTURE_2D)
}

// cleanup
func (t Texture) Cleanup() {
	if t.buffer != 0 {
		t.buffer.Delete()
	}
}

func (ls *AssetLoaderSystem) LoadTexture(path string) (*Texture, error) {
	ls.lock.Lock()
	defer ls.lock.Unlock()

	if t, found := ls.textures[path]; found {
		return t, nil
	}

	// load file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// decode image
	im, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	bounds := im.Bounds()

	// convert to rgba
	rgba, ok := im.(*image.RGBA)
	if !ok {
		rgba = image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, im, image.Pt(0, 0), draw.Src)
	}

	// create texture
	t := &Texture{
		File: path,
		w:    bounds.Dx(),
		h:    bounds.Dy(),
	}

	ls.context.MainThread(func() {
		t.buffer = gl.GenTexture()
		t.buffer.Bind(gl.TEXTURE_2D)

		// set texture parameters
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE) // gl.REPEAT
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE) // gl.REPEAT
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)

		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR) // gl.LINEAR_MIPMAP_LINEAR
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

		// give image(s) to opengl
		gl.TexImage2D(gl.TEXTURE_2D, 0 /*level*/, gl.RGBA,
			rgba.Bounds().Dx(), rgba.Bounds().Dy(),
			0, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)

		// generate mipmaps
		gl.GenerateMipmap(gl.TEXTURE_2D)

		t.buffer.Unbind(gl.TEXTURE_2D)
	})

	ls.textures[path] = t
	return t, nil
}

// TODO:
func (ls *AssetLoaderSystem) NewFramebuffer(w, h int) *Texture {
	t := &Texture{
		buffer: gl.GenTexture(),
		w:      w,
		h:      h,
	}

	t.buffer.Bind(gl.TEXTURE_2D)
	{
		// set texture parameters
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE) // gl.REPEAT
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE) // gl.REPEAT
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)

		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR) // gl.NEAREST
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR) // gl.NEAREST

		// create storage
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
			t.w, t.h,
			0, gl.RGBA, gl.UNSIGNED_BYTE, nil)

		// generate mipmaps
		gl.GenerateMipmap(gl.TEXTURE_2D)
	}
	t.buffer.Unbind(gl.TEXTURE_2D)

	return t
}

func (ls *AssetLoaderSystem) Cleanup() {
	// TODO: unload textures, buffers, programs and empty caches

	for _, m := range ls.meshbuffers {
		ls.context.MainThread(func() {
			m.Cleanup()
		})
	}

	for _, s := range ls.shaderPrograms {
		ls.context.MainThread(func() {
			s.Cleanup()
		})
	}

	for _, t := range ls.textures {
		ls.context.MainThread(func() {
			t.Cleanup()
		})
	}
}
