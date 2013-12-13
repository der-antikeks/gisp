package engine

import (
	"fmt"
	"sync"

	"github.com/der-antikeks/gisp/math"

	"github.com/go-gl/gl"
)

type program struct {
	program gl.Program
	enabled bool

	uniforms   map[string]gl.UniformLocation
	attributes map[string]struct {
		location gl.AttribLocation
		enabled  bool
	}
}

var programCache2 struct {
	cache map[string]*program
	sync.Mutex
}

var programLibrary map[string]struct {
	vertex, fragment string
	uniforms         map[string]interface{} // default value
	attributes       map[string]uint        // size
}

func init() {
	programCache2.cache = make(map[string]*program)

	programLibrary = map[string]struct {
		vertex, fragment string
		uniforms         map[string]interface{}
		attributes       map[string]uint
	}{
		"basic": {
			vertex: `
				#version 330 core

				// Input vertex data, different for all executions of this shader.
				in vec3 vertexPosition;
				in vec3 vertexNormal;
				in vec2 vertexUV;
				in vec2 vertexUV2;
				in vec3 vertexColor;

				// Values that stay constant for the whole mesh.
				uniform mat4 projectionMatrix;
				uniform mat4 viewMatrix;
				uniform mat4 modelMatrix;
				uniform mat4 modelViewMatrix;
				uniform mat3 normalMatrix;

				// Output data, will be interpolated for each fragment.
				out vec2 UV;
				out vec3 Color;

				void main(){
					// Output position of the vertex
					//gl_Position = projectionMatrix * viewMatrix * modelMatrix * vec4(vertexPosition, 1.0);
					gl_Position = projectionMatrix * modelViewMatrix * vec4(vertexPosition, 1.0);

					// UV of the vertex
					UV = vertexUV;

					Color = vertexColor;
				}`,
			fragment: `
				#version 330 core

				// Interpolated values from the vertex shaders
				in vec2 UV;
				in vec3 Color;

				// Values that stay constant for the whole mesh.
				uniform mat4 viewMatrix;
				uniform vec3 diffuse;
				uniform float opacity;
				uniform sampler2D diffuseMap;

				// Output data
				out vec4 fragmentColor;

				void main()
				{
					fragmentColor = vec4( diffuse, opacity );

					vec4 texelColor = texture( diffuseMap, UV );
					fragmentColor = fragmentColor * texelColor;

					fragmentColor = fragmentColor * vec4( Color, opacity );
				}`,
			uniforms: map[string]interface{}{
				"projectionMatrix": [16]float32{}, // matrix.Float32()
				"viewMatrix":       [16]float32{},
				"modelMatrix":      [16]float32{},
				"modelViewMatrix":  [16]float32{},
				"normalMatrix":     [9]float32{}, // matrix.Matrix3Float32()

				"diffuseMap": nil, // texture
				"opacity":    1.0,
				"diffuse":    math.Color{1, 1, 1},
			},
			attributes: map[string]uint{
				"vertexPosition": 3,
				"vertexNormal":   3,
				"vertexUV":       2,
				"vertexColor":    3,
			},
		},
	}
}

type ShaderMaterial struct {
	Material

	program    *program
	uniforms   map[string]interface{} // value
	attributes map[string]uint        // size
}

func NewShaderMaterial(name string) (*ShaderMaterial, error) {
	// shader is in library
	data, found := programLibrary[name]
	if !found {
		return nil, fmt.Errorf("unknown shader name: %v", name)
	}

	// is program cached
	programCache2.Lock()
	defer programCache2.Unlock()

	prg, exists := programCache2.cache[name]
	if !exists {
		// vertex shader
		vshader := gl.CreateShader(gl.VERTEX_SHADER)
		vshader.Source(data.vertex)
		vshader.Compile()
		if vshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
			return nil, fmt.Errorf("vertex shader error: %v", vshader.GetInfoLog())
		}
		defer vshader.Delete()

		// fragment shader
		fshader := gl.CreateShader(gl.FRAGMENT_SHADER)
		fshader.Source(data.fragment)
		fshader.Compile()
		if fshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
			return nil, fmt.Errorf("fragment shader error: %v", fshader.GetInfoLog())
		}
		defer fshader.Delete()

		// program
		prg = &program{
			program:  gl.CreateProgram(),
			uniforms: make(map[string]gl.UniformLocation),
			attributes: make(map[string]struct {
				location gl.AttribLocation
				enabled  bool
			}),
		}

		prg.program.AttachShader(vshader)
		prg.program.AttachShader(fshader)
		prg.program.Link()
		if prg.program.Get(gl.LINK_STATUS) != gl.TRUE {
			return nil, fmt.Errorf("linker error: %v", prg.program.GetInfoLog())
		}

		// locations
		for n, _ := range data.uniforms {
			prg.uniforms[n] = prg.program.GetUniformLocation(n)
		}

		for n, _ := range data.attributes {
			prg.attributes[n] = struct {
				location gl.AttribLocation
				enabled  bool
			}{
				location: prg.program.GetAttribLocation(n),
				enabled:  false,
			}
		}

		// add to cache
		programCache2.cache[name] = prg
	}

	// new material
	mat := &ShaderMaterial{
		program:    prg,
		uniforms:   make(map[string]interface{}),
		attributes: make(map[string]uint),
	}

	// default values
	for n, v := range data.uniforms {
		mat.uniforms[n] = v
	}

	for n, v := range data.attributes {
		mat.attributes[n] = v
	}

	return mat, nil
}

func (m *ShaderMaterial) Program() *Program { return nil }
func (m *ShaderMaterial) Opaque() bool      { return true }
func (m *ShaderMaterial) Wireframe() bool   { return true }

func (m *ShaderMaterial) Dispose() {
	m.program.program.Delete()
}

func (m *ShaderMaterial) DisableAttributes()          {}
func (m *ShaderMaterial) EnableAttribute(name string) {}
func (m *ShaderMaterial) UpdateUniforms()             {}
func (m *ShaderMaterial) Unbind()                     {}
