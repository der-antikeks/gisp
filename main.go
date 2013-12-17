package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/der-antikeks/gisp/engine"
	"github.com/der-antikeks/gisp/math"

	// TODO: remove
	glfw "github.com/go-gl/glfw3"
)

var (
	renderer *engine.Renderer
	controls engine.Control
)

func main() {
	runtime.LockOSThread()
	var err error

	// renderer
	if renderer, err = engine.NewRenderer("gophers in space!", 640, 480); err != nil {
		log.Fatalf("could not create renderer: %v\n", err)
	}
	defer renderer.Unload()

	//renderer.SetClearColor(math.Color{0.2, 0.2, 0.23})
	renderer.SetKeyCallback(onKeyPress)
	renderer.SetMouseButtonCallback(onMouseButton)

	//targetA := engine.NewRenderTarget(1024, 768)
	//targetB := engine.NewRenderTarget(1024, 768)
	//targetC := engine.NewRenderTarget(1024, 768)
	//targetD := engine.NewRenderTarget(1024, 768)
	var scene *engine.Scene
	var camera engine.Camera
	var pass *engine.RenderPass

	// background
	scene, camera = generateBackground()
	renderer.AddPass(engine.NewRenderPass(scene, camera, nil))

	// blur shader
	/*
		hBlurMaterial, err := engine.NewMaterial("blur")
		if err != nil {
			log.Fatalf("could not load shader material: %v\n", err)
		}
		hBlurMaterial.SetUniform("diffuseMap", targetA)
		hBlurMaterial.SetUniform("size", 3.0/1024.0)
		hBlurMaterial.SetUniform("vertical", false)
		renderer.AddPass(engine.NewShaderPass(hBlurMaterial, targetB))

		vBlurMaterial, err := engine.NewMaterial("blur")
		if err != nil {
			log.Fatalf("could not load shader material: %v\n", err)
		}
		vBlurMaterial.SetUniform("diffuseMap", targetB)
		vBlurMaterial.SetUniform("size", 3.0/1024.0)
		vBlurMaterial.SetUniform("vertical", true)
		renderer.AddPass(engine.NewShaderPass(vBlurMaterial, nil))
	*/

	// test scene
	scene, camera = generateScene()
	pass = engine.NewRenderPass(scene, camera, nil)
	pass.SetClear(false)
	renderer.AddPass(pass)

	// blend shader
	/*
		blendMaterial, err := engine.NewMaterial("blend")
		if err != nil {
			log.Fatalf("could not load shader material: %v\n", err)
		}
		blendMaterial.SetUniform("diffuseMapA",targetA)
		blendMaterial.SetUniform("diffuseMapB",targetC)
		blendMaterial.SetUniform("ratio",0.5)
		renderer.AddPass(engine.NewShaderPass(blendMaterial, nil))
	*/

	// hud
	scene, camera = generateHud()
	pass = engine.NewRenderPass(scene, camera, nil)
	pass.SetClear(false)
	renderer.AddPass(pass)

	// main loop
	lastTime := time.Now()
	meanDelta := 0.0
	frames := 0

	renderTicker := time.Tick(time.Duration(1000/72) * time.Millisecond)
	worldUpdateTicker := time.Tick(time.Duration(5) * time.Second)
	worldCleanupTicker := time.Tick(time.Duration(10) * time.Second)

	for renderer.Running() {
		select {
		case <-worldUpdateTicker: //go updateWorld()
		case <-worldCleanupTicker: //go cleanupWorld()
		case <-renderTicker:
			//default: // override fps limiter

			// calc fps
			currentTime := time.Now()
			delta := currentTime.Sub(lastTime)
			lastTime = currentTime

			meanDelta = meanDelta*0.9 + delta.Seconds()*0.1

			if frames++; frames%50 == 0 {
				fmt.Println("fps: ", 1.0/meanDelta)
				frames = 0
			}

			// animate
			update(delta)

			// draw
			renderer.Render()
		}
	}
}

func onKeyPress(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
	switch key {
	case glfw.KeyEscape:
		renderer.Quit()
	}
}

func onMouseButton(button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if button == glfw.MouseButton2 {
		if action != glfw.Release {
			renderer.SetMouseVisible(false)
		} else {
			renderer.SetMouseVisible(true)
		}
	}
}

var (
	objectAngle        float64
	moon, rotatingCube engine.Object
)

func update(delta time.Duration) {
	objectAngle += 1 * delta.Seconds()

	if moon != nil {
		moon.SetRotation(math.QuaternionFromAxisAngle(math.Vector{0, 0.5, 0.5}, objectAngle))
	}

	if rotatingCube != nil {
		rotatingCube.SetRotation(math.QuaternionFromAxisAngle(math.Vector{0, 1, 0}, objectAngle))
	}

	if controls != nil {
		controls.Update(delta)
	}
}

func generateScene() (*engine.Scene, engine.Camera) {
	// camera
	camera := engine.NewPerspectiveCamera(45.0, 4.0/3.0, 0.1, 100.0)
	//camera := engine.NewOrthographicCamera(-10, 10, 10, -10, 0.1, 100.0)
	camera.SetPosition(math.Vector{0, 0, -10})
	camera.LookAt(math.Vector{0, 0, 0})

	// controls
	//controls = engine.NewFlyControl(camera)
	controls = engine.NewOrbitControl(camera)
	renderer.SetController(controls)
	renderer.SetMouseVisible(true)

	// object 1
	cube := engine.NewCubeGeometry(2)

	texture, err := engine.LoadTexture("assets/uvtemplate.png")
	if err != nil {
		log.Fatalf("could not load texture: %v\n", err)
	}

	opaque, err := engine.NewMaterial("basic")
	if err != nil {
		log.Fatalf("could not load shader material: %v\n", err)
	}
	opaque.SetUniform("diffuseMap", texture)
	opaque.SetUniform("diffuse", math.Color{0, 1, 1})
	opaque.SetUniform("opacity", 1.0)

	obj1 := engine.NewMesh(cube, opaque)

	// object 2
	transparent, err := engine.NewMaterial("basic")
	if err != nil {
		log.Fatalf("could not load shader material: %v\n", err)
	}
	transparent.SetUniform("diffuseMap", texture)
	transparent.SetUniform("diffuse", math.Color{1, 0, 1})
	transparent.SetUniform("opacity", 0.7)

	obj2 := engine.NewMesh(cube, transparent)
	obj2.SetPosition(math.Vector{-3, 0, 0})
	obj2.SetRotation(math.QuaternionFromAxisAngle(math.Vector{0, 0, 1}, math.Pi/4.0))

	// object 3
	obj3 := engine.NewMesh(cube, transparent)
	obj3.SetPosition(math.Vector{3, 0, 0})

	// object 4
	obj4 := engine.NewMesh(cube, opaque)
	obj4.SetPosition(math.Vector{0, 3, 0})

	// moon
	sphere := engine.NewSphereGeometry(3, 100, 50)

	moonTex, err := engine.LoadTexture("assets/planets/moon_1024.jpg")
	if err != nil {
		log.Fatalf("could not load texture: %v\n", err)
	}

	moonMat, err := engine.NewMaterial("phong")
	if err != nil {
		log.Fatalf("could not load shader material: %v\n", err)
	}
	moonMat.SetUniform("diffuseMap", moonTex)

	moon = engine.NewMesh(sphere, moonMat)
	moon.SetPosition(math.Vector{0, 0, 10})

	// plane
	plane := engine.NewMesh(engine.NewPlaneGeometry(10, 10), opaque)
	plane.SetRotation(math.QuaternionFromAxisAngle(math.Vector{1, 0, 0}, math.Pi))
	plane.SetPosition(math.Vector{5, 0, 10})

	// fighter
	fighter, err := engine.LoadObject("assets/fighter/fighter.obj", "")
	if err != nil {
		log.Fatalf("could not load object: %v\n", err)
	}
	fighter.SetPosition(math.Vector{-5, 0, -5})
	scale := 2.0 / 1.0
	fighter.SetScale(math.Vector{scale, scale, scale})
	fighter.SetRotation(math.QuaternionFromAxisAngle(math.Vector{0, 0, 1}, math.Pi/4))

	// rotating cube
	rotatingCube, err = engine.LoadObject("assets/cube/cube.obj", "")
	if err != nil {
		log.Fatalf("could not load object: %v\n", err)
	}

	rotatingCube.SetPosition(math.Vector{5, 0, -5})

	if orb, ok := controls.(*engine.OrbitControl); ok {
		orb.SetTarget(rotatingCube.Position())
	}

	// scene
	scene := engine.NewScene()
	scene.AddChild(obj1, obj2, obj3, obj4, moon, plane, fighter, rotatingCube)

	// late adding
	moon2 := engine.NewMesh(sphere, moonMat)
	moon2.SetPosition(math.Vector{-5, 0, 0})
	moon.AddChild(moon2)

	return scene, camera
}

func generateHud() (*engine.Scene, engine.Camera) {
	// camera
	camera := engine.NewOrthographicCamera(-100, 100, 100, -100, 0, 1)
	//camera.SetPosition(math.Vector{0, 0, -1})
	//camera.LookAt(math.Vector{0, 0, 0})

	// opague material
	texture, err := engine.LoadTexture("assets/uvtemplate.png")
	if err != nil {
		log.Fatalf("could not load texture: %v\n", err)
	}

	opaque, err := engine.NewMaterial("basic")
	if err != nil {
		log.Fatalf("could not load shader material: %v\n", err)
	}
	opaque.SetUniform("diffuseMap", texture)
	opaque.SetUniform("diffuse", math.Color{1, 0, 1})
	opaque.SetUniform("opacity", 1.0)

	// right plane
	planeR := engine.NewMesh(engine.NewPlaneGeometry(50, 50), opaque)
	planeR.SetPosition(math.Vector{75, -75, 0})

	// transparent
	transparent, err := engine.NewMaterial("basic")
	if err != nil {
		log.Fatalf("could not load shader material: %v\n", err)
	}
	transparent.SetUniform("diffuseMap", texture)
	transparent.SetUniform("diffuse", math.Color{1, 0, 1})
	transparent.SetUniform("opacity", 0.7)

	// left plane
	planeL := engine.NewMesh(engine.NewPlaneGeometry(50, 50), transparent)
	planeL.SetPosition(math.Vector{-75, -75, 0})

	// font
	font, err := engine.LoadFont("assets/luxisr.ttf")
	if err != nil {
		log.Fatalf("could not load font: %v\n", err)
	}
	fontMesh := font.Printf("Testing Font 012345679")
	fontMesh.SetPosition(math.Vector{-100, 90, 0})

	// scene
	scene := engine.NewScene()
	scene.AddChild(planeR, planeL, fontMesh)

	return scene, camera
}

func generateBackground() (*engine.Scene, engine.Camera) {
	// camera
	camera := engine.NewOrthographicCamera(-1, 1, 1, -1, 0, 1)

	// material
	texture, err := engine.LoadTexture("assets/uvtemplate.png")
	if err != nil {
		log.Fatalf("could not load texture: %v\n", err)
	}

	opaque, err := engine.NewMaterial("basic")
	if err != nil {
		log.Fatalf("could not load shader material: %v\n", err)
	}
	opaque.SetUniform("diffuseMap", texture)
	opaque.SetUniform("diffuse", math.Color{0.1, 0.1, 0.5})
	opaque.SetUniform("opacity", 1.0)

	// plane
	plane := engine.NewMesh(engine.NewPlaneGeometry(2, 2), opaque)
	plane.SetPosition(math.Vector{0, 0, -1})

	// scene
	scene := engine.NewScene()
	scene.AddChild(plane)

	return scene, camera
}
