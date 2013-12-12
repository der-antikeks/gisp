package engine

import (
	"sync"

	"github.com/der-antikeks/gisp/math"

	//"github.com/go-gl/gl"
)

type Material interface {
	Program() *Program

	Wireframe() bool
	Opaque() bool

	UpdateUniforms()
	Unbind()

	Dispose()
}

var programCache struct {
	cache map[string]*Program
	sync.Mutex
}

/*
	basic material
*/
type BasicMaterial struct {
	Material

	program *Program

	wireframe bool

	diffuse    math.Color
	opacity    float64
	diffuseMap Texture

	/*
		offsetRepeat math.Vector
		lightMap     *Texture

		bumpMap   *Texture
		bumpScale float64

		normalMap   *Texture
		normalScale math.Vector

		ambient   math.Color
		emissive  math.Color
		specular  math.Color
		shininess float64
	*/
}

func NewBasicMaterial() *BasicMaterial {
	programCache.Lock()
	defer programCache.Unlock()

	if programCache.cache == nil {
		programCache.cache = make(map[string]*Program)
	}

	program, exists := programCache.cache["basic"]
	if !exists {
		vs := `
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
			}`

		fs := `
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
			}`

		attributes := []string{"vertexPosition", "vertexNormal", "vertexUV", "vertexColor"}
		uniforms := []string{
			"projectionMatrix",
			"viewMatrix",
			"modelMatrix",
			"modelViewMatrix",
			"normalMatrix",

			"diffuseMap", "opacity", "diffuse",
		}

		var err error
		program, err = NewProgram(vs, fs, attributes, uniforms)
		if err != nil {
			panic(err.Error())
		}

		programCache.cache["basic"] = program
	}

	return &BasicMaterial{
		program: program,

		diffuse: math.Color{1, 1, 1},
		opacity: 1.0,
	}
}

func (m *BasicMaterial) Program() *Program {
	return m.program
}

func (m *BasicMaterial) Opaque() bool {
	return m.opacity == 1.0
}

func (m *BasicMaterial) Wireframe() bool {
	return m.wireframe
}

func (m *BasicMaterial) Dispose() {
	m.program.Dispose()
	m.diffuseMap.Dispose()
}

func (m *BasicMaterial) SetDiffuseColor(c math.Color) { m.diffuse = c }
func (m *BasicMaterial) DiffuseColor() math.Color     { return m.diffuse }

func (m *BasicMaterial) SetOpacity(f float64) { m.opacity = f }
func (m *BasicMaterial) Opacity() float64     { return m.opacity }

func (m *BasicMaterial) SetDiffuseMap(t Texture) { m.diffuseMap = t }
func (m *BasicMaterial) DiffuseMap() Texture     { return m.diffuseMap }

func (m *BasicMaterial) UpdateUniforms() {
	var usedTextureUnits int

	r, g, b := m.DiffuseColor().R, m.DiffuseColor().G, m.DiffuseColor().B
	m.program.Uniform("diffuse").Uniform3f(float32(r), float32(g), float32(b))
	m.program.Uniform("opacity").Uniform1f(float32(m.Opacity()))

	// bind texture in Texture Unit 0, Set uniform textureMap sampler to use Texture Unit 0
	m.DiffuseMap().Bind(usedTextureUnits)
	m.program.Uniform("diffuseMap").Uniform1i(usedTextureUnits)
	usedTextureUnits++
}

func (m *BasicMaterial) Unbind() {
	m.DiffuseMap().Unbind()
}

/*
	phong material
*/
type PhongMaterial struct {
	Material

	program *Program

	wireframe bool

	diffuse    math.Color
	opacity    float64
	diffuseMap Texture

	offsetRepeat math.Vector
	lightMap     Texture

	bumpMap   Texture
	bumpScale float64

	normalMap   Texture
	normalScale math.Vector

	ambient   math.Color
	emissive  math.Color
	specular  math.Color
	shininess float64
}

func NewPhongMaterial() *PhongMaterial {
	programCache.Lock()
	defer programCache.Unlock()

	if programCache.cache == nil {
		programCache.cache = make(map[string]*Program)
	}

	program, exists := programCache.cache["phong"]
	if !exists {
		vs := `
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
			}`

		fs := `
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
			}`

		attributes := []string{"vertexPosition", "vertexNormal", "vertexUV", "vertexColor"}
		uniforms := []string{
			"projectionMatrix",
			"viewMatrix",
			"modelMatrix",
			"modelViewMatrix",
			"normalMatrix",

			"diffuseMap", "opacity", "diffuse",
		}

		var err error
		program, err = NewProgram(vs, fs, attributes, uniforms)
		if err != nil {
			panic(err.Error())
		}

		programCache.cache["phong"] = program
	}

	return &PhongMaterial{
		program: program,

		diffuse:  math.Color{1, 1, 1},
		ambient:  math.Color{1, 1, 1},
		emissive: math.Color{1, 1, 1},
		specular: math.Color{1, 1, 1},

		opacity: 1.0,
	}
}

func (m *PhongMaterial) Program() *Program {
	return m.program
}

func (m *PhongMaterial) Opaque() bool {
	return m.opacity == 1.0
}

func (m *PhongMaterial) Wireframe() bool {
	return m.wireframe
}

func (m *PhongMaterial) Dispose() {
	m.program.Dispose()
	m.diffuseMap.Dispose()
}

func (m *PhongMaterial) SetDiffuseColor(c math.Color) { m.diffuse = c }
func (m *PhongMaterial) DiffuseColor() math.Color     { return m.diffuse }

func (m *PhongMaterial) SetAmbientColor(c math.Color) { m.ambient = c }
func (m *PhongMaterial) AmbientColor() math.Color     { return m.ambient }

func (m *PhongMaterial) SetSpecularColor(c math.Color) { m.specular = c }
func (m *PhongMaterial) SpecularColor() math.Color     { return m.specular }

func (m *PhongMaterial) SetEmissiveColor(c math.Color) { m.emissive = c }
func (m *PhongMaterial) EmissiveColor() math.Color     { return m.emissive }

func (m *PhongMaterial) SetOpacity(f float64) { m.opacity = f }
func (m *PhongMaterial) Opacity() float64     { return m.opacity }

func (m *PhongMaterial) SetShininess(f float64) { m.shininess = f }
func (m *PhongMaterial) Shininess() float64     { return m.shininess }

func (m *PhongMaterial) SetDiffuseMap(t Texture) { m.diffuseMap = t }
func (m *PhongMaterial) DiffuseMap() Texture     { return m.diffuseMap }

func (m *PhongMaterial) UpdateUniforms() {
	var usedTextureUnits int

	r, g, b := m.DiffuseColor().R, m.DiffuseColor().G, m.DiffuseColor().B
	m.program.Uniform("diffuse").Uniform3f(float32(r), float32(g), float32(b))
	m.program.Uniform("opacity").Uniform1f(float32(m.Opacity()))

	// bind texture in Texture Unit 0, Set uniform textureMap sampler to use Texture Unit 0
	m.DiffuseMap().Bind(usedTextureUnits)
	m.program.Uniform("diffuseMap").Uniform1i(usedTextureUnits)
	usedTextureUnits++
}

func (m *PhongMaterial) Unbind() {
	m.DiffuseMap().Unbind()
}

/*
	wobble material
*/
type WobbleMaterial struct {
	Material

	program *Program

	wireframe bool

	opacity    float64
	diffuseMap Texture

	time float64
}

func NewWobbleMaterial() *WobbleMaterial {
	programCache.Lock()
	defer programCache.Unlock()

	if programCache.cache == nil {
		programCache.cache = make(map[string]*Program)
	}

	program, exists := programCache.cache["wobble"]
	if !exists {
		vs := `
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
			}`

		fs := `
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
			}`

		attributes := []string{"vertexPosition", "vertexNormal", "vertexUV", "vertexColor"}
		uniforms := []string{
			"projectionMatrix",
			"viewMatrix",
			"modelMatrix",
			"modelViewMatrix",
			"normalMatrix",

			"diffuseMap", "time", "opacity",
		}

		var err error
		program, err = NewProgram(vs, fs, attributes, uniforms)
		if err != nil {
			panic(err.Error())
		}

		programCache.cache["wobble"] = program
	}

	return &WobbleMaterial{
		program: program,

		opacity: 1.0,
		time:    1.0,
	}
}

func (m *WobbleMaterial) Program() *Program {
	return m.program
}

func (m *WobbleMaterial) Opaque() bool {
	return m.opacity == 1.0
}

func (m *WobbleMaterial) Wireframe() bool {
	return m.wireframe
}

func (m *WobbleMaterial) Dispose() {
	m.program.Dispose()
	m.diffuseMap.Dispose()
}

func (m *WobbleMaterial) SetOpacity(f float64) { m.opacity = f }
func (m *WobbleMaterial) Opacity() float64     { return m.opacity }

func (m *WobbleMaterial) SetTime(f float64) { m.time = f }
func (m *WobbleMaterial) Time() float64     { return m.time }

func (m *WobbleMaterial) SetDiffuseMap(t Texture) { m.diffuseMap = t }
func (m *WobbleMaterial) DiffuseMap() Texture     { return m.diffuseMap }

func (m *WobbleMaterial) UpdateUniforms() {
	var usedTextureUnits int

	m.program.Uniform("opacity").Uniform1f(float32(m.Opacity()))
	m.program.Uniform("time").Uniform1f(float32(m.Time()))

	// bind texture in Texture Unit 0, Set uniform textureMap sampler to use Texture Unit 0
	m.DiffuseMap().Bind(usedTextureUnits)
	m.program.Uniform("diffuseMap").Uniform1i(usedTextureUnits)
	usedTextureUnits++
}

func (m *WobbleMaterial) Unbind() {
	m.DiffuseMap().Unbind()
}

/*
	blur material
*/
type BlurMaterial struct {
	Material

	program *Program

	diffuseMap Texture
	size       float64
	vertical   bool
}

func NewBlurMaterial(size float64, vertical bool) *BlurMaterial {
	programCache.Lock()
	defer programCache.Unlock()

	if programCache.cache == nil {
		programCache.cache = make(map[string]*Program)
	}

	program, exists := programCache.cache["blur"]
	if !exists {
		vs := `
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
			}`

		fs := `
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
			}`

		attributes := []string{"vertexPosition", "vertexNormal", "vertexUV", "vertexColor"}
		uniforms := []string{
			"projectionMatrix",
			"viewMatrix",
			"modelMatrix",
			"modelViewMatrix",
			"normalMatrix",

			"diffuseMap", "size", "vertical",
		}

		var err error
		program, err = NewProgram(vs, fs, attributes, uniforms)
		if err != nil {
			panic(err.Error())
		}

		programCache.cache["blur"] = program
	}

	return &BlurMaterial{
		program: program,

		size:     size,
		vertical: vertical,
	}
}

func (m *BlurMaterial) Program() *Program {
	return m.program
}

func (m *BlurMaterial) Opaque() bool    { return true }
func (m *BlurMaterial) Wireframe() bool { return false }

func (m *BlurMaterial) Dispose() {
	m.program.Dispose()
	m.diffuseMap.Dispose()
}

func (m *BlurMaterial) SetSize(f float64) { m.size = f }
func (m *BlurMaterial) Size() float64     { return m.size }

func (m *BlurMaterial) SetVertical(b bool) { m.vertical = b }
func (m *BlurMaterial) Vertical() bool     { return m.vertical }

func (m *BlurMaterial) SetDiffuseMap(t Texture) { m.diffuseMap = t }
func (m *BlurMaterial) DiffuseMap() Texture     { return m.diffuseMap }

func (m *BlurMaterial) UpdateUniforms() {
	var usedTextureUnits int

	m.program.Uniform("size").Uniform1f(float32(m.Size()))

	if m.Vertical() {
		m.program.Uniform("vertical").Uniform1i(1)
	} else {
		m.program.Uniform("vertical").Uniform1i(0)
	}

	// bind texture in Texture Unit 0, Set uniform textureMap sampler to use Texture Unit 0
	m.DiffuseMap().Bind(usedTextureUnits)
	m.program.Uniform("diffuseMap").Uniform1i(usedTextureUnits)
	usedTextureUnits++
}

func (m *BlurMaterial) Unbind() {
	m.DiffuseMap().Unbind()
}

/*
	blend material
*/
type BlendMaterial struct {
	Material

	program *Program

	diffuseMapA Texture
	diffuseMapB Texture
	ratio       float64
	opacity     float64
}

func NewBlendMaterial() *BlendMaterial {
	programCache.Lock()
	defer programCache.Unlock()

	if programCache.cache == nil {
		programCache.cache = make(map[string]*Program)
	}

	program, exists := programCache.cache["blend"]
	if !exists {
		vs := `
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
			}`

		fs := `
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
			}`

		attributes := []string{"vertexPosition", "vertexNormal", "vertexUV", "vertexColor"}
		uniforms := []string{
			"projectionMatrix",
			"viewMatrix",
			"modelMatrix",
			"modelViewMatrix",
			"normalMatrix",

			"diffuseMapA", "diffuseMapB", "ratio", "opacity",
		}

		var err error
		program, err = NewProgram(vs, fs, attributes, uniforms)
		if err != nil {
			panic(err.Error())
		}

		programCache.cache["blend"] = program
	}

	return &BlendMaterial{
		program: program,

		opacity: 1.0,
		ratio:   0.5,
	}
}

func (m *BlendMaterial) Program() *Program {
	return m.program
}

func (m *BlendMaterial) Opaque() bool    { return m.opacity == 1.0 }
func (m *BlendMaterial) Wireframe() bool { return false }

func (m *BlendMaterial) Dispose() {
	m.program.Dispose()
	m.diffuseMapA.Dispose()
	m.diffuseMapB.Dispose()
}

func (m *BlendMaterial) SetOpacity(f float64) { m.opacity = f }
func (m *BlendMaterial) Opacity() float64     { return m.opacity }

func (m *BlendMaterial) SetRatio(f float64) { m.ratio = f }
func (m *BlendMaterial) Ratio() float64     { return m.ratio }

func (m *BlendMaterial) SetDiffuseMapA(t Texture) { m.diffuseMapA = t }
func (m *BlendMaterial) DiffuseMapA() Texture     { return m.diffuseMapA }

func (m *BlendMaterial) SetDiffuseMapB(t Texture) { m.diffuseMapB = t }
func (m *BlendMaterial) DiffuseMapB() Texture     { return m.diffuseMapB }

func (m *BlendMaterial) UpdateUniforms() {
	var usedTextureUnits int

	m.program.Uniform("ratio").Uniform1f(float32(m.Ratio()))
	m.program.Uniform("opacity").Uniform1f(float32(m.Opacity()))

	// bind texture in Texture Unit 0, Set uniform textureMap sampler to use Texture Unit 0
	m.DiffuseMapA().Bind(usedTextureUnits)
	m.program.Uniform("diffuseMapA").Uniform1i(usedTextureUnits)
	usedTextureUnits++

	m.DiffuseMapB().Bind(usedTextureUnits)
	m.program.Uniform("diffuseMapB").Uniform1i(usedTextureUnits)
	usedTextureUnits++
}

func (m *BlendMaterial) Unbind() {
	m.DiffuseMapA().Unbind()
	m.DiffuseMapB().Unbind()
}
