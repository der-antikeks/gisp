package game

import (
	"errors"
	"log"
	m "math"
	"math/rand"
	"sync"

	"github.com/der-antikeks/gisp/math"
)

var (
	ErrNoSuchEntity    = errors.New("no such entity")
	ErrNoSuchComponent = errors.New("no such component")
)

type EntitySystem struct {
	lock sync.RWMutex

	observers map[*aspect]struct{ add, update, remove *Observer }

	next       Entity
	pool       []Entity
	components map[Entity]map[ComponentType]Component
	aspects    map[Entity][]*aspect

	loader *AssetLoaderSystem
}

func NewEntitySystem(loader *AssetLoaderSystem) *EntitySystem {
	s := &EntitySystem{
		observers: map[*aspect]struct{ add, update, remove *Observer }{
			nil: struct{ add, update, remove *Observer }{
				add:    NewObserver(),
				update: NewObserver(),
				remove: NewObserver(),
			}},

		next:       1,
		pool:       []Entity{},
		components: map[Entity]map[ComponentType]Component{},
		aspects:    map[Entity][]*aspect{},

		loader: loader,
	}

	go func() {
		// TODO: remove locks
	}()

	return s
}

// Create a new Entity
func (s *EntitySystem) Entity() Entity {
	s.lock.Lock()
	defer s.lock.Unlock()

	var e Entity
	if l := len(s.pool); l > 0 {
		e, s.pool = s.pool[l-1], s.pool[:l-1]
	} else {
		e = s.next
		s.next++
	}

	s.components[e] = map[ComponentType]Component{}
	s.aspects[e] = []*aspect{}

	return e
}

// Delete Entity from Engine and send RemoveEvents to all registered observers
func (s *EntitySystem) Delete(e Entity) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.components[e]; !ok {
		return ErrNoSuchEntity
	}

	for _, a := range s.aspects[e] {
		s.observers[a].remove.Publish(MessageEntityRemove{Removed: e})
	}
	s.observers[nil].remove.Publish(MessageEntityRemove{Removed: e})

	delete(s.components, e)
	delete(s.aspects, e)
	s.pool = append(s.pool, e)
	return nil
}

func (s *EntitySystem) Query(types ...ComponentType) []Entity {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var ret []Entity
	var found bool
	for e, ecs := range s.components {
		found = true
		for _, t := range types {
			if _, ok := ecs[t]; !ok {
				found = false
				break
			}
		}
		if found {
			ret = append(ret, e)
		}
	}
	return ret
}

func (s *EntitySystem) Set(e Entity, components ...Component) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.components[e]; !ok {
		return ErrNoSuchEntity
	}

	// add/update component of entity
	var updated, added bool
	for _, c := range components {
		if !updated || !added {
			if _, ok := s.components[e][c.Type()]; ok {
				updated = true
			} else {
				added = true
			}
		}

		s.components[e][c.Type()] = c
	}

	// send update to old aspect observers
	if updated {
		for _, a := range s.aspects[e] {
			s.observers[a].update.Publish(MessageEntityUpdate{Updated: e})
		}
		s.observers[nil].update.Publish(MessageEntityUpdate{Updated: e})
	}

	// add new aspect observers
	if added {
		ts := make([]ComponentType, len(s.components[e]))
		for t := range s.components[e] {
			ts = append(ts, t)
		}

		var already bool
		for a := range s.observers {
			if a == nil {
				continue
			}

			already = false
			for _, h := range s.aspects[e] {
				if a == h {
					already = true
					break
				}
			}

			// add aspect to entity, send add to observers
			if !already && a.subset(ts) {
				s.observers[a].add.Publish(MessageEntityAdd{Added: e})
				s.aspects[e] = append(s.aspects[e], a)
			}
		}
		s.observers[nil].add.Publish(MessageEntityAdd{Added: e})
	}

	return nil
}

func (s *EntitySystem) Remove(e Entity, types ...ComponentType) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.components[e]; !ok {
		return ErrNoSuchEntity
	}

	for _, t := range types {
		delete(s.components[e], t)
	}

	ts := make([]ComponentType, len(s.components[e]))
	for t := range s.components[e] {
		ts = append(ts, t)
	}

	for i, a := range s.aspects[e] {
		if !a.subset(ts) {
			// aspect does not accept entity anymore
			// remove aspect from entities slice
			copy(s.aspects[e][i:], s.aspects[e][i+1:])
			s.aspects[e][len(s.aspects[e])-1] = nil
			s.aspects[e] = s.aspects[e][:len(s.aspects[e])-1]

			s.observers[a].remove.Publish(MessageEntityRemove{Removed: e})
		}
	}
	return nil
}

func (s *EntitySystem) Get(e Entity, t ComponentType) (Component, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if _, ok := s.components[e]; !ok {
		return nil, ErrNoSuchEntity
	}

	c, ok := s.components[e][t]
	if !ok {
		return nil, ErrNoSuchComponent
	}

	return c, nil
}

func (s *EntitySystem) getAspect(ts []ComponentType) *aspect {
	s.lock.Lock()
	defer s.lock.Unlock()

	// old aspect
	for a := range s.observers {
		if a.equals(ts) {
			return a
		}
	}

	// new aspect
	a := &aspect{ts}
	s.observers[a] = struct{ add, update, remove *Observer }{
		add:    NewObserver(),
		update: NewObserver(),
		remove: NewObserver(),
	}

	// connect new aspect with existing entities
	for e, cm := range s.components {
		ts := make([]ComponentType, len(cm))
		for t := range cm {
			ts = append(ts, t)
		}
		if a.subset(ts) {
			s.aspects[e] = append(s.aspects[e], a)
		}
	}

	return a
}

func (s *EntitySystem) OnAdd(ts ...ComponentType) *Observer {
	return s.observers[s.getAspect(ts)].add
}

func (s *EntitySystem) OnUpdate(ts ...ComponentType) *Observer {
	return s.observers[s.getAspect(ts)].update
}

func (s *EntitySystem) OnRemove(ts ...ComponentType) *Observer {
	return s.observers[s.getAspect(ts)].remove
}

func (s *EntitySystem) CreateSplashScreen() {
	s.createCube()
	s.createSphere()
}

func (s *EntitySystem) CreateMainMenu() {
	for i := 0; i < 2000; i++ {
		s.createRndCube()
	}
}

func (s *EntitySystem) createCube() Entity {
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
	geo := s.getGeometry("cube")

	// material
	mat := s.getMaterial("flat")
	mat.Set("lightPosition", math.Vector{5, 5, 0, 1})
	mat.Set("lightDiffuse", math.Color{1, 0, 0})
	mat.Set("opacity", 0.8)

	tex, err := s.loader.LoadTexture("assets/cube/cube.png")
	if err != nil {
		log.Fatal("could not load texture:", err)
	}
	mat.Set("diffuseMap", tex)

	// scene
	stc := SceneTree{Name: "mainscene"}

	// Entity
	cube := s.Entity()
	if err := s.Set(
		cube,
		trans, geo, mat, vel, stc,
	); err != nil {
		log.Fatal("could not create cube:", err)
	}

	return cube
}

func (s *EntitySystem) createRndCube() Entity {
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
	geo := s.getGeometry("cube")

	// material
	mat := s.getMaterial("basic")
	mat.Set("diffuse", math.Color{
		r(0.5, 1),
		r(0.5, 1),
		r(0.5, 1),
	})
	mat.Set("opacity", r(0.2, 1))

	// scene
	stc := SceneTree{Name: "mainscene"}

	// Entity
	cube := s.Entity()
	if err := s.Set(
		cube,
		trans, geo, mat, stc,
	); err != nil {
		log.Fatal("could not create cube:", err)
	}

	return cube
}

func (s *EntitySystem) createSphere() Entity {
	// Transformation
	trans := Transformation{
		Position: math.Vector{5, 0, 0},
		Rotation: math.Quaternion{0, 0, 0, 1},
		Scale:    math.Vector{1, 1, 1},
		Up:       math.Vector{0, 1, 0},
	}

	// geometry
	geo := s.getGeometry("sphere")

	// material
	mat := s.getMaterial("phong")

	tex, err := s.loader.LoadTexture("assets/uvtemplate.png")
	if err != nil {
		log.Fatal("could not load texture:", err)
	}
	mat.Set("diffuseMap", tex)

	// scene
	stc := SceneTree{Name: "mainscene"}

	// Entity
	sphere := s.Entity()
	if err := s.Set(
		sphere,
		trans, geo, mat, stc,
	); err != nil {
		log.Fatal("could not create sphere:", err)
	}

	return sphere
}

func (s *EntitySystem) getMaterial(e string) Material {
	mat := Material{
		Program:  e,
		uniforms: map[string]interface{}{},
		program:  s.loader.GetShader(e),
	}

	// preset with standard values
	for n, v := range mat.program.uniforms {
		if v.standard != nil {
			mat.uniforms[n] = v.standard
		}
	}

	return mat
}

func (s *EntitySystem) getGeometry(e string) Geometry {
	mb := s.loader.GetMeshBuffer(e)

	geo := Geometry{
		File:     e,
		mesh:     mb,
		Bounding: mb.Bounding,
	}

	return geo
}

func (s *EntitySystem) CreatePerspectiveCamera(fov, aspect, near, far float64) Entity {
	t := Transformation{
		Position: math.Vector{0, 0, -10},
		//Rotation: math.Quaternion{},
		Scale: math.Vector{1, 1, 1},
		Up:    math.Vector{0, 1, 0},
	}
	t.Rotation = math.QuaternionLookAt(t.Position, math.Vector{0, 0, 0}, t.Up)

	c := s.Entity()
	if err := s.Set(
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

func (s *EntitySystem) CreateOrthographicCamera(left, right, top, bottom, near, far float64) Entity {
	t := Transformation{
		Position: math.Vector{0, 0, -10},
		//Rotation: math.Quaternion{},
		Scale: math.Vector{1, 1, 1},
		Up:    math.Vector{0, 1, 0},
	}
	t.Rotation = math.QuaternionLookAt(t.Position, math.Vector{0, 0, 0}, t.Up)

	c := s.Entity()
	if err := s.Set(
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
