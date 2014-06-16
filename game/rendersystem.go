package game

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/der-antikeks/mathgl/mgl32"

	"github.com/go-gl/gl"
)

/*
	manage render passes, priorities and render to screen/buffer

	AddPass(CameraEntity, priority)
		render to framebuffer, render to screen
		render opaque, transparent
*/
type renderSystem struct {
	camChan, updChan chan interface{}
	renderPasses     []Entity
	updatePrio       bool

	currentGeometry *meshbuffer
	currentProgram  *shaderprogram
	currentTextures []gl.Texture // usedTextureUnits
}

var (
	renderInstance *renderSystem
	renderOnce     sync.Once
)

func RenderSystem() *renderSystem {
	renderOnce.Do(func() {
		renderInstance = &renderSystem{
			camChan:      make(chan interface{}),
			updChan:      make(chan interface{}),
			renderPasses: []Entity{},
		}

		go func() {
			renderInstance.Restart()

			for {
				select {

				case event := <-renderInstance.camChan:
					switch e := event.(type) {
					case MessageEntityAdd:
						renderInstance.addRenderPass(e.Added)
					case MessageEntityUpdate:
						renderInstance.updateRenderPass(e.Updated)
					case MessageEntityRemove:
						renderInstance.removeRenderPass(e.Removed)
					}

				case event := <-renderInstance.updChan:
					switch e := event.(type) {
					case MessageUpdate:
						if err := renderInstance.update(e.Delta); err != nil {
							log.Println("could not render:", err)
						}
					}
				}
			}
		}()
	})

	return renderInstance
}

func (s *renderSystem) Restart() {
	GameStateSystem().OnUpdate().Subscribe(s.updChan, PriorityRender)

	EntitySystem().OnAdd(TransformationType, ProjectionType, SceneType).Subscribe(s.camChan, PriorityRender)
	EntitySystem().OnUpdate(TransformationType, ProjectionType, SceneType).Subscribe(s.camChan, PriorityRender)
	EntitySystem().OnRemove(TransformationType, ProjectionType, SceneType).Subscribe(s.camChan, PriorityRender)
}

func (s *renderSystem) Stop() {
	GameStateSystem().OnUpdate().Unsubscribe(s.updChan)

	EntitySystem().OnAdd(TransformationType, ProjectionType, SceneType).Unsubscribe(s.camChan)
	EntitySystem().OnRemove(TransformationType, ProjectionType, SceneType).Unsubscribe(s.camChan)
}

func (s *renderSystem) addRenderPass(camera Entity) {
	s.renderPasses = append(s.renderPasses, camera)
	s.updatePrio = true
}

func (s *renderSystem) updateRenderPass(_ Entity) {
	s.updatePrio = true
}

func (s *renderSystem) removeRenderPass(camera Entity) {
	for i, f := range s.renderPasses {
		if f == camera {
			s.renderPasses = append(s.renderPasses[:i], s.renderPasses[i+1:]...)
			return
		}
	}
}

func (s *renderSystem) update(delta time.Duration) error {
	if s.updatePrio {
		// TODO: sort renderpasses based on priority
		s.updatePrio = false
	}

	for _, camera := range s.renderPasses {
		if err := s.renderScene(delta, camera); err != nil {
			return err
		}
	}

	// swap buffers
	GlContextSystem(nil).Update()
	return nil
}

func (s *renderSystem) renderScene(delta time.Duration, camera Entity) error {
	if camera == NoEntity {
		return fmt.Errorf("no camera found for renderSystem")
	}

	// get camera components
	ec, err := EntitySystem().Get(camera, ProjectionType)
	if err != nil {
		return err
	}
	p := ec.(Projection)

	ec, err = EntitySystem().Get(camera, TransformationType)
	if err != nil {
		return err
	}
	t := ec.(Transformation)

	ec, err = EntitySystem().Get(camera, SceneType)
	if err != nil {
		return err
	}
	sc := ec.(Scene).Name

	// set rendertarget
	color := mgl32.Vec3{0, 0, 0}
	alpha := 1.0
	clear := true

	if target := p.Target; target != nil {
		color = target.Color
		alpha = target.Alpha
		clear = target.Clear

		// TODO: cache binding?
		GlContextSystem(nil).MainThread(func() {
			target.frameBuffer.Bind()
		})
		defer GlContextSystem(nil).MainThread(func() {
			target.frameBuffer.Unbind()
		})
		//gl.Viewport(0, 0, target.texture.w, target.texture.h)
	} else {
		// w, h := GlContextSystem(nil).Size()
		// gl.Viewport(0, 0, w, h)
		// TODO: already set in WindowManager onResize(), must be changed after frambuffer?
	}

	if clear {
		GlContextSystem(nil).MainThread(func() {
			gl.ClearColor(gl.GLclampf(color[0]), gl.GLclampf(color[1]), gl.GLclampf(color[2]), gl.GLclampf(alpha))
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		})
	}

	// update scene matrix (all objects)
	// update camera matrix if not child of scene
	// calculate frustum of camera
	projScreenMatrix := p.Matrix.Mul4(t.MatrixWorld().Inv())
	frustum := Mat4ToFrustum(projScreenMatrix)

	// fetch all objects of scene visible in frustum
	visible := SpatialSystem().IntersectsFrustum(sc, frustum)

	//sort by z, by material, etc.
	pos := t.MatrixWorld().Mul4x1(mgl32.Vec4{0, 0, 0, 1})
	opaque, transparent, _ := s.sortEntities(pos, visible)

	// opaque pass (front-to-back order)
	GlContextSystem(nil).MainThread(func() {
		gl.Disable(gl.BLEND)

		for _, e := range opaque {
			s.renderEntity(e, t, p)
		}
	})

	// transparent pass (back-to-front order)
	GlContextSystem(nil).MainThread(func() {
		gl.Enable(gl.BLEND)
		gl.BlendEquationSeparate(gl.FUNC_ADD, gl.FUNC_ADD)
		gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

		for _, e := range transparent {
			s.renderEntity(e, t, p)
		}
	})

	return nil
}

func (s *renderSystem) sortEntities(cp mgl32.Vec4, drawable []Entity) (opaque, transparent, light []Entity) {
	opaque = make([]Entity, 0)
	transparent = make([]Entity, 0)
	light = make([]Entity, 0)

	var err error
	var ec Component
	zorder := map[Entity]float32{}

	for _, e := range drawable {
		ec, err = EntitySystem().Get(e, TransformationType)
		if err != nil {
			continue
		}
		t := ec.(Transformation)

		ec, err = EntitySystem().Get(e, GeometryType)
		if err != nil {
			continue
		}
		g := ec.(Geometry)

		ec, err = EntitySystem().Get(e, MaterialType)
		if err != nil {
			continue
		}
		m := ec.(Material)

		c, _ := g.Bounding.Sphere()
		c = t.MatrixWorld().Mul4x1(c)

		zorder[e] = c.Sub(cp).Len()

		if m.opaque() {
			opaque = append(opaque, e)
		} else {
			transparent = append(transparent, e)
		}
	}

	// front-to-back order
	sort.Sort(byZ{
		entities: opaque,
		zorder:   zorder,
	})

	// back-to-front order
	sort.Sort(sort.Reverse(byZ{
		entities: transparent,
		zorder:   zorder,
	}))

	// TODO: shader program sorting?

	return opaque, transparent, light
}

type byZ struct {
	entities []Entity
	zorder   map[Entity]float32
}

func (a byZ) Len() int {
	return len(a.entities)
}
func (a byZ) Swap(i, j int) {
	a.entities[i], a.entities[j] = a.entities[j], a.entities[i]
}
func (a byZ) Less(i, j int) bool {
	return a.zorder[a.entities[i]] < a.zorder[a.entities[j]]
}

func (s *renderSystem) renderEntity(object Entity, cameratransform Transformation, projection Projection) error {
	ec, err := EntitySystem().Get(object, MaterialType)
	if err != nil {
		return err
	}
	material := ec.(Material)

	ec, err = EntitySystem().Get(object, GeometryType)
	if err != nil {
		return err
	}
	geometry := ec.(Geometry)

	ec, err = EntitySystem().Get(object, TransformationType)
	if err != nil {
		return err
	}
	objecttransform := ec.(Transformation)

	return s.render(
		objecttransform, material, geometry,
		cameratransform, projection)
}

func (s *renderSystem) render(
	objecttransform Transformation,
	material Material,
	geometry Geometry,

	cameratransform Transformation,
	projection Projection) error {

	// ### bind material
	var updateAttributes bool
	if material.program != s.currentProgram {
		s.currentProgram = material.program
		s.currentProgram.program.Use()

		updateAttributes = true
	}

	// unbind old material (Textures)
	s.unbindTextures() // TODO: no caching of old material bindings?

	// ### bind geometry
	if geometry.mesh != s.currentGeometry || updateAttributes {
		// disable old attributes
		s.currentProgram.DisableAttributes()

		// bind new buffers
		s.currentGeometry = geometry.mesh
		s.currentGeometry.VertexArrayObject.Bind()
		//defer s.currentGeometry.VertexArrayObject.Unbind()

		// vertices
		s.currentGeometry.PositionBuffer.Bind(gl.ARRAY_BUFFER)
		defer s.currentGeometry.PositionBuffer.Unbind(gl.ARRAY_BUFFER)
		s.currentProgram.EnableAttribute("vertexPosition")

		// normal
		s.currentGeometry.NormalBuffer.Bind(gl.ARRAY_BUFFER)
		defer s.currentGeometry.NormalBuffer.Unbind(gl.ARRAY_BUFFER)
		s.currentProgram.EnableAttribute("vertexNormal")

		// uv
		s.currentGeometry.UvBuffer.Bind(gl.ARRAY_BUFFER)
		defer s.currentGeometry.UvBuffer.Unbind(gl.ARRAY_BUFFER)
		s.currentProgram.EnableAttribute("vertexUV")
	}

	// ### update uniforms

	// update projection uniform
	s.UpdateUniform("projectionMatrix", projection.Matrix /*.Float32()*/)

	// viewMatrix
	viewMatrix := cameratransform.MatrixWorld().Inv()
	s.UpdateUniform("viewMatrix", viewMatrix /*.Float32()*/)

	// Model matrix : an identity matrix (model will be at the origin)
	s.UpdateUniform("modelMatrix", objecttransform.MatrixWorld() /*.Float32()*/)

	// modelViewMatrix
	modelViewMatrix := viewMatrix.Mul4(objecttransform.MatrixWorld())
	s.UpdateUniform("modelViewMatrix", modelViewMatrix /*.Float32()*/)

	// normalMatrix
	normalMatrix := mgl32.Mat4Normal(modelViewMatrix)
	s.UpdateUniform("normalMatrix", normalMatrix /*.Matrix3Float32()*/)

	// update material values
	for n, v := range material.uniforms {
		if err := s.UpdateUniform(n, v); err != nil {
			return err
		}
	}

	// ### draw
	s.currentGeometry.FaceBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
	defer s.currentGeometry.FaceBuffer.Unbind(gl.ELEMENT_ARRAY_BUFFER)
	gl.DrawElements(gl.TRIANGLES, s.currentGeometry.FaceCount*3, gl.UNSIGNED_SHORT, nil /* uintptr(start) */) // gl.UNSIGNED_INT, UNSIGNED_SHORT

	return nil
}

func (s *renderSystem) bindTexture(buffer gl.Texture) int {
	for s, b := range s.currentTextures {
		if b == buffer {
			return s
		}
	}

	slot := len(s.currentTextures)
	s.currentTextures = append(s.currentTextures, buffer)

	buffer.Bind(gl.TEXTURE_2D)
	gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(slot))

	return slot
}

func (s *renderSystem) unbindTextures() {
	for _, buffer := range s.currentTextures {
		buffer.Unbind(gl.TEXTURE_2D)
	}
	s.currentTextures = s.currentTextures[:0]
}

func (s *renderSystem) UpdateUniform(name string, value interface{}) error {
	if _, found := s.currentProgram.uniforms[name]; !found {
		log.Fatalf("unsupported uniform: %v", name)
		//return fmt.Errorf("unsupported uniform: %v", name)
	}

	switch t := value.(type) {
	default:
		log.Fatalf("%v has unknown type: %T", name, t)
		//return fmt.Errorf("%v has unknown type: %T", name, t)

	case *Texture:
		s.currentProgram.uniforms[name].location.Uniform1i(s.bindTexture(t.buffer))

	case int:
		s.currentProgram.uniforms[name].location.Uniform1i(t)
	case float32:
		s.currentProgram.uniforms[name].location.Uniform1f(t)
	case float64:
		s.currentProgram.uniforms[name].location.Uniform1f(float32(t))

	case mgl32.Mat3 /*[9]float32*/ :
		s.currentProgram.uniforms[name].location.UniformMatrix3fv(false, t)
	case mgl32.Mat4 /*[16]float32*/ :
		s.currentProgram.uniforms[name].location.UniformMatrix4fv(false, t)

	case mgl32.Vec2:
		s.currentProgram.uniforms[name].location.Uniform2f(t[0], t[1])
	case mgl32.Vec3:
		s.currentProgram.uniforms[name].location.Uniform3f(t[0], t[1], t[2])
	case mgl32.Vec4:
		s.currentProgram.uniforms[name].location.Uniform4f(t[0], t[1], t[2], t[3])

	case []int:
		var va []int32
		for _, v := range t {
			va = append(va, int32(v))
		}
		s.currentProgram.uniforms[name].location.Uniform1iv(len(t), va)
	case []float32:
		s.currentProgram.uniforms[name].location.Uniform1fv(len(t), t)
	case []float64:
		var va []float32
		for _, v := range t {
			va = append(va, float32(v))
		}
		s.currentProgram.uniforms[name].location.Uniform1fv(len(t), va)

	case []mgl32.Vec2:
		var va []float32
		for _, v := range t {
			va = append(va, v[0], v[1])
		}
		s.currentProgram.uniforms[name].location.Uniform2fv(len(t), va)
	case []mgl32.Vec3:
		var va []float32
		for _, v := range t {
			va = append(va, v[0], v[1], v[2])
		}
		s.currentProgram.uniforms[name].location.Uniform3fv(len(t), va)
	case []mgl32.Vec4:
		var va []float32
		for _, v := range t {
			va = append(va, v[0], v[1], v[2], v[3])
		}
		s.currentProgram.uniforms[name].location.Uniform4fv(len(t), va)

	case bool:
		if t {
			s.currentProgram.uniforms[name].location.Uniform1i(1)
		} else {
			s.currentProgram.uniforms[name].location.Uniform1i(0)
		}
	}

	return nil
}
