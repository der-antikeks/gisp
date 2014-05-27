package game

import (
	"log"
	m "math"
	"math/rand"
	"time"

	"github.com/der-antikeks/gisp/math"
)

type EntityManager struct {
	engine *Engine
}

func NewEntityManager(e *Engine) *EntityManager {
	return &EntityManager{
		engine: e,
	}
}

func (em *EntityManager) Initalize() {
	s := em.engine.Entity()
	if err := em.engine.Set(
		s,
		GameStateComponent{"init", time.Now()},
	); err != nil {
		log.Fatal("could not initialize:", err)
	}
}

func (em *EntityManager) CreateSplashScreen() {
	em.createCube()
	em.createSphere()
}

func (em *EntityManager) CreateMainMenu() {
	for i := 0; i < 2000; i++ {
		em.createRndCube()
	}
}

func (em *EntityManager) createCube() Entity {
	// Transformation
	trans := Transformation{
		Position: math.Vector{0, 0, 0},
		Rotation: math.QuaternionFromAxisAngle(math.Vector{1, 0.5, 0}, m.Pi/4.0),
		Scale:    math.Vector{1, 1, 1},
		Up:       math.Vector{0, 1, 0},
	}

	// velocity, rate of change per second
	vel := Velocity{
		Velocity: math.Vector{0, 0, 0},
		Angular: math.Vector{
			45.0 * math.DEG2RAD,
			5.0 * math.DEG2RAD,
			0,
		},
	}

	// geometry
	geo := em.getGeometry("cube")

	// material
	mat := em.getMaterial("flat")
	mat.Set("lightPosition", math.Vector{5, 5, 0, 1})
	mat.Set("lightDiffuse", math.Color{1, 0, 0})
	mat.Set("opacity", 0.8)

	tex, err := LoadTexture("assets/cube/cube.png")
	if err != nil {
		log.Fatal("could not load texture:", err)
	}
	mat.Set("diffuseMap", tex)

	// scene
	stc := SceneTree{Name: "mainscene"}

	// Entity
	cube := em.engine.Entity()
	if err := em.engine.Set(
		cube,
		trans, geo, mat, vel, stc,
	); err != nil {
		log.Fatal("could not create cube:", err)
	}

	return cube
}

func (em *EntityManager) createRndCube() Entity {
	// helper
	r := func(min, max float64) float64 {
		return rand.Float64()*(max-min) + min
	}
	d := func() float64 {
		if rand.Intn(2) == 0 {
			return -1
		}
		return 1
	}

	// transformation
	trans := Transformation{
		Position: math.Vector{
			r(1, 100) * d(),
			r(1, 100) * d(),
			r(1, 100) * d(),
		},
		Rotation: math.QuaternionFromAxisAngle((math.Vector{
			rand.Float64(),
			rand.Float64(),
			rand.Float64(),
		}).Normalize(), r(0, m.Pi*2)),
		Scale: math.Vector{1, 1, 1},
		Up:    math.Vector{0, 1, 0},
	}

	// geometry
	geo := em.getGeometry("cube")

	// material
	mat := em.getMaterial("basic")
	mat.Set("diffuse", math.Color{
		r(0.5, 1),
		r(0.5, 1),
		r(0.5, 1),
	})
	mat.Set("opacity", r(0.2, 1))

	// scene
	stc := SceneTree{Name: "mainscene"}

	// Entity
	cube := em.engine.Entity()
	if err := em.engine.Set(
		cube,
		trans, geo, mat, stc,
	); err != nil {
		log.Fatal("could not create cube:", err)
	}

	return cube
}

func (em *EntityManager) createSphere() Entity {
	// Transformation
	trans := Transformation{
		Position: math.Vector{5, 0, 0},
		Rotation: math.Quaternion{0, 0, 0, 1},
		Scale:    math.Vector{1, 1, 1},
		Up:       math.Vector{0, 1, 0},
	}

	// geometry
	geo := em.getGeometry("sphere")

	// material
	mat := em.getMaterial("phong")

	tex, err := LoadTexture("assets/uvtemplate.png")
	if err != nil {
		log.Fatal("could not load texture:", err)
	}
	mat.Set("diffuseMap", tex)

	// scene
	stc := SceneTree{Name: "mainscene"}

	// Entity
	sphere := em.engine.Entity()
	if err := em.engine.Set(
		sphere,
		trans, geo, mat, stc,
	); err != nil {
		log.Fatal("could not create sphere:", err)
	}

	return sphere
}

func (em *EntityManager) getMaterial(id string) Material {
	mat := Material{
		Program:  id,
		uniforms: map[string]interface{}{},
		program:  GetShader(id),
	}

	// preset with standard values
	for n, v := range mat.program.uniforms {
		if v.standard != nil {
			mat.uniforms[n] = v.standard
		}
	}

	return mat
}

func (em *EntityManager) getGeometry(id string) Geometry {
	mb := GetMeshBuffer(id)

	geo := Geometry{
		File:     id,
		mesh:     mb,
		Bounding: mb.Bounding,
	}

	return geo
}

func (em *EntityManager) CreatePerspectiveCamera(fov, aspect, near, far float64) Entity {
	t := Transformation{
		Position: math.Vector{0, 0, -10},
		//Rotation: math.Quaternion{},
		Scale: math.Vector{1, 1, 1},
		Up:    math.Vector{0, 1, 0},
	}
	t.Rotation = math.QuaternionLookAt(t.Position, math.Vector{0, 0, 0}, t.Up)

	c := em.engine.Entity()
	if err := em.engine.Set(
		c,
		Projection{
			Matrix: math.NewPerspectiveMatrix(fov, aspect, near, far),
		},
		t,
		SceneTree{Name: "mainscene"},
	); err != nil {
		log.Fatal("could not create camera:", err)
	}

	return c
}

func (em *EntityManager) CreateOrthographicCamera(left, right, top, bottom, near, far float64) Entity {
	t := Transformation{
		Position: math.Vector{0, 0, -10},
		//Rotation: math.Quaternion{},
		Scale: math.Vector{1, 1, 1},
		Up:    math.Vector{0, 1, 0},
	}
	t.Rotation = math.QuaternionLookAt(t.Position, math.Vector{0, 0, 0}, t.Up)

	c := em.engine.Entity()
	if err := em.engine.Set(
		c,
		Projection{ // TODO: top/bottom switched?
			Matrix: math.NewOrthoMatrix(left, right, bottom, top, near, far),
		},
		t,
		SceneTree{Name: "mainscene"},
	); err != nil {
		log.Fatal("could not create camera:", err)
	}

	return c
}
