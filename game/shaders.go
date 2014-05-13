package game

import (
	"log"
	"sync"

	"github.com/der-antikeks/gisp/math"

	"github.com/go-gl/gl"
)

var shaderCache = struct {
	sync.Mutex
	shaders map[string]*shader
}{
	shaders: map[string]*shader{},
}

var shaderLib = map[string]struct {
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
}

type shader struct {
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

func GetShader(name string) *shader {
	shaderCache.Lock()
	defer shaderCache.Unlock()

	if s, found := shaderCache.shaders[name]; found {
		return s
	}

	data, found := shaderLib[name]
	if !found {
		log.Fatal("unknown shader: ", name)
	}

	s := &shader{
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

	MainThread(func() {
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

	shaderCache.shaders[name] = s
	return s
}

func (s *shader) DisableAttributes() {
	for n, a := range s.attributes {
		if a.enabled {
			a.location.DisableArray()
			a.enabled = false
			s.attributes[n] = a
		}
	}
}

func (s *shader) EnableAttribute(name string) {
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
