package game

import (
	"log"
	m "math"
	"time"

	"github.com/der-antikeks/gisp/ecs"
	"github.com/der-antikeks/gisp/math"
)

type EntityManager struct {
	engine *ecs.Engine
}

func NewEntityManager(e *ecs.Engine) *EntityManager {
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

func (em *EntityManager) CreateMainMenu() {}

func (em *EntityManager) createCube() {
	// Transformation
	trans := Transformation{
		Position: math.Vector{-1, 2, 0},
		Rotation: math.QuaternionFromAxisAngle(math.Vector{1, 0.5, 0}, m.Pi/4.0),
		Scale:    math.Vector{1, 1, 1},
		Up:       math.Vector{0, 1, 0},
	}

	geo := em.getGeometry("cube")
	mat := em.getMaterial("basic")
	mat.Set("diffuse", math.Color{1, 0, 0})
	mat.Set("opacity", 0.8)

	// Entity
	cube := em.engine.Entity()
	if err := em.engine.Set(
		cube,
		trans, geo, mat,
	); err != nil {
		log.Fatal("could not create cube:", err)
	}
}

func (em *EntityManager) createSphere() {
	// Transformation
	trans := Transformation{
		Position: math.Vector{0, 0, 0},
		Rotation: math.Quaternion{0, 0, 0, 1},
		Scale:    math.Vector{1, 1, 1},
		Up:       math.Vector{0, 1, 0},
	}

	geo := em.getGeometry("sphere")
	mat := em.getMaterial("basic")
	mat.Set("diffuse", math.Color{0, 1, 0})

	// Entity
	sphere := em.engine.Entity()
	if err := em.engine.Set(
		sphere,
		trans, geo, mat,
	); err != nil {
		log.Fatal("could not create sphere:", err)
	}
}

func (em *EntityManager) getMaterial(id string) Material {
	mat := Material{
		Program:  id,
		uniforms: map[string]interface{}{},
		program:  GetShader(id),
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

func (em *EntityManager) CreatePerspectiveCamera(fov, aspect, near, far float64) ecs.Entity {
	t := Transformation{
		Position: math.Vector{0, 0, -10},
		//Rotation: math.Quaternion{},
		Scale: math.Vector{1, 1, 1},
		Up:    math.Vector{0, 1, 0},
	}
	t.Rotation = math.QuaternionFromRotationMatrix(math.LookAt(t.Position, math.Vector{0, 0, 0}, t.Up))

	c := em.engine.Entity()
	if err := em.engine.Set(
		c,
		Projection{
			Matrix: math.NewPerspectiveMatrix(fov, aspect, near, far),
		},
		t,
	); err != nil {
		log.Fatal("could not create camera:", err)
	}

	return c
}

func (em *EntityManager) CreateOrthographicCamera(left, right, top, bottom, near, far float64) ecs.Entity {
	t := Transformation{
		Position: math.Vector{0, 0, -10},
		//Rotation: math.Quaternion{},
		Scale: math.Vector{1, 1, 1},
		Up:    math.Vector{0, 1, 0},
	}
	t.Rotation = math.QuaternionFromRotationMatrix(math.LookAt(t.Position, math.Vector{0, 0, 0}, t.Up))

	c := em.engine.Entity()
	if err := em.engine.Set(
		c,
		Projection{ // TODO: top/bottom switched?
			Matrix: math.NewOrthoMatrix(left, right, bottom, top, near, far),
		},
		t,
	); err != nil {
		log.Fatal("could not create camera:", err)
	}

	return c
}
