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

var programCache struct {
	cache map[string]*program
	sync.Mutex
}

var programLibrary map[string]struct {
	vertex, fragment string
	uniforms         map[string]interface{} // default value
	attributes       map[string]uint        // size
}

func init() {
	programCache.cache = make(map[string]*program)

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
				"projectionMatrix": nil, //[16]float32{}, // matrix.Float32()
				"viewMatrix":       nil, //[16]float32{},
				"modelMatrix":      nil, //[16]float32{},
				"modelViewMatrix":  nil, //[16]float32{},
				"normalMatrix":     nil, //[9]float32{}, // matrix.Matrix3Float32()

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
		"phong": {
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

					// Color of the vertex
					Color = vertexColor; // * lightColor * cosTheta / (distance*distance);
				}`,
			fragment: `
				#version 330 core

				// Interpolated values from the vertex shaders
				in vec2 UV;
				in vec3 Color;

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
			uniforms: map[string]interface{}{
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
			attributes: map[string]uint{
				"vertexPosition": 3,
				"vertexNormal":   3,
				"vertexUV":       2,
				"vertexColor":    3,
			},
		},
		"wobble": {
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

				void main(){
					// Output position of the vertex, clipspace
					gl_Position = projectionMatrix * modelViewMatrix * vec4(vertexPosition, 1.0);

					// UV of the vertex
					UV = vertexUV;
				}`,
			fragment: `
				#version 330 core

				// Interpolated values from the vertex shaders
				in vec2 UV;

				// Values that stay constant for the whole mesh.
				uniform float opacity;
				uniform float time;
				uniform sampler2D diffuseMap;

				// Output data
				out vec4 fragmentColor;

				void main()
				{
					fragmentColor = texture(
						diffuseMap, 
						UV + 0.005*vec2(
							sin(time + 1024.0*UV.x), 
							cos(time + 768.0*UV.y)
						)
					);
				}`,
			uniforms: map[string]interface{}{
				"projectionMatrix": nil, //[16]float32{}, // matrix.Float32()
				"viewMatrix":       nil, //[16]float32{},
				"modelMatrix":      nil, //[16]float32{},
				"modelViewMatrix":  nil, //[16]float32{},
				"normalMatrix":     nil, //[9]float32{}, // matrix.Matrix3Float32()

				"diffuseMap": nil, // texture
				"opacity":    1.0,
				"time":       1.0,
			},
			attributes: map[string]uint{
				"vertexPosition": 3,
				"vertexNormal":   3,
				"vertexUV":       2,
				"vertexColor":    3,
			},
		},
		"blur": {
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

				void main(){
					// Output position of the vertex, clipspace
					gl_Position = projectionMatrix * modelViewMatrix * vec4(vertexPosition, 1.0);

					// UV of the vertex
					UV = vertexUV;
				}`,
			fragment: `
				#version 330 core

				// Interpolated values from the vertex shaders
				in vec2 UV;

				// Values that stay constant for the whole mesh.
				uniform float size;
				uniform int vertical;
				uniform sampler2D diffuseMap;

				// Output data
				out vec4 fragmentColor;

				void main()
				{
					fragmentColor = vec4(0.0);

					if (vertical == 1) {
						fragmentColor += texture(diffuseMap, vec2(UV.x, UV.y - 4.0*size)) * 0.051;
						fragmentColor += texture(diffuseMap, vec2(UV.x, UV.y - 3.0*size)) * 0.0918;
						fragmentColor += texture(diffuseMap, vec2(UV.x, UV.y - 2.0*size)) * 0.12245;
						fragmentColor += texture(diffuseMap, vec2(UV.x, UV.y - 1.0*size)) * 0.1531;
						fragmentColor += texture(diffuseMap, vec2(UV.x, UV.y)) * 0.1633;
						fragmentColor += texture(diffuseMap, vec2(UV.x, UV.y + 1.0*size)) * 0.1531;
						fragmentColor += texture(diffuseMap, vec2(UV.x, UV.y + 2.0*size)) * 0.12245;
						fragmentColor += texture(diffuseMap, vec2(UV.x, UV.y + 3.0*size)) * 0.0918;
						fragmentColor += texture(diffuseMap, vec2(UV.x, UV.y + 4.0*size)) * 0.051;
					} else {
						fragmentColor += texture(diffuseMap, vec2(UV.x - 4.0*size, UV.y)) * 0.051;
						fragmentColor += texture(diffuseMap, vec2(UV.x - 3.0*size, UV.y)) * 0.0918;
						fragmentColor += texture(diffuseMap, vec2(UV.x - 2.0*size, UV.y)) * 0.12245;
						fragmentColor += texture(diffuseMap, vec2(UV.x - 1.0*size, UV.y)) * 0.1531;
						fragmentColor += texture(diffuseMap, vec2(UV.x, UV.y)) * 0.1633;
						fragmentColor += texture(diffuseMap, vec2(UV.x + 1.0*size, UV.y)) * 0.1531;
						fragmentColor += texture(diffuseMap, vec2(UV.x + 2.0*size, UV.y)) * 0.12245;
						fragmentColor += texture(diffuseMap, vec2(UV.x + 3.0*size, UV.y)) * 0.0918;
						fragmentColor += texture(diffuseMap, vec2(UV.x + 4.0*size, UV.y)) * 0.051;
					}
				}`,
			uniforms: map[string]interface{}{
				"projectionMatrix": nil, //[16]float32{}, // matrix.Float32()
				"viewMatrix":       nil, //[16]float32{},
				"modelMatrix":      nil, //[16]float32{},
				"modelViewMatrix":  nil, //[16]float32{},
				"normalMatrix":     nil, //[9]float32{}, // matrix.Matrix3Float32()

				"diffuseMap": nil, // texture
				"size":       1.0,
				"vertical":   true,
			},
			attributes: map[string]uint{
				"vertexPosition": 3,
				"vertexNormal":   3,
				"vertexUV":       2,
				"vertexColor":    3,
			},
		},
		"blend": {
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

				void main(){
					// Output position of the vertex, clipspace
					gl_Position = projectionMatrix * modelViewMatrix * vec4(vertexPosition, 1.0);

					// UV of the vertex
					UV = vertexUV;
				}`,
			fragment: `
				#version 330 core

				// Interpolated values from the vertex shaders
				in vec2 UV;

				// Values that stay constant for the whole mesh.
				uniform float opacity;
				uniform float ratio;
				uniform sampler2D diffuseMapA;
				uniform sampler2D diffuseMapB;

				// Output data
				out vec4 fragmentColor;

				void main()
				{
					fragmentColor = opacity * mix(texture(diffuseMapA, UV), texture(diffuseMapB, UV), ratio);
				}`,
			uniforms: map[string]interface{}{
				"projectionMatrix": nil, //[16]float32{}, // matrix.Float32()
				"viewMatrix":       nil, //[16]float32{},
				"modelMatrix":      nil, //[16]float32{},
				"modelViewMatrix":  nil, //[16]float32{},
				"normalMatrix":     nil, //[9]float32{}, // matrix.Matrix3Float32()

				"diffuseMapA": nil, // texture
				"diffuseMapB": nil,
				"ratio":       0.5,
				"opacity":     1.0,
			},
			attributes: map[string]uint{
				"vertexPosition": 3,
				"vertexNormal":   3,
				"vertexUV":       2,
				"vertexColor":    3,
			},
		},
		"font": {
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

				void main(){
					// Output position of the vertex, clipspace
					gl_Position = projectionMatrix * modelViewMatrix * vec4(vertexPosition, 1.0);

					// UV of the vertex
					UV = vertexUV;
				}`,
			fragment: `
				#version 330 core

				// Interpolated values from the vertex shaders
				in vec2 UV;

				// Values that stay constant for the whole mesh.
				uniform vec3 diffuse;
				uniform float smoothing;
				uniform sampler2D distanceFieldMap;

				// Output data
				out vec4 fragmentColor;

				void main()
				{
					float distance = texture(distanceFieldMap, UV).a;
					float opacity = smoothstep(0.5 - smoothing, 0.5 + smoothing, distance);
					fragmentColor = vec4(diffuse, opacity);
				}`,
			uniforms: map[string]interface{}{
				"projectionMatrix": nil, //[16]float32{}, // matrix.Float32()
				"viewMatrix":       nil, //[16]float32{},
				"modelMatrix":      nil, //[16]float32{},
				"modelViewMatrix":  nil, //[16]float32{},
				"normalMatrix":     nil, //[9]float32{}, // matrix.Matrix3Float32()

				"distanceFieldMap": nil, // texture
				"smoothing":        0.25,
				"diffuse":          math.Color{1, 1, 1},
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

type Material struct {
	program    *program
	wireframe  bool
	opaque     bool
	uniforms   map[string]interface{} // value
	attributes map[string]uint        // size
}

func NewMaterial(name string) (*Material, error) {
	// is shader in library?
	data, found := programLibrary[name]
	if !found {
		return nil, fmt.Errorf("unknown shader name: %v", name)
	}

	// is program cached?
	programCache.Lock()
	defer programCache.Unlock()

	prg, exists := programCache.cache[name]
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
		programCache.cache[name] = prg
	}

	// new material
	mat := &Material{
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

func (m *Material) SetWireframe(b bool) {
	m.wireframe = b
}

func (m *Material) Wireframe() bool {
	return m.wireframe
}

func (m *Material) SetOpaque(b bool) {
	m.opaque = b
}

func (m *Material) Opaque() bool {
	if o, ok := m.uniforms["opacity"]; ok {
		return o == 1.0
	}
	return m.opaque
}

func (m *Material) UseProgram() bool {
	if m.program.enabled {
		return false
	}

	programCache.Lock()
	defer programCache.Unlock()
	for _, prg := range programCache.cache {
		if prg.enabled {
			prg.enabled = false
			break
		}
	}

	m.program.program.Use()
	m.program.enabled = true
	return true
}

func (m *Material) Dispose() {
	m.program.program.Delete()

	for _, u := range m.uniforms {
		switch t := u.(type) {
		case Texture:
			t.Dispose()
		}
	}
}

func (m *Material) DisableAttributes() {
	for n, v := range m.program.attributes {
		if v.enabled {
			v.location.DisableArray()
			v.enabled = false
			m.program.attributes[n] = v
		}
	}
}

func (m *Material) EnableAttribute(name string) {
	if _, ok := m.attributes[name]; !ok {
		//return err
		panic("unknown attribute: " + name)
	}

	if v := m.program.attributes[name]; !v.enabled {
		v.location.EnableArray()
		v.enabled = true

		m.program.attributes[name] = v
	}

	m.program.attributes[name].location.AttribPointer(m.attributes[name], gl.FLOAT, false, 0, nil)
}

func (m *Material) SetUniform(name string, value interface{}) {
	if _, ok := m.uniforms[name]; !ok {
		return
	}
	m.uniforms[name] = value
}

func (m *Material) Uniform(name string) interface{} {
	return m.uniforms[name]
}

func (m *Material) UpdateUniforms() /*error*/ {
	var usedTextureUnits int

	for n, v := range m.uniforms {
		switch t := v.(type) {
		case Texture:
			t.Bind(usedTextureUnits)
			//m.program.uniforms[n].Uniform1i(usedTextureUnits)
			if err := m.UpdateUniform(n, usedTextureUnits); err != nil {
				//return err
				panic(err.Error())
			}
			usedTextureUnits++

		case nil: // ignore nil

		default:
			if err := m.UpdateUniform(n, v); err != nil {
				//return err
				panic(err.Error())
			}
		}
	}

	return //nil
}

func (m *Material) UpdateUniform(name string, value interface{}) error {
	switch t := value.(type) {
	case int:
		m.program.uniforms[name].Uniform1i(t)
	case float64:
		m.program.uniforms[name].Uniform1f(float32(t))
	case float32:
		m.program.uniforms[name].Uniform1f(t)

	case [16]float32:
		m.program.uniforms[name].UniformMatrix4fv(false, t)
	case [9]float32:
		m.program.uniforms[name].UniformMatrix3fv(false, t)

	case math.Color:
		m.program.uniforms[name].Uniform3f(float32(t.R), float32(t.G), float32(t.B))

	case bool:
		if t {
			m.program.uniforms[name].Uniform1i(1)
		} else {
			m.program.uniforms[name].Uniform1i(0)
		}

	default:
		panic(fmt.Sprintf("%v has unknown type: %T", name, value))
		return fmt.Errorf("%v has unknown type: %T", name, value)
	}

	return nil
}

func (m *Material) Unbind() {
	for _, v := range m.uniforms {
		switch t := v.(type) {
		case Texture:
			t.Unbind()
		}
	}
}
