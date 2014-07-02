package main

import (
	"bufio"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-gl/gl"
	glfw "github.com/go-gl/glfw3"
	"github.com/go-gl/glh"
	"github.com/go-gl/mathgl/mgl32"
)

func main() {
	// init
	width, height := 800, 400

	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	rand.Seed(time.Now().Unix())
	runtime.LockOSThread()

	// setup glfw
	if !glfw.Init() {
		log.Fatal("Can't open GLFW")
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, 0)
	glfw.WindowHint(glfw.Samples, 4)

	window, err := glfw.CreateWindow(width, height, "Testing", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	window.MakeContextCurrent()
	glfw.SwapInterval(1)

	// setup gl
	gl.Init()
	gl.ClearColor(0.0, 0.0, 0.0, 0.0)
	gl.ClearDepth(1)
	gl.ClearStencil(0)

	//gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	gl.Enable(gl.CULL_FACE)
	gl.FrontFace(gl.CCW)
	gl.CullFace(gl.BACK)

	gl.ShadeModel(gl.SMOOTH)
	gl.Hint(gl.PERSPECTIVE_CORRECTION_HINT, gl.NICEST)

	//gl.Enable(gl.BLEND)
	//gl.BlendEquationSeparate(gl.FUNC_ADD, gl.FUNC_ADD)
	//gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

	// first pass, geometry calculation program
	// vertex, fragment (!)
	geometryProgram := NewShaderProgram(LoadShader(`
		#version 330 core

		layout(location = 0) in vec3 vertexPosition;
		layout(location = 1) in vec3 vertexNormal;
		layout(location = 2) in vec2 vertexUV;

		uniform mat4 projectionMatrix;
		uniform mat4 viewMatrix;
		uniform mat4 modelMatrix;
		uniform mat3 normalMatrix;

		out vec3 Position;
		out vec3 Normal;
		out vec2 UV;

		void main() {
			Position = (modelMatrix * vec4(vertexPosition, 1.0)).xyz;
			Normal = (modelMatrix * vec4(vertexNormal, 0.0)).xyz;
			UV = vertexUV;

			gl_Position = projectionMatrix * viewMatrix * modelMatrix * vec4(vertexPosition, 1.0);
		}
	`, `
		#version 330 core

		in vec3 Position;
		in vec3 Normal;
		in vec2 UV;

		// material attributes
		uniform vec3 Diffuse;
		uniform sampler2D DiffuseMap;
		uniform float SpecularIntensity;
		uniform float SpecularPower;

		// g-buffer
		layout(location = 0) out vec4 fragmentColor;
		layout(location = 1) out vec4 fragmentPosition; // world-space, move to eye-space?
		layout(location = 2) out vec4 fragmentNormal;

		void main() {
			vec3 materialColor = texture(DiffuseMap, UV).rgb;

			fragmentColor = vec4(materialColor * Diffuse, SpecularIntensity);
			fragmentPosition = vec4(Position, 1.0);
			fragmentNormal = vec4(normalize(Normal), SpecularPower);
		}
	`), []ShaderUniform{
		{Name: "projectionMatrix"},
		{Name: "viewMatrix"},
		{Name: "modelMatrix"},
		{Name: "normalMatrix"},

		{Name: "Diffuse"},
		{Name: "DiffuseMap"},
		{Name: "SpecularIntensity"},
		{Name: "SpecularPower"},
	}, []ShaderAttribute{
		{Name: "vertexPosition", Stride: 3, Typ: gl.FLOAT},
		{Name: "vertexNormal", Stride: 3, Typ: gl.FLOAT},
		{Name: "vertexUV", Stride: 2, Typ: gl.FLOAT},
	})
	defer geometryProgram.Delete()

	// setup mesh buffers
	cubeMesh := MeshBufferFromMesh(LoadObjFile("assets/cube/cube.obj"))
	defer cubeMesh.Delete()

	sphereMesh := MeshBufferFromMesh(GenerateSphere(1.5, 10, 5))
	defer sphereMesh.Delete()

	// setup texture
	gridTexture := LoadTexture("assets/uvtemplate.png")
	defer gridTexture.Delete()

	moonTexture := LoadTexture("assets/planets/moon_1024.jpg")
	defer moonTexture.Delete()

	// second pass, lighting calculation program
	// vertex, fragment (!)
	lightingProgram := NewShaderProgram(LoadShader(`
		#version 330 core

		layout(location = 0) in vec3 vertexPosition;
		layout(location = 1) in vec2 vertexUV;

		uniform mat4 projectionMatrix;
		uniform mat4 viewMatrix;
		uniform mat4 modelMatrix;

		out vec2 UV;

		void main() {
			UV = vertexUV;
			gl_Position = projectionMatrix * viewMatrix * modelMatrix * vec4(vertexPosition, 1.0);
		}
	`, `
		#version 330 core

		in vec2 UV;

		uniform sampler2D colorMap;
		uniform sampler2D positionMap;
		uniform sampler2D normalMap;

		layout(location = 0) out vec4 fragmentColor;

		void main() {
			vec4 color = texture(colorMap, UV);
			vec3 position = texture(positionMap, UV).xyz;
			vec3 normal = normalize(texture(normalMap, UV).xyz);

			vec3 lightColor = vec3(1, 1, 1);
			vec3 lightPosition = vec3(4, 8, 0);
			vec3 lightDir = normalize(lightPosition - position);

			// ambient, simulates indirect lighting
			vec3 amb = lightColor * vec3(0.1, 0.1, 0.1);

			// diffuse, direct lightning
			float cosTheta = clamp(dot(normal, lightDir), 0.0, 1.0);
			vec3 diff = lightColor * cosTheta;

			// specular, reflective highlight, like a mirror
			float cosAlpha = clamp(dot(normalize(-position), reflect(-lightDir, normal)), 0.0, 1.0);
			vec3 spec = vec3(0.3, 0.3, 0.3) * lightColor * pow(cosAlpha, 5.0);

			fragmentColor = vec4(color.rgb * (amb + diff + spec), color.a);
		}
	`), []ShaderUniform{
		{Name: "projectionMatrix"},
		{Name: "viewMatrix"},
		{Name: "modelMatrix"},
		{Name: "normalMatrix"},
		{Name: "colorMap"},
		{Name: "positionMap"},
		{Name: "normalMap"},
	}, []ShaderAttribute{
		{Name: "vertexPosition", Stride: 3, Typ: gl.FLOAT},
		{Name: "vertexUV", Stride: 2, Typ: gl.FLOAT},
	})
	defer lightingProgram.Delete()

	// vertex, fragment (!)
	lightTestProgram := NewShaderProgram(LoadShader(`
		#version 330 core

		layout(location = 0) in vec3 vertexPosition;

		uniform mat4 projectionMatrix;
		uniform mat4 viewMatrix;
		uniform mat4 modelMatrix;

		void main() {
			gl_Position = projectionMatrix * viewMatrix * modelMatrix * vec4(vertexPosition, 1.0);
		}
	`, `
		#version 330 core

		uniform sampler2D colorMap;
		uniform sampler2D positionMap;
		uniform sampler2D normalMap;

		struct Light {
			// basic
			vec3 Position;
			vec3 Color;
			float AmbientIntensity;
			float DiffuseIntensity;
			
			// attenuation
			float Range;
			float Falloff;

			// spot
			vec3 Direction;
			float Angle;
		};
		uniform Light light;

		uniform vec3 camPosition;
		uniform vec2 screenSize;

		layout(location = 0) out vec4 fragmentColor;

		vec2 calcUV() {
			return gl_FragCoord.xy / screenSize;
		}

		vec3 calcPhong(vec3 normal, vec3 lightDir, vec3 viewDir, float specPower, float specIntensity) {
			// ambient, simulates indirect lighting
			vec3 ambColor = light.Color * light.AmbientIntensity;

			// diffuse, direct lightning
			float diffFactor = clamp(dot(normal, lightDir), 0.0, 1.0);
			vec3 diffColor = light.Color * light.DiffuseIntensity * diffFactor;

			// specular, reflective highlight, like a mirror
			float specFactor = clamp(dot(viewDir, reflect(-lightDir, normal)), 0.0, 1.0);
			specFactor = pow(specFactor, specPower);
			vec3 specColor = light.Color * specIntensity * specFactor;

			return (ambColor + diffColor + specColor);
		}

		// attenuation, distance fading effect
		float calcAttenuation(float distance) {
			float attFactor = clamp(1 - (distance / light.Range), 0.0, 1.0);
			return pow(attFactor, light.Falloff);
		}

		// spot, shedding light only within a limited cone
		float calcSpotCone(vec3 lightDir) {
			float spotFactor = dot(-lightDir, light.Direction);
			return clamp((1.0 - (1.0 - spotFactor) * 1.0 / (1.0 - light.Angle)), 0.0, 1.0);
		}

		void main() {
			// get data from g-buffer
			vec2 UV = calcUV();
			vec3 color = texture(colorMap, UV).rgb;
			vec3 position = texture(positionMap, UV).xyz;
			vec3 normal = normalize(texture(normalMap, UV).xyz);
			float matSpecularIntensity = texture(colorMap, UV).a;
			float matSpecularPower = texture(normalMap, UV).w;

			// calculate directions
			vec3 viewDir = normalize(camPosition - position);
			vec3 lightDir = normalize(light.Position - position);
			float distance = length(light.Position - position);

			// calculate factors
			vec3 phong = calcPhong(normal, lightDir, viewDir, matSpecularPower, matSpecularIntensity);
			float att = calcAttenuation(distance);
			float spot = calcSpotCone(lightDir);

			// combine colors
			fragmentColor = vec4(color * phong * att * spot, 1.0);
		}
	`), []ShaderUniform{
		{Name: "projectionMatrix"},
		{Name: "viewMatrix"},
		{Name: "modelMatrix"},

		{Name: "colorMap"},
		{Name: "positionMap"},
		{Name: "normalMap"},

		{Name: "light.Position"},
		{Name: "light.Color"},
		{Name: "light.AmbientIntensity"},
		{Name: "light.DiffuseIntensity"},

		{Name: "light.Range"},
		{Name: "light.Falloff"},

		{Name: "light.Direction"},
		{Name: "light.Angle"},

		{Name: "camPosition"},
		{Name: "screenSize"},
	}, []ShaderAttribute{
		{Name: "vertexPosition", Stride: 3, Typ: gl.FLOAT},
	})
	defer lightTestProgram.Delete()

	// rendertarget
	tw, th := 1024, 1024
	colorTexture, positionTexture, normalTexture, frameBuffer := GenerateMRT(tw, th)

	// mesh buffer
	mesh := GeneratePlane(10, 5)
	planeMesh := NewMeshBuffer([]MeshBufferAttribute{
		{Name: "position", Target: gl.ARRAY_BUFFER, Usage: gl.STATIC_DRAW},
		{Name: "uv", Target: gl.ARRAY_BUFFER, Usage: gl.STATIC_DRAW},
		{Name: "normal", Target: gl.ARRAY_BUFFER, Usage: gl.STATIC_DRAW},
		{Name: "index", Target: gl.ELEMENT_ARRAY_BUFFER, Usage: gl.STATIC_DRAW},
	})
	planeMesh.UpdateAttribute("position", mesh.Positions)
	planeMesh.UpdateAttribute("uv", mesh.UVs)
	planeMesh.UpdateAttribute("normal", mesh.Normals)
	planeMesh.UpdateAttribute("index", mesh.Indices)
	defer planeMesh.Delete()

	// cameras
	cameraPosition := mgl32.Vec3{0, 5, 10}
	geometryCamera := Camera{
		Projection: mgl32.Perspective(45.0, float32(tw)/float32(th), 1.0, 100.0),
		View:       mgl32.LookAtV(cameraPosition, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0}),
	}

	lightingCamera := Camera{
		Projection: mgl32.Perspective(45.0, float32(width)/float32(height), 1.0, 100.0),
		View:       mgl32.LookAtV(mgl32.Vec3{0, 0, 5}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0}), // mgl32.Ident4()
	}

	// create objects
	var objects []Renderable
	var x, y, z, a float32
	for x = -4; x <= 4; x += 4 {
		for z = 4; z >= -4; z -= 4 {
			a += math.Pi / 4.0

			g := cubeMesh
			t := gridTexture

			if len(objects)%2 == 0 {
				g = sphereMesh
				t = moonTexture
			}

			objects = append(objects, Renderable{
				Transform:       mgl32.Translate3D(x, y, z).Mul4(mgl32.HomogRotate3D(a, (mgl32.Vec3{1, 0.8, 0.5}).Normalize())),
				AngularVelocity: (mgl32.Vec3{1, 0.8, 0.5}).Normalize().Mul(a * float32(math.Pi/8.0)),
				Geometry:        g,
				Material: Material{
					Color:   mgl32.Vec3{1, 1, 1}, //(mgl32.Vec3{x, 1, z}).Normalize(),
					Texture: t,
					Opacity: 1.0,

					SpecularIntensity: 0.3,
					SpecularPower:     5.0,
				},
			})
		}
	}

	objects = append(objects, Renderable{
		Transform:       mgl32.Translate3D(0, -2, 0).Mul4(mgl32.LookAtV(mgl32.Vec3{0, -1, 0}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 0, 1})),
		AngularVelocity: mgl32.Vec3{},
		Geometry:        planeMesh,
		Material: Material{
			Color:   mgl32.Vec3{1, 1, 1},
			Texture: gridTexture,
			Opacity: 1.0,

			SpecularIntensity: 0.3,
			SpecularPower:     5.0,
		},
	}, Renderable{
		Transform:       mgl32.Translate3D(0, 0, -5).Mul4(mgl32.LookAtV(mgl32.Vec3{0, 0, 1}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})),
		AngularVelocity: mgl32.Vec3{},
		Geometry:        planeMesh,
		Material: Material{
			Color:   mgl32.Vec3{1, 1, 1},
			Texture: gridTexture,
			Opacity: 1.0,

			SpecularIntensity: 0.3,
			SpecularPower:     5.0,
		},
	}, Renderable{
		Transform:       mgl32.Translate3D(-5, 0, 0).Mul4(mgl32.LookAtV(mgl32.Vec3{-1, 0, 0}, mgl32.Vec3{0, 0, 0}, mgl32.Vec3{0, 1, 0})),
		AngularVelocity: mgl32.Vec3{},
		Geometry:        planeMesh,
		Material: Material{
			Color:   mgl32.Vec3{1, 1, 1},
			Texture: gridTexture,
			Opacity: 1.0,

			SpecularIntensity: 0.3,
			SpecularPower:     5.0,
		},
	})

	// texture cache
	currentTextures := map[gl.Texture]int{}
	bindTexture := func(t gl.Texture) int {
		if slot, found := currentTextures[t]; found {
			return slot
		}

		slot := len(currentTextures)

		gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(slot))
		t.Bind(gl.TEXTURE_2D)

		currentTextures[t] = slot
		return slot
	}
	unbindTextures := func() {
		for t := range currentTextures {
			t.Unbind(gl.TEXTURE_2D)
			delete(currentTextures, t)
		}
	}

	// create lights
	lights := []SpotLight{{
		Position:         mgl32.Vec3{4, 4, 0},
		Color:            mgl32.Vec3{1, 1, 1},
		AmbientIntensity: 0.0,
		DiffuseIntensity: 1.0,

		Range:   10,
		Falloff: 0.5,

		Direction: (mgl32.Vec3{0, 0, 0}).Sub(mgl32.Vec3{4, 4, 0}).Normalize(),
		Angle:     mgl32.DegToRad(90.0),
	}, {
		Position:         mgl32.Vec3{-4, 2, 6},
		Color:            mgl32.Vec3{0.5, 0.5, 1},
		AmbientIntensity: 0.0,
		DiffuseIntensity: 1.0,

		Range:   13,
		Falloff: 1.0,

		Direction: (mgl32.Vec3{0, 0, -1}).Normalize(),
		Angle:     mgl32.DegToRad(45.0),
	}}

	mesh = GenerateSphere(1, 20, 10)
	pointLightBoundingMesh := NewMeshBuffer([]MeshBufferAttribute{
		{Name: "position", Target: gl.ARRAY_BUFFER, Usage: gl.STATIC_DRAW},
		{Name: "index", Target: gl.ELEMENT_ARRAY_BUFFER, Usage: gl.STATIC_DRAW},
	})
	pointLightBoundingMesh.UpdateAttribute("position", mesh.Positions)
	pointLightBoundingMesh.UpdateAttribute("index", mesh.Indices)
	defer pointLightBoundingMesh.Delete()

	mesh = GenerateCylinder(0, 1, 1, 8, 1, true)
	mesh.Transform(mgl32.HomogRotate3DX(-math.Pi / 2.0)) // pre-rotate
	spotLightBoundingMesh := NewMeshBuffer([]MeshBufferAttribute{
		{Name: "position", Target: gl.ARRAY_BUFFER, Usage: gl.STATIC_DRAW},
		{Name: "uv", Target: gl.ARRAY_BUFFER, Usage: gl.STATIC_DRAW},
		{Name: "normal", Target: gl.ARRAY_BUFFER, Usage: gl.STATIC_DRAW},
		{Name: "index", Target: gl.ELEMENT_ARRAY_BUFFER, Usage: gl.STATIC_DRAW},
	})
	spotLightBoundingMesh.UpdateAttribute("position", mesh.Positions)
	spotLightBoundingMesh.UpdateAttribute("uv", mesh.UVs)
	spotLightBoundingMesh.UpdateAttribute("normal", mesh.Normals)
	spotLightBoundingMesh.UpdateAttribute("index", mesh.Indices)
	defer spotLightBoundingMesh.Delete()

	// main loop
	var (
		lastTime    = time.Now()
		currentTime time.Time
		delta       time.Duration

		currentGeometry *MeshBuffer
	)
	for ok := true; ok; ok = (window.GetKey(glfw.KeyEscape) != glfw.Press && !window.ShouldClose()) {
		currentTime = time.Now()
		delta = currentTime.Sub(lastTime)
		lastTime = currentTime

		// update objects
		for i, o := range objects {
			v := o.AngularVelocity.Mul(float32(delta.Seconds()))
			r := mgl32.AnglesToQuat(v[0], v[1], v[2], mgl32.XYZ).Mat4()

			objects[i].Transform = o.Transform.Mul4(r)
		}

		// render to fbo, geometry calculation pass
		func() {
			frameBuffer.Bind()
			defer frameBuffer.Unbind()

			gl.DepthMask(true) // only geometry pass updates depth buffer
			defer gl.DepthMask(false)

			gl.Viewport(0, 0, tw, th)
			gl.ClearColor(0.0, 0.0, 0.0, 0.0)
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

			gl.Enable(gl.DEPTH_TEST)
			defer gl.Disable(gl.DEPTH_TEST)
			gl.Disable(gl.BLEND) // irrelevant in geometry pass

			defer unbindTextures()

			// use program
			geometryProgram.Use()
			defer geometryProgram.Unuse()

			// update camera uniforms
			geometryProgram.UpdateUniform("projectionMatrix", geometryCamera.Projection)
			geometryProgram.UpdateUniform("viewMatrix", geometryCamera.View)

			for _, o := range objects {
				// update material uniforms
				geometryProgram.UpdateUniform("Diffuse", o.Material.Color)
				//geometryProgram.UpdateUniform("opacity", o.Material.Opacity) // no opacity in geometry pass

				geometryProgram.UpdateUniform("SpecularIntensity", o.Material.SpecularIntensity)
				geometryProgram.UpdateUniform("SpecularPower", o.Material.SpecularPower)

				// update object uniforms
				geometryProgram.UpdateUniform("modelMatrix", o.Transform)

				modelViewMatrix := geometryCamera.View.Mul4(o.Transform)
				normalMatrix := mgl32.Mat4Normal(modelViewMatrix)
				geometryProgram.UpdateUniform("normalMatrix", normalMatrix)

				// bind textures
				geometryProgram.UpdateUniform("DiffuseMap", bindTexture(o.Material.Texture))

				// bind attributes
				if currentGeometry != o.Geometry {
					if currentGeometry != nil {
						currentGeometry.Unbind()
						geometryProgram.DisableAttributes()
					}
					currentGeometry = o.Geometry
					currentGeometry.Bind() // enable vao

					geometryProgram.BindAttribute("vertexPosition", currentGeometry, "position")
					geometryProgram.BindAttribute("vertexNormal", currentGeometry, "normal")
					geometryProgram.BindAttribute("vertexUV", currentGeometry, "uv")
				}

				// draw elements
				currentGeometry.WithAttribute("index", func(a MeshBufferAttribute) {
					gl.DrawElements(gl.TRIANGLES, a.Count, a.Typ, nil)
				})
			}
		}()

		currentGeometry.Unbind()
		currentGeometry = nil

		// render to screen, lighting calculation pass
		func() {
			gl.Enable(gl.BLEND)
			gl.BlendEquation(gl.FUNC_ADD)
			gl.BlendFunc(gl.ONE, gl.ONE)
			//gl.BlendEquationSeparate(gl.FUNC_ADD, gl.FUNC_ADD)
			//gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

			gl.Viewport(0, 0, width, height)
			gl.ClearColor(0.0, 0.0, 0.0, 1.0)
			gl.Clear(gl.COLOR_BUFFER_BIT)

			gl.Disable(gl.CULL_FACE) // light is not rendered when camera is inside of bounding sphere
			defer gl.Enable(gl.CULL_FACE)

			defer unbindTextures()

			// point light pass
			//   shadow as spotlights
			// directional light pass
			//   like spotlights but with ortho instead of perspective

			/*
				for each spotlight
					render shadowmap
					blend to screen with g-buffer and shadowmap textures
			*/

			// spot light testing
			func() {
				lightTestProgram.Use()
				defer lightTestProgram.Unuse()

				// update uniforms
				lightTestProgram.UpdateUniform("projectionMatrix", geometryCamera.Projection)
				lightTestProgram.UpdateUniform("viewMatrix", geometryCamera.View)
				lightTestProgram.UpdateUniform("camPosition", cameraPosition)

				lightTestProgram.UpdateUniform("screenSize", mgl32.Vec2{float32(width), float32(height)})

				// bind textures
				lightTestProgram.UpdateUniform("colorMap", bindTexture(colorTexture))
				lightTestProgram.UpdateUniform("positionMap", bindTexture(positionTexture))
				lightTestProgram.UpdateUniform("normalMap", bindTexture(normalTexture))

				for _, light := range lights {
					if distance := light.Position.Sub(cameraPosition).Len(); distance < light.Range {
						// camera inside of light volume
						// render back faces of the light volume (ignore stenciling)
					} else {
						// camera outside of light volume
						// render front faces of the light volume and mask them out in the stencil buffer
						// render back faces of the light volume comparing to stencil
					}

					scaleLength, scaleWidth := light.Range, light.Range*float32(math.Tan(float64(light.Angle/2)))
					modelMatrix := mgl32.Translate3D(light.Position[0], light.Position[1], light.Position[2]).
						Mul4(mgl32.LookAtV(mgl32.Vec3{}, light.Direction, mgl32.Vec3{0, 1, 0})).
						Mul4(mgl32.Scale3D(scaleWidth, scaleWidth, scaleLength))
					lightTestProgram.UpdateUniform("modelMatrix", modelMatrix)

					lightTestProgram.UpdateUniform("light.Position", light.Position)
					lightTestProgram.UpdateUniform("light.Color", light.Color)
					lightTestProgram.UpdateUniform("light.AmbientIntensity", light.AmbientIntensity)
					lightTestProgram.UpdateUniform("light.DiffuseIntensity", light.DiffuseIntensity)

					lightTestProgram.UpdateUniform("light.Range", light.Range)
					lightTestProgram.UpdateUniform("light.Falloff", light.Falloff)

					lightTestProgram.UpdateUniform("light.Direction", light.Direction)
					lightTestProgram.UpdateUniform("light.Angle", math.Cos(float64(light.Angle/2)))

					// bind attributes
					if currentGeometry != spotLightBoundingMesh {
						if currentGeometry != nil {
							currentGeometry.Unbind()
							lightTestProgram.DisableAttributes()
						}
						currentGeometry = spotLightBoundingMesh
						currentGeometry.Bind()

						lightTestProgram.BindAttribute("vertexPosition", currentGeometry, "position")
					}

					// draw elements
					currentGeometry.WithAttribute("index", func(a MeshBufferAttribute) {
						gl.DrawElements(gl.TRIANGLES, a.Count, a.Typ, nil)
					})
				}
			}()

			return

			func() {
				// use program
				lightingProgram.Use()
				defer lightingProgram.Unuse()

				// update uniforms
				lightingProgram.UpdateUniform("projectionMatrix", lightingCamera.Projection)
				lightingProgram.UpdateUniform("viewMatrix", lightingCamera.View)

				lightingProgram.UpdateUniform("modelMatrix", mgl32.Ident4())

				// bind textures
				lightingProgram.UpdateUniform("colorMap", bindTexture(colorTexture))
				lightingProgram.UpdateUniform("positionMap", bindTexture(positionTexture))
				lightingProgram.UpdateUniform("normalMap", bindTexture(normalTexture))

				// bind attributes
				if currentGeometry != planeMesh {
					if currentGeometry != nil {
						currentGeometry.Unbind()
						lightingProgram.DisableAttributes()
					}
					currentGeometry = planeMesh
					currentGeometry.Bind()

					lightingProgram.BindAttribute("vertexPosition", currentGeometry, "position")
					lightingProgram.BindAttribute("vertexUV", currentGeometry, "uv")
				}

				// draw elements
				currentGeometry.WithAttribute("index", func(a MeshBufferAttribute) {
					gl.DrawElements(gl.TRIANGLES, a.Count, a.Typ, nil)
				})
			}()
		}()

		// Swap buffers
		window.SwapBuffers()
		glfw.PollEvents()
	}
}

type Material struct {
	Color   mgl32.Vec3
	Texture gl.Texture

	SpecularIntensity float32 // shininess factor
	SpecularPower     float32 //

	Opacity float32 // separate forward rendering pass if < 1.0

	Billboard, Wireframe,
	CastShadow, ReceiveShadow bool
}

type Renderable struct {
	Transform       mgl32.Mat4
	AngularVelocity mgl32.Vec3
	Geometry        *MeshBuffer
	Material        Material
}

type BaseLight struct {
	Color            mgl32.Vec3
	AmbientIntensity float32
	DiffuseIntensity float32

	CastShadow bool
}

type DirectionalLight struct {
	Direction mgl32.Vec3

	//BaseLight
	Color            mgl32.Vec3
	AmbientIntensity float32
	DiffuseIntensity float32

	CastShadow bool
}

type PointLight struct {
	Position    mgl32.Vec3
	Attenuation struct {
		Constant,
		Linear,
		Quadratic float32
	}

	//BaseLight
	Color            mgl32.Vec3
	AmbientIntensity float32
	DiffuseIntensity float32

	CastShadow bool
}

type SpotLight struct {
	// base
	Position         mgl32.Vec3
	Color            mgl32.Vec3
	AmbientIntensity float32
	DiffuseIntensity float32

	// attenuation
	Range   float32 // 100% light range
	Falloff float32 // f=1 - 50% light at 1/2 range, f<1 - moves 50% toward range, f>1 - toward light position (0 distance)

	// spot light
	Direction mgl32.Vec3
	Angle     float32 //  maximum angle between the light direction and the light to pixel vector

	// shadows
	CastShadow bool
}

type Camera struct {
	Projection, View mgl32.Mat4
}

type ShaderUniform struct {
	Name string
	//Typ  gl.GLenum // Vec3, Mat4, Texture ...

	location gl.UniformLocation
}

type ShaderAttribute struct {
	Name   string
	Stride uint      // Vec3 = 3, Vec2 = 2 ...
	Typ    gl.GLenum // gl.FLOAT, gl.UNSIGNED_SJORT ...
	// generate like in MeshBufferAttribute?

	location gl.AttribLocation
	enabled  bool
}

type ShaderProgram struct {
	program    gl.Program
	uniforms   map[string]ShaderUniform
	attributes map[string]ShaderAttribute
}

func NewShaderProgram(p gl.Program, uniforms []ShaderUniform, attributes []ShaderAttribute) *ShaderProgram {
	sp := &ShaderProgram{
		program:    p,
		uniforms:   map[string]ShaderUniform{},
		attributes: map[string]ShaderAttribute{},
	}

	for _, u := range uniforms {
		u.location = p.GetUniformLocation(u.Name)
		sp.uniforms[u.Name] = u
	}

	for _, a := range attributes {
		a.location = p.GetAttribLocation(a.Name)
		sp.attributes[a.Name] = a
	}

	return sp
}

func (sp *ShaderProgram) Delete() {
	sp.program.Delete()
}

func (sp *ShaderProgram) Use() {
	sp.program.Use()
}

func (sp *ShaderProgram) Unuse() {
	sp.program.Unuse()
}

func (sp *ShaderProgram) UpdateUniform(name string, value interface{}) {
	if _, found := sp.uniforms[name]; !found {
		log.Fatalf("unsupported uniform: %v", name)
	}

	switch t := value.(type) {
	default:
		log.Fatalf("%v has unknown type: %T", name, t)

	case int:
		sp.uniforms[name].location.Uniform1i(t)
	case float32:
		sp.uniforms[name].location.Uniform1f(t)
	case float64:
		sp.uniforms[name].location.Uniform1f(float32(t))

	case mgl32.Mat3:
		sp.uniforms[name].location.UniformMatrix3fv(false, t)
	case mgl32.Mat4:
		sp.uniforms[name].location.UniformMatrix4fv(false, t)

	case mgl32.Vec2:
		sp.uniforms[name].location.Uniform2f(t[0], t[1])
	case mgl32.Vec3:
		sp.uniforms[name].location.Uniform3f(t[0], t[1], t[2])
	case mgl32.Vec4:
		sp.uniforms[name].location.Uniform4f(t[0], t[1], t[2], t[3])

	case bool:
		if t {
			sp.uniforms[name].location.Uniform1i(1)
		} else {
			sp.uniforms[name].location.Uniform1i(0)
		}
	}
}

func (sp *ShaderProgram) BindAttribute(name string, buffer *MeshBuffer, bufname string) {
	a, found := sp.attributes[name]
	if !found {
		log.Fatalf("unsupported attribute: %v", name)
	}

	// enable vao access to attribute, disabled by sp.Unuse()
	if !a.enabled {
		a.location.EnableArray()
		a.enabled = true
		sp.attributes[name] = a
	}

	buffer.WithAttribute(bufname, func(_ MeshBufferAttribute) {
		// vertex attribute will get data from currently bound buffer
		a.location.AttribPointer(a.Stride, a.Typ, false, 0, nil)
	})
}

func (sp *ShaderProgram) DisableAttributes() {
	for n, a := range sp.attributes {
		if a.enabled {
			a.location.DisableArray()
			a.enabled = false

			sp.attributes[n] = a
		}
	}
}

func LoadShaderFile(vertex, fragment string) gl.Program {
	// load vertex shader
	vdata, err := ioutil.ReadFile(vertex)
	if err != nil {
		log.Fatal("unknown vertex file: ", vertex)
	}

	// load fragment shader
	fdata, err := ioutil.ReadFile(fragment)
	if err != nil {
		log.Fatal("unknown fragment file: ", fragment)
	}

	return LoadShader(string(vdata), string(fdata))
}

func LoadShader(vertex, fragment string) gl.Program {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal("load shader panic: ", err)
		}
	}()

	program := gl.CreateProgram()

	// vertex shader
	vshader := gl.CreateShader(gl.VERTEX_SHADER)
	vshader.Source(vertex)
	vshader.Compile()
	if vshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		log.Fatalf("vertex shader error: %v", vshader.GetInfoLog())
	}
	defer vshader.Delete()

	// fragment shader
	fshader := gl.CreateShader(gl.FRAGMENT_SHADER)
	fshader.Source(fragment)
	fshader.Compile()
	if fshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		log.Fatalf("fragment shader error: %v", fshader.GetInfoLog())
	}
	defer fshader.Delete()

	// program
	program.AttachShader(vshader)
	program.AttachShader(fshader)
	program.Link()
	if program.Get(gl.LINK_STATUS) != gl.TRUE {
		log.Fatalf("linker error: %v", program.GetInfoLog())
	}

	return program
}

type MeshBufferAttribute struct {
	Name   string
	Target gl.GLenum // gl.ARRAY_BUFFER, gl.ELEMENT_ARRAY_BUFFER
	Usage  gl.GLenum // gl.STATIC_DRAW, gl.DYNAMIC_DRAW, gl.STREAM_DRAW ...

	Stride uint // overwritten at update
	Typ    gl.GLenum
	Count  int // current length of data
	buffer gl.Buffer
}

type MeshBuffer struct {
	vao     gl.VertexArray
	buffers map[string]MeshBufferAttribute
}

func NewMeshBuffer(buffers []MeshBufferAttribute) *MeshBuffer {
	mb := &MeshBuffer{
		vao:     gl.GenVertexArray(),
		buffers: map[string]MeshBufferAttribute{},
	}

	for _, b := range buffers {
		b.buffer = gl.GenBuffer()
		mb.buffers[b.Name] = b
	}

	return mb
}

func MeshBufferFromMesh(mesh *Mesh) *MeshBuffer {
	mb := NewMeshBuffer([]MeshBufferAttribute{
		{Name: "position", Target: gl.ARRAY_BUFFER, Usage: gl.STATIC_DRAW},
		{Name: "uv", Target: gl.ARRAY_BUFFER, Usage: gl.STATIC_DRAW},
		{Name: "normal", Target: gl.ARRAY_BUFFER, Usage: gl.STATIC_DRAW},
		{Name: "index", Target: gl.ELEMENT_ARRAY_BUFFER, Usage: gl.STATIC_DRAW},
	})

	mb.UpdateAttribute("position", mesh.Positions)
	mb.UpdateAttribute("uv", mesh.UVs)
	mb.UpdateAttribute("normal", mesh.Normals)
	mb.UpdateAttribute("index", mesh.Indices)

	return mb
}

func (mb *MeshBuffer) Delete() {
	mb.vao.Delete()

	for _, a := range mb.buffers {
		a.buffer.Delete()
	}
}

func (mb *MeshBuffer) Bind() {
	mb.vao.Bind()
}

func (mb *MeshBuffer) Unbind() {
	mb.vao.Unbind()
}

func (mb *MeshBuffer) WithAttribute(name string, do func(a MeshBufferAttribute)) {
	a, found := mb.buffers[name]
	if !found {
		log.Fatalf("unsupported attribute: %v", name)
	}

	// independent of vao state?
	a.buffer.Bind(a.Target)
	defer a.buffer.Unbind(a.Target)

	do(a)
}

func (mb *MeshBuffer) UpdateAttribute(name string, value interface{}) {
	a, found := mb.buffers[name]
	if !found {
		log.Fatalf("unsupported attribute: %v", name)
	}

	var count, stride int
	var typ gl.GLenum
	switch t := value.(type) {
	default:
		log.Fatalf("%v has unknown type: %T", name, t)

	case []uint16:
		count = len(t)
		stride = 1
		typ = gl.UNSIGNED_SHORT
	case []mgl32.Vec2:
		count = len(t)
		stride = 2
		typ = gl.FLOAT
	case []mgl32.Vec3:
		count = len(t)
		stride = 3
		typ = gl.FLOAT
	}

	mb.Bind()
	defer mb.Unbind()

	a.buffer.Bind(a.Target)
	defer a.buffer.Unbind(a.Target)

	// consider using gl.BufferSubData when replacing the entire data
	gl.BufferData(a.Target, count*stride*int(glh.Sizeof(typ)), value, a.Usage)

	a.Count = count
	a.Stride = uint(stride)
	a.Typ = typ
	mb.buffers[name] = a
}

type Mesh struct {
	Indices   []uint16
	Positions []mgl32.Vec3
	UVs       []mgl32.Vec2
	Normals   []mgl32.Vec3
}

func (m Mesh) Transform(mat mgl32.Mat4) {
	normal := mgl32.Mat4Normal(mat)

	for i, p := range m.Positions {
		m.Positions[i] = mat.Mul4x1(p.Vec4(1.0)).Vec3()
	}

	for i, n := range m.Normals {
		m.Normals[i] = normal.Mul3x1(n)
	}
}

type Vertex struct {
	position mgl32.Vec3
	uv       mgl32.Vec2
	normal   mgl32.Vec3
}

func (v Vertex) Key(precision int) string {
	return fmt.Sprintf("%v_%v_%v_%v_%v_%v_%v_%v",
		mgl32.Round(v.position[0], precision),
		mgl32.Round(v.position[1], precision),
		mgl32.Round(v.position[2], precision),

		mgl32.Round(v.normal[0], precision),
		mgl32.Round(v.normal[1], precision),
		mgl32.Round(v.normal[2], precision),

		mgl32.Round(v.uv[0], precision),
		mgl32.Round(v.uv[1], precision),
	)
}

type Face struct {
	A, B, C int
}

func LoadObjFile(path string) *Mesh {
	// open object file, init reader
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)

	// cache
	var (
		positions []mgl32.Vec3
		uvs       []mgl32.Vec2
		normals   []mgl32.Vec3
		vertices  []Vertex
		faces     []Face
	)

	// helpers
	mustFloat32 := func(v string) float32 {
		f, err := strconv.ParseFloat(v, 32)
		if err != nil {
			log.Fatal(err)
		}
		return float32(f)
	}
	mustUint64 := func(v string) uint64 {
		u, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			log.Fatal(err)
		}
		return u
	}

	// parse obj file
	for {
		if line, err := reader.ReadString('\n'); err == nil {
			fields := strings.Split(strings.TrimSpace(line), " ")

			switch strings.ToLower(fields[0]) {
			case "v": // geometric vertices: x, y, z, [w]
				positions = append(positions, mgl32.Vec3{
					mustFloat32(fields[1]),
					mustFloat32(fields[2]),
					mustFloat32(fields[3]),
				})

			case "vt": // texture vertices: u, v, [w]
				uvs = append(uvs, mgl32.Vec2{
					mustFloat32(fields[1]),
					1.0 - mustFloat32(fields[2]),
				})

			case "vn": // vertex normals: i, j, k
				normals = append(normals, mgl32.Vec3{
					mustFloat32(fields[1]),
					mustFloat32(fields[2]),
					mustFloat32(fields[3]),
				})

			case "f": // face: v/vt/vn v/vt/vn v/vt/vn

				// quad instead of tri, split up
				// f v/vt/vn v/vt/vn v/vt/vn v/vt/vn
				var fcs [][]string
				if len(fields) == 5 {
					fcs = [][]string{
						[]string{"f", fields[1], fields[2], fields[4]},
						[]string{"f", fields[2], fields[3], fields[4]},
					}
				} else {
					fcs = [][]string{fields}
				}

				for _, fields := range fcs {
					face := make([]Vertex, 3)

					// v/vt/vn
					for i, f := range fields[1:4] {
						a := strings.Split(f, "/")

						// vertex
						face[i].position = positions[mustUint64(a[0])-1]

						// uv
						if len(a) > 1 && a[1] != "" {
							face[i].uv = uvs[mustUint64(a[1])-1]
						}

						// normal
						if len(a) == 3 {
							face[i].normal = normals[mustUint64(a[2])-1]
						}
					}

					offset := len(vertices)
					vertices = append(vertices, face...)
					faces = append(faces, Face{offset, offset + 1, offset + 2})
				}
			default:
				// ignore
			}
		} else if err == io.EOF {
			break
		} else {
			log.Fatal(err)
		}
	}

	// search and mark duplicate vertices
	lookup := map[string]int{}
	unique := []Vertex{}
	changed := map[int]int{}
	for i, v := range vertices {
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
	for _, f := range faces {
		a, b, c := changed[f.A], changed[f.B], changed[f.C]
		if a == b || b == c || c == a {
			// degenerated face, remove
			continue
		}

		cleaned = append(cleaned, Face{a, b, c})
	}

	// copy values to buffers
	n := len(unique)
	m := Mesh{
		Indices:   make([]uint16, len(cleaned)*3),
		Positions: make([]mgl32.Vec3, n),
		UVs:       make([]mgl32.Vec2, n),
		Normals:   make([]mgl32.Vec3, n),
	}

	for i, v := range unique {
		m.Positions[i] = v.position
		m.UVs[i] = v.uv
		m.Normals[i] = v.normal
	}

	for i, f := range cleaned {
		m.Indices[i*3] = uint16(f.A)
		m.Indices[i*3+1] = uint16(f.B)
		m.Indices[i*3+2] = uint16(f.C)
	}

	return &m
}

func GeneratePlane(width, height float32) *Mesh {
	// dimensions
	halfWidth := width / 2.0
	halfHeight := height / 2.0

	// positions
	a := mgl32.Vec3{halfWidth, halfHeight, 0}
	b := mgl32.Vec3{-halfWidth, halfHeight, 0}
	c := mgl32.Vec3{-halfWidth, -halfHeight, 0}
	d := mgl32.Vec3{halfWidth, -halfHeight, 0}

	// uvs
	tl := mgl32.Vec2{0, 1}
	tr := mgl32.Vec2{1, 1}
	bl := mgl32.Vec2{0, 0}
	br := mgl32.Vec2{1, 0}

	// normals
	n := mgl32.Vec3{0, 0, 1}

	// copy values to buffers
	return &Mesh{
		Indices:   []uint16{0, 1, 2, 2, 3, 0},
		Positions: []mgl32.Vec3{a, b, c, d},
		UVs:       []mgl32.Vec2{tr, tl, bl, br},
		Normals:   []mgl32.Vec3{n, n, n, n},
	}
}

func GenerateCylinder(radiusTop, radiusBottom, height float64, radialSegments, heightSegments int, open bool) *Mesh {
	// dimensions
	if radialSegments < 3 {
		radialSegments = 3
	}
	if heightSegments < 1 {
		heightSegments = 1
	}

	halfHeight := height / 2.0
	tanTheta := (radiusBottom - radiusTop) / height

	// cache
	var (
		positions [][]mgl32.Vec3
		uvs       [][]mgl32.Vec2

		vertices []Vertex
		faces    []Face
	)

	// create cylindrical positions
	for y := 0; y <= heightSegments; y++ {
		var positionsRow []mgl32.Vec3
		var uvsRow []mgl32.Vec2

		v := float64(y) / float64(heightSegments)
		radius := v*(radiusBottom-radiusTop) + radiusTop

		for x := 0; x <= radialSegments; x++ {
			u := float64(x) / float64(radialSegments)
			s, c := math.Sincos(u * math.Pi * 2)

			position := mgl32.Vec3{
				float32(radius * s),
				float32(-v*height + halfHeight),
				float32(radius * c),
			}

			positionsRow = append(positionsRow, position)
			uvsRow = append(uvsRow, mgl32.Vec2{
				float32(u),
				float32(1.0 - v),
			})
		}

		positions = append(positions, positionsRow)
		uvs = append(uvs, uvsRow)
	}

	// combine positions to faces
	for x := 0; x < radialSegments; x++ {
		var na, nb mgl32.Vec3
		if radiusTop != 0 {
			na = positions[0][x]
			nb = positions[0][x+1]
		} else {
			na = positions[1][x]
			nb = positions[1][x+1]
		}

		// normals
		na[1] = float32(math.Sqrt(float64(na[0]*na[0]+na[2]*na[2])) * tanTheta)
		na = na.Normalize()
		nb[1] = float32(math.Sqrt(float64(nb[0]*nb[0]+nb[2]*nb[2])) * tanTheta)
		nb = nb.Normalize()

		for y := 0; y < heightSegments; y++ {
			// positions
			v1 := positions[y][x]
			v2 := positions[y+1][x]
			v3 := positions[y+1][x+1]
			v4 := positions[y][x+1]

			// uvs
			uv1 := uvs[y][x]
			uv2 := uvs[y+1][x]
			uv3 := uvs[y+1][x+1]
			uv4 := uvs[y][x+1]

			offset := len(vertices)
			vertices = append(vertices,
				Vertex{
					position: v1,
					normal:   na,
					uv:       uv1,
				}, Vertex{
					position: v2,
					normal:   na,
					uv:       uv2,
				}, Vertex{
					position: v4,
					normal:   nb,
					uv:       uv4,
				},
				Vertex{
					position: v2,
					normal:   na,
					uv:       uv2,
				}, Vertex{
					position: v3,
					normal:   nb,
					uv:       uv3,
				}, Vertex{
					position: v4,
					normal:   nb,
					uv:       uv4,
				})
			faces = append(faces,
				Face{offset, offset + 1, offset + 2},
				Face{offset + 3, offset + 4, offset + 5},
			)
		}
	}

	// top
	if !open && radiusTop > 0 {
		// center
		c := mgl32.Vec3{0, float32(halfHeight), 0}

		// normal
		n := mgl32.Vec3{0, 1, 0}

		for x := 0; x < radialSegments; x++ {
			// positions
			v1 := positions[0][x]
			v2 := positions[0][x+1]

			// uvs
			uv1 := uvs[0][x]
			uv2 := uvs[0][x+1]
			uv3 := mgl32.Vec2{uv2[0], 0}

			offset := len(vertices)
			vertices = append(vertices,
				Vertex{
					position: v1,
					normal:   n,
					uv:       uv1,
				}, Vertex{
					position: v2,
					normal:   n,
					uv:       uv2,
				}, Vertex{
					position: c,
					normal:   n,
					uv:       uv3,
				})
			faces = append(faces, Face{offset, offset + 1, offset + 2})
		}
	}

	// bottom
	if !open && radiusBottom > 0 {
		// last position ring
		y := heightSegments

		// center
		c := mgl32.Vec3{0, float32(-halfHeight), 0}

		// normal
		n := mgl32.Vec3{0, -1, 0}

		for x := 0; x < radialSegments; x++ {
			// positions
			v1 := positions[y][x+1]
			v2 := positions[y][x]

			// uvs
			uv1 := uvs[y][x+1]
			uv2 := uvs[y][x]
			uv3 := mgl32.Vec2{uv2[0], 1}

			offset := len(vertices)
			vertices = append(vertices,
				Vertex{
					position: v1,
					normal:   n,
					uv:       uv1,
				}, Vertex{
					position: v2,
					normal:   n,
					uv:       uv2,
				}, Vertex{
					position: c,
					normal:   n,
					uv:       uv3,
				})
			faces = append(faces, Face{offset, offset + 1, offset + 2})
		}
	}

	// search and mark duplicate vertices
	lookup := map[string]int{}
	unique := []Vertex{}
	changed := map[int]int{}
	for i, v := range vertices {
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
	for _, f := range faces {
		a, b, c := changed[f.A], changed[f.B], changed[f.C]
		if a == b || b == c || c == a {
			// degenerated face, remove
			continue
		}

		cleaned = append(cleaned, Face{a, b, c})
	}

	// copy values to buffers
	n := len(unique)
	m := Mesh{
		Indices:   make([]uint16, len(cleaned)*3),
		Positions: make([]mgl32.Vec3, n),
		UVs:       make([]mgl32.Vec2, n),
		Normals:   make([]mgl32.Vec3, n),
	}

	for i, v := range unique {
		m.Positions[i] = v.position
		m.UVs[i] = v.uv
		m.Normals[i] = v.normal
	}

	for i, f := range cleaned {
		m.Indices[i*3] = uint16(f.A)
		m.Indices[i*3+1] = uint16(f.B)
		m.Indices[i*3+2] = uint16(f.C)
	}

	return &m
}

func GenerateSphere(radius float64, widthSegments, heightSegments int) *Mesh {
	// dimensions
	if widthSegments < 3 {
		widthSegments = 3
	}
	if heightSegments < 2 {
		heightSegments = 2
	}

	phiStart, phiLength := 0.0, math.Pi*2
	thetaStart, thetaLength := 0.0, math.Pi

	// cache
	var (
		positions [][]mgl32.Vec3
		uvs       [][]mgl32.Vec2
		vertices  []Vertex
		faces     []Face
	)

	for y := 0; y <= heightSegments; y++ {
		var positionsRow []mgl32.Vec3
		var uvsRow []mgl32.Vec2

		for x := 0; x <= widthSegments; x++ {
			u := float32(x) / float32(widthSegments)
			v := float32(y) / float32(heightSegments)

			position := mgl32.Vec3{
				float32(-radius * math.Cos(phiStart+float64(u)*phiLength) * math.Sin(thetaStart+float64(v)*thetaLength)),
				float32(radius * math.Cos(thetaStart+float64(v)*thetaLength)),
				float32(radius * math.Sin(phiStart+float64(u)*phiLength) * math.Sin(thetaStart+float64(v)*thetaLength)),
			}

			positionsRow = append(positionsRow, position)
			uvsRow = append(uvsRow, mgl32.Vec2{u, 1.0 - v})
		}

		positions = append(positions, positionsRow)
		uvs = append(uvs, uvsRow)
	}

	for y := 0; y < heightSegments; y++ {
		for x := 0; x < widthSegments; x++ {
			// positions
			v1 := positions[y][x+1]
			v2 := positions[y][x]
			v3 := positions[y+1][x]
			v4 := positions[y+1][x+1]

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

			if math.Abs(float64(v1[1])) == radius {
				offset := len(vertices)
				vertices = append(vertices,
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
				faces = append(faces, Face{offset, offset + 1, offset + 2})

			} else if math.Abs(float64(v3[1])) == radius {
				offset := len(vertices)
				vertices = append(vertices,
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
				faces = append(faces, Face{offset, offset + 1, offset + 2})

			} else {
				offset := len(vertices)
				vertices = append(vertices,
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
					}, Vertex{
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
				faces = append(faces,
					Face{offset, offset + 1, offset + 2},
					Face{offset + 3, offset + 4, offset + 5},
				)
			}
		}
	}

	// search and mark duplicate vertices
	lookup := map[string]int{}
	unique := []Vertex{}
	changed := map[int]int{}
	for i, v := range vertices {
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
	for _, f := range faces {
		a, b, c := changed[f.A], changed[f.B], changed[f.C]
		if a == b || b == c || c == a {
			// degenerated face, remove
			continue
		}

		cleaned = append(cleaned, Face{a, b, c})
	}

	// copy values to buffers
	n := len(unique)
	m := Mesh{
		Indices:   make([]uint16, len(cleaned)*3),
		Positions: make([]mgl32.Vec3, n),
		UVs:       make([]mgl32.Vec2, n),
		Normals:   make([]mgl32.Vec3, n),
	}

	for i, v := range unique {
		m.Positions[i] = v.position
		m.UVs[i] = v.uv
		m.Normals[i] = v.normal
	}

	for i, f := range cleaned {
		m.Indices[i*3] = uint16(f.A)
		m.Indices[i*3+1] = uint16(f.B)
		m.Indices[i*3+2] = uint16(f.C)
	}

	return &m
}

func LoadTexture(path string) gl.Texture {
	// load file
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// decode image
	im, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	// convert to rgba
	rgba, ok := im.(*image.RGBA)
	if !ok {
		bounds := im.Bounds()
		rgba = image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, im, image.Pt(0, 0), draw.Src)
	}

	buffer := gl.GenTexture()
	buffer.Bind(gl.TEXTURE_2D)

	// set texture parameters
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE) // gl.REPEAT
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE) // gl.REPEAT
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR) // gl.LINEAR_MIPMAP_LINEAR
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// give image to opengl
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
		rgba.Bounds().Dx(), rgba.Bounds().Dy(),
		0, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)

	// generate mipmaps
	gl.GenerateMipmap(gl.TEXTURE_2D)

	buffer.Unbind(gl.TEXTURE_2D)

	return buffer
}

func GenerateMRT(w, h int) (color, position, normal gl.Texture, fbo gl.Framebuffer) {
	// generate framebuffer
	frameBuffer := gl.GenFramebuffer()
	frameBuffer.Bind()
	defer frameBuffer.Unbind()

	// generate color target
	colorTexture := gl.GenTexture()
	colorTexture.Bind(gl.TEXTURE_2D)
	defer colorTexture.Unbind(gl.TEXTURE_2D)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, w, h, 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, colorTexture, 0)

	// generate position target
	positionTexture := gl.GenTexture()
	positionTexture.Bind(gl.TEXTURE_2D)
	defer positionTexture.Unbind(gl.TEXTURE_2D)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, w, h, 0, gl.RGBA, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT1, gl.TEXTURE_2D, positionTexture, 0)

	// generate normal target
	normalTexture := gl.GenTexture()
	normalTexture.Bind(gl.TEXTURE_2D)
	defer normalTexture.Unbind(gl.TEXTURE_2D)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA16F, w, h, 0, gl.RGBA, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT2, gl.TEXTURE_2D, normalTexture, 0)

	// generate depth texture
	depthTexture := gl.GenTexture()
	depthTexture.Bind(gl.TEXTURE_2D)
	defer depthTexture.Unbind(gl.TEXTURE_2D)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.DEPTH_COMPONENT16, w, h, 0, gl.DEPTH_COMPONENT, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, depthTexture, 0)

	//instead of depth texture
	// generate depth buffer
	/*
		depthBuffer := gl.GenRenderbuffer()
		depthBuffer.Bind()
		defer depthBuffer.Unbind()
		gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT16, w, h)
		depthBuffer.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER)
	*/
	gl.DrawBuffers(3, []gl.GLenum{gl.COLOR_ATTACHMENT0, gl.COLOR_ATTACHMENT1, gl.COLOR_ATTACHMENT2})

	// check
	if e := gl.CheckFramebufferStatus(gl.FRAMEBUFFER); e != gl.FRAMEBUFFER_COMPLETE {
		log.Fatalf("could not initialize framebuffer: %x", e)
	}

	return colorTexture, positionTexture, normalTexture, frameBuffer
}
