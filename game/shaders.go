package game

import (
	"log"
	"sync"

	"github.com/der-antikeks/gisp/math"

	"github.com/go-gl/gl"
)

var shaderProgramCache = struct {
	sync.Mutex
	shaderPrograms map[string]*shaderprogram
}{
	shaderPrograms: map[string]*shaderprogram{},
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

func GetShader(name string) *shaderprogram {
	shaderProgramCache.Lock()
	defer shaderProgramCache.Unlock()

	if s, found := shaderProgramCache.shaderPrograms[name]; found {
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

	shaderProgramCache.shaderPrograms[name] = s
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
