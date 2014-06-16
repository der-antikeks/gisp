package game

import (
	"errors"
	"log"
	"math"
	"math/rand"
	"sync"

	"github.com/der-antikeks/mathgl/mgl32"
)

var (
	ErrNoSuchEntity    = errors.New("no such entity")
	ErrNoSuchComponent = errors.New("no such component")
)

/*
	create, load/save entities, manage components

	old engine / entity manager
	NewEntity()
	Delete(Entity)
	Set/Get/RemoveComponent()
	SubscribeOnAdd(Aspect(type,...), chan Entity)
*/
type entitySystem struct {
	lock sync.RWMutex

	observers map[*aspect]struct{ add, update, remove *Observer }

	next       Entity
	pool       []Entity
	components map[Entity]map[ComponentType]Component
	aspects    map[Entity][]*aspect
}

var (
	entityInstance *entitySystem
	entityOnce     sync.Once
)

func EntitySystem() *entitySystem {
	entityOnce.Do(func() {
		entityInstance = &entitySystem{
			next:       1,
			pool:       []Entity{},
			components: map[Entity]map[ComponentType]Component{},
			aspects:    map[Entity][]*aspect{},
		}

		entityInstance.observers = map[*aspect]struct{ add, update, remove *Observer }{
			nil: struct{ add, update, remove *Observer }{
				add:    NewObserver(entityInstance.onAddSendAll(nil)),
				update: NewObserver(nil),
				remove: NewObserver(nil),
			},
		}

		go func() {
			// TODO: remove locks
		}()
	})

	return entityInstance
}

// Create a new Entity
func (s *entitySystem) Entity() Entity {
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
func (s *entitySystem) Delete(e Entity) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, ok := s.components[e]; !ok {
		return ErrNoSuchEntity
	}

	for _, a := range s.aspects[e] {
		s.observers[a].remove.PublishAndWait(MessageEntityRemove{Removed: e})
	}
	s.observers[nil].remove.PublishAndWait(MessageEntityRemove{Removed: e})

	delete(s.components, e)
	delete(s.aspects, e)
	s.pool = append(s.pool, e)
	return nil
}

func (s *entitySystem) Query(types ...ComponentType) []Entity {
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

func (s *entitySystem) Set(e Entity, components ...Component) error {
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

func (s *entitySystem) Remove(e Entity, types ...ComponentType) error {
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
			s.observers[a].remove.PublishAndWait(MessageEntityRemove{Removed: e})

			// remove aspect from entities slice
			copy(s.aspects[e][i:], s.aspects[e][i+1:])
			s.aspects[e][len(s.aspects[e])-1] = nil
			s.aspects[e] = s.aspects[e][:len(s.aspects[e])-1]
		}
	}
	return nil
}

func (s *entitySystem) Get(e Entity, t ComponentType) (Component, error) {
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

func (s *entitySystem) getAspect(ts []ComponentType) *aspect {
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
		add:    NewObserver(s.onAddSendAll(ts)),
		update: NewObserver(nil),
		remove: NewObserver(nil),
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

func (s *entitySystem) OnAdd(ts ...ComponentType) *Observer {
	return s.observers[s.getAspect(ts)].add
}

func (s *entitySystem) onAddSendAll(ts []ComponentType) func() <-chan interface{} {
	return func() <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			for _, e := range s.Query(ts...) {
				c <- MessageEntityAdd{Added: e}
			}
		}()
		return c
	}
}

func (s *entitySystem) OnUpdate(ts ...ComponentType) *Observer {
	return s.observers[s.getAspect(ts)].update
}

func (s *entitySystem) OnRemove(ts ...ComponentType) *Observer {
	return s.observers[s.getAspect(ts)].remove
}

func (s *entitySystem) CreateSplashScreen() {
	s.createCube()
	s.createSphere()

	s.createBillboard()

	s.createString()
}

func (s *entitySystem) CreateMainMenu() {
	// multiple cubes on subscene
	for i := 0; i < 2000; i++ {
		s.createRndCube()
	}

	// subscene camera
	t := Transformation{
		Position: mgl32.Vec3{0, 0, -10},
		//Rotation: mgl32.Quat{},
		Scale: mgl32.Vec3{1, 1, 1},
		Up:    mgl32.Vec3{0, 1, 0},
	}
	t.Rotation = mgl32.QuatLookAtV(t.Position, mgl32.Vec3{0, 0, 0}, t.Up)

	target := AssetLoaderSystem(nil).NewRenderTarget(800, 600)
	target.Color = mgl32.Vec3{0.2, 0.2, 0.2}
	target.Alpha = 1.0
	target.Clear = true

	w, h := GlContextSystem(nil).Size()
	if err := s.Set(
		s.Entity(),
		t,
		Projection{
			Matrix:   mgl32.Perspective(45.0, float32(w)/float32(h), 0.1, 200.0),
			Target:   target,
			Priority: 0,
		},
		Scene{Name: "subscene"},
	); err != nil {
		log.Fatal("could not create subscene camera:", err)
	}

	// billboard window to subscene
	// Transformation
	t = Transformation{
		Position: mgl32.Vec3{0, 2, 5},
		Rotation: mgl32.Quat{1, mgl32.Vec3{0, 0, 0}},
		Scale:    mgl32.Vec3{1, 1, 1},
		Up:       mgl32.Vec3{0, 1, 0},
	}

	// geometry
	geo := s.getGeometry("plane")

	// material
	mat := s.getMaterial("fixedbb")
	mat.Set("diffuseMap", target.texture)
	mat.Set("opacity", 0.8)
	mat.Set("size", mgl32.Vec2{1, 1})

	// Entity
	if err := s.Set(
		s.Entity(),
		t, geo, mat,
		Scene{Name: "mainscene"},
	); err != nil {
		log.Fatal("could not create billboard:", err)
	}

}

func (s *entitySystem) createCube() Entity {
	// Transformation
	trans := Transformation{
		Position: mgl32.Vec3{0, 0, 0},
		Rotation: mgl32.QuatRotate(math.Pi/4.0, mgl32.Vec3{1, 0.5, 0}),
		Scale:    mgl32.Vec3{1, 1, 1},
		Up:       mgl32.Vec3{0, 1, 0},
	}

	// velocity, rate of change per second
	vel := Velocity{
		Velocity: mgl32.Vec3{0, 0, 0},
		Angular: mgl32.Vec3{
			mgl32.DegToRad(45.0),
			mgl32.DegToRad(5.0),
			0,
		},
	}

	// geometry
	geo := s.getGeometry("fighter/fighter.obj")

	// material
	mat := s.getMaterial("flat")
	mat.Set("lightPosition", []mgl32.Vec3{{5, 5, 0}})
	mat.Set("lightDiffuse", []mgl32.Vec3{{1, 0, 0}})
	mat.Set("opacity", 0.8)

	tex, err := AssetLoaderSystem(nil).LoadTexture("fighter/fighter.png")
	if err != nil {
		log.Fatal("could not load texture:", err)
	}
	mat.Set("diffuseMap", tex)

	// scene
	stc := Scene{Name: "mainscene"}

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

func (s *entitySystem) createRndCube() Entity {
	// helper
	r := func(min, max float32) float32 {
		return rand.Float32()*(max-min) + min
	}
	d := func() float32 {
		if rand.Intn(2) == 0 {
			return -1
		}
		return 1
	}

	// transformation
	trans := Transformation{
		Position: mgl32.Vec3{
			r(1, 100) * d(),
			r(1, 100) * d(),
			r(1, 100) * d(),
		},
		Rotation: mgl32.QuatRotate(r(0, math.Pi*2), (mgl32.Vec3{
			rand.Float32(),
			rand.Float32(),
			rand.Float32(),
		}).Normalize()),
		Scale: mgl32.Vec3{1, 1, 1},
		Up:    mgl32.Vec3{0, 1, 0},
	}

	// velocity
	vel := Velocity{
		Velocity: mgl32.Vec3{
			r(-1, 1),
			r(-1, 1),
			r(-1, 1),
		},
		Angular: mgl32.Vec3{
			mgl32.DegToRad(r(0, 45)),
			mgl32.DegToRad(r(0, 45)),
			mgl32.DegToRad(r(0, 45)),
		},
	}

	// geometry
	geo := s.getGeometry("cube")

	// material
	mat := s.getMaterial("basic")
	mat.Set("diffuse", mgl32.Vec3{
		r(0.5, 1),
		r(0.5, 1),
		r(0.5, 1),
	})
	mat.Set("opacity", float64(r(0.2, 1)))

	// scene
	stc := Scene{Name: "subscene"}

	// Entity
	cube := s.Entity()
	if err := s.Set(
		cube,
		trans, vel, geo, mat, stc,
	); err != nil {
		log.Fatal("could not create cube:", err)
	}

	return cube
}

func (s *entitySystem) createSphere() Entity {
	// Transformation
	trans := Transformation{
		Position: mgl32.Vec3{5, 0, 0},
		Rotation: mgl32.Quat{1, mgl32.Vec3{0, 0, 0}},
		Scale:    mgl32.Vec3{1, 1, 1},
		Up:       mgl32.Vec3{0, 1, 0},
	}

	// geometry
	geo := s.getGeometry("sphere")

	// material
	//mat := s.getMaterial("phong")
	mat := s.getMaterial("flat")

	mat.Set("lightCount", 3)
	mat.Set("lightDiffuse", []mgl32.Vec3{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}})
	mat.Set("lightPosition", []mgl32.Vec3{{5, 10, 0}, {-5, 10, 0}, {5, -10, 0}})
	mat.Set("lightPower", []float64{100.0, 50.0, 100.0})

	tex, err := AssetLoaderSystem(nil).LoadTexture("uvtemplate.png")
	if err != nil {
		log.Fatal("could not load texture:", err)
	}
	mat.Set("diffuseMap", tex)

	// scene
	stc := Scene{Name: "mainscene"}

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

func (s *entitySystem) createBillboard() Entity {
	// Transformation
	trans := Transformation{
		Position: mgl32.Vec3{2, 0, -5},
		Rotation: mgl32.Quat{1, mgl32.Vec3{0, 0, 0}},
		Scale:    mgl32.Vec3{1, 1, 1},
		Up:       mgl32.Vec3{0, 1, 0},
	}

	// geometry
	geo := s.getGeometry("plane")

	// material
	mat := s.getMaterial("billboard")
	tex, err := AssetLoaderSystem(nil).LoadTexture("uvtemplate.png")
	if err != nil {
		log.Fatal("could not load texture:", err)
	}
	mat.Set("diffuseMap", tex)
	mat.Set("opacity", 0.8)

	// scene
	stc := Scene{Name: "mainscene"}

	// Entity
	billboard := s.Entity()
	if err := s.Set(
		billboard,
		trans, geo, mat, stc,
	); err != nil {
		log.Fatal("could not create billboard:", err)
	}

	return billboard
}

func (s *entitySystem) createString() Entity {
	// transformation
	trans := Transformation{
		Position: mgl32.Vec3{2, 0, 0},
		Rotation: mgl32.Quat{1, mgl32.Vec3{0, 0, 0}},
		Scale:    mgl32.Vec3{10, 10, 10},
		Up:       mgl32.Vec3{0, 1, 0},
	}

	// load font
	font, err := AssetLoaderSystem(nil).LoadSDFFont("luxisr.ttf", 32.0, 32, 127)
	if err != nil {
		log.Fatal("could not load font file:", err)
	}

	// geometry
	m, b := AssetLoaderSystem(nil).CreateString(font, "Lorem ipsum dolor sit amet, consectetur adipiscing elit.")
	geo := Geometry{
		mesh:     m,
		Bounding: b,
	}

	// material
	mat := s.getMaterial("sdf")
	//mat.Set("diffuseMap", font.sdf)
	mat.Set("distanceFieldMap", font.sdf)
	mat.Set("diffuse", mgl32.Vec3{0, 1, 1})
	mat.Set("opacity", 1.0/(4*10)) // 0.25 / (float64(spread) * scale)

	// velocity
	vel := Velocity{
		Velocity: mgl32.Vec3{0, 0, 0},
		Angular: mgl32.Vec3{
			0,
			mgl32.DegToRad(45.0),
			0,
		},
	}

	// scene
	stc := Scene{Name: "mainscene"}

	// Entity
	str := s.Entity()
	if err := s.Set(
		str,
		trans, geo, mat, vel, stc,
	); err != nil {
		log.Fatal("could not create string:", err)
	}

	return str
}

func (s *entitySystem) getMaterial(e string) Material {
	mat := Material{
		Program:  e,
		uniforms: map[string]interface{}{},
		program:  AssetLoaderSystem(nil).LoadShader(e),
	}

	// preset with standard values
	for n, v := range mat.program.uniforms {
		if v.standard != nil {
			mat.uniforms[n] = v.standard
		}
	}

	return mat
}

func (s *entitySystem) getGeometry(e string) Geometry {
	var m *meshbuffer
	var b Boundary

	switch e {
	default:
		m, b = AssetLoaderSystem(nil).LoadOBJ(e)
	case "sphere":
		m, b = AssetLoaderSystem(nil).SpherePrimitive(2, 100, 50)
	case "cube":
		m, b = AssetLoaderSystem(nil).CubePrimitive(2)
	case "plane":
		m, b = AssetLoaderSystem(nil).PlanePrimitive(1, 1)
	}

	geo := Geometry{
		File:     e,
		mesh:     m,
		Bounding: b,
	}

	return geo
}

func (s *entitySystem) CreatePerspectiveCamera(fov, aspect, near, far float32) Entity {
	t := Transformation{
		Position: mgl32.Vec3{0, 0, -10},
		//Rotation: mgl32.Quat{},
		Scale: mgl32.Vec3{1, 1, 1},
		Up:    mgl32.Vec3{0, 1, 0},
	}
	t.Rotation = mgl32.QuatLookAtV(t.Position, mgl32.Vec3{0, 0, 0}, t.Up)

	c := s.Entity()
	if err := s.Set(
		c,
		Projection{
			Matrix:   mgl32.Perspective(fov, aspect, near, far),
			Priority: 10,
		},
		t,
		Scene{Name: "mainscene"},
	); err != nil {
		log.Fatal("could not create camera:", err)
	}

	return c
}

func (s *entitySystem) CreateOrthographicCamera(left, right, top, bottom, near, far float32) Entity {
	t := Transformation{
		Position: mgl32.Vec3{0, 0, -10},
		//Rotation: mgl32.Quat{},
		Scale: mgl32.Vec3{1, 1, 1},
		Up:    mgl32.Vec3{0, 1, 0},
	}
	t.Rotation = mgl32.QuatLookAtV(t.Position, mgl32.Vec3{0, 0, 0}, t.Up)

	c := s.Entity()
	if err := s.Set(
		c,
		Projection{ // TODO: top/bottom switched?
			Matrix:   mgl32.Ortho(left, right, bottom, top, near, far),
			Priority: 10,
		},
		t,
		Scene{Name: "mainscene"},
	); err != nil {
		log.Fatal("could not create camera:", err)
	}

	return c
}
