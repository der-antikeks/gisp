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
	drawChan, camChan, updChan chan interface{}

	scenes map[string]struct {
		drawable []Entity
		lights   []Entity
		camera   Entity
	}

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
			drawChan: make(chan interface{}),
			camChan:  make(chan interface{}),
			updChan:  make(chan interface{}),

			scenes: map[string]struct {
				drawable []Entity
				lights   []Entity
				camera   Entity
			}{},
		}

		go func() {
			renderInstance.Restart()

			for {
				select {
				case event := <-renderInstance.drawChan:
					switch e := event.(type) {
					case MessageEntityAdd:
						sn := renderInstance.getScene(e.Added)
						sc := renderInstance.scenes[sn]
						sc.drawable = append(sc.drawable, e.Added)
						renderInstance.scenes[sn] = sc

					case MessageEntityRemove:
						sn := renderInstance.getScene(e.Removed)
						sc := renderInstance.scenes[sn]
						for i, f := range sc.drawable {
							if f == e.Removed {
								sc.drawable = append(sc.drawable[:i], sc.drawable[i+1:]...)
								break
							}
						}
						renderInstance.scenes[sn] = sc
					}

				case event := <-renderInstance.camChan:
					switch e := event.(type) {
					case MessageEntityAdd:
						sn := renderInstance.getScene(e.Added)
						sc := renderInstance.scenes[sn]
						sc.camera = e.Added
						renderInstance.scenes[sn] = sc
					case MessageEntityRemove:
						sn := renderInstance.getScene(e.Removed)
						sc := renderInstance.scenes[sn]
						if sc.camera == e.Removed {
							sc.camera = NoEntity
						}
						renderInstance.scenes[sn] = sc
					}

				case event := <-renderInstance.updChan:
					switch e := event.(type) {
					case MessageUpdate:
						if err := renderInstance.Update(e.Delta); err != nil {
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

	EntitySystem().OnAdd(TransformationType, GeometryType, MaterialType, SceneType).Subscribe(s.drawChan, PriorityRender)
	EntitySystem().OnRemove(TransformationType, GeometryType, MaterialType, SceneType).Subscribe(s.drawChan, PriorityRender)

	EntitySystem().OnAdd(TransformationType, ProjectionType, SceneType).Subscribe(s.camChan, PriorityRender)
	EntitySystem().OnRemove(TransformationType, ProjectionType, SceneType).Subscribe(s.camChan, PriorityRender)
}

func (s *renderSystem) Stop() {
	GameStateSystem().OnUpdate().Unsubscribe(s.updChan)

	EntitySystem().OnAdd(TransformationType, GeometryType, MaterialType, SceneType).Unsubscribe(s.drawChan)
	EntitySystem().OnRemove(TransformationType, GeometryType, MaterialType, SceneType).Unsubscribe(s.drawChan)

	EntitySystem().OnAdd(TransformationType, ProjectionType, SceneType).Unsubscribe(s.camChan)
	EntitySystem().OnRemove(TransformationType, ProjectionType, SceneType).Unsubscribe(s.camChan)

	//s.drawable = []Entity{}
	//s.camera = NoEntity
	// TODO: empty scenes?
}

func (s *renderSystem) AddRenderPass(camera Entity, priority int) {
	// TODO
}

func (s *renderSystem) Update(delta time.Duration) error {
	for n := range s.scenes {
		if err := s.updateScene(delta, n); err != nil {
			return err
		}
	}
	return nil
}

func (s *renderSystem) getScene(e Entity) string {
	ec, err := EntitySystem().Get(e, SceneType)
	if err != nil {
		return ""
	}
	return ec.(Scene).Name
}

func (s *renderSystem) updateScene(delta time.Duration, sc string) error {
	color := mgl32.Vec3{0, 0, 0}
	alpha := 1.0
	s.setClearColor(color, alpha)

	// w, h := GlContextSystem(nil).Size()
	// gl.Viewport(0, 0, w, h) TODO: already set in WindowManager onResize(), must be changed with frambuffer?

	// TODO: clearing should depend on rendertarget
	clear := true
	if clear {
		GlContextSystem(nil).MainThread(func() {
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		})
	}

	// TODO: move rendertarget to camera? or camera like interface (Viewport(w,h), Projection, etc.)
	if s.scenes[sc].camera == NoEntity {
		return fmt.Errorf("no camera found for renderSystem")
	}
	// update scene matrix (all objects)
	// update camera matrix if not child of scene
	// calculate frustum of camera
	ec, err := EntitySystem().Get(s.scenes[sc].camera, ProjectionType)
	if err != nil {
		return err
	}
	p := ec.(Projection)

	ec, err = EntitySystem().Get(s.scenes[sc].camera, TransformationType)
	if err != nil {
		return err
	}
	t := ec.(Transformation)

	projScreenMatrix := p.Matrix.Mul4(t.MatrixWorld().Inv())
	frustum := Mat4ToFrustum(projScreenMatrix)
	// fetch all objects visible in frustum
	//opaque, transparent, _ := SpatialSystem().VisibleEntities(sc, t.Position, frustum)
	//sort by z, by material, etc.
	opaque, transparent := s.visibleEntities(frustum, t.Position, s.scenes[sc].drawable)

	// opaque pass (front-to-back order)
	GlContextSystem(nil).MainThread(func() {
		gl.Disable(gl.BLEND)

		for _, e := range opaque {
			s.renderEntity(e, s.scenes[sc].camera)
		}
	})

	// transparent pass (back-to-front order)
	GlContextSystem(nil).MainThread(func() {
		gl.Enable(gl.BLEND)
		gl.BlendEquationSeparate(gl.FUNC_ADD, gl.FUNC_ADD)
		gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

		for _, e := range transparent {
			s.renderEntity(e, s.scenes[sc].camera)
		}
	})

	// swap buffers
	GlContextSystem(nil).Update()
	return nil
}

func (s *renderSystem) setClearColor(color mgl32.Vec3, alpha float64) {
	GlContextSystem(nil).MainThread(func() {
		gl.ClearColor(gl.GLclampf(color[0]), gl.GLclampf(color[1]), gl.GLclampf(color[2]), gl.GLclampf(alpha))
	})
}

// TODO: replace with spatial system

func (s *renderSystem) visibleEntities(frustum Frustum, cp mgl32.Vec3, drawable []Entity) (opaque, transparent []Entity) {
	opaque = make([]Entity, 0)
	transparent = make([]Entity, 0)
	var err error
	var ec Component

	zorder := map[Entity]float32{}
	cp4 := mgl32.Vec4{cp[0], cp[1], cp[2], 0} // TODO

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

		c, r := g.Bounding.Sphere()
		c = t.MatrixWorld().Mul4x1(c)
		r *= mgl32.ExtractMaxScale(t.MatrixWorld())

		if frustum.IntersectsSphere(c, r) {
			zorder[e] = c.Sub(cp4).Len()

			if m.opaque() {
				opaque = append(opaque, e)
			} else {
				transparent = append(transparent, e)
			}
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

	return opaque, transparent
}

func (s *renderSystem) renderEntity(object, camera Entity) error {
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

	ec, err = EntitySystem().Get(camera, ProjectionType)
	if err != nil {
		return err
	}
	projection := ec.(Projection)

	ec, err = EntitySystem().Get(camera, TransformationType)
	if err != nil {
		return err
	}
	cameratransform := ec.(Transformation)

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
	case float64:
		s.currentProgram.uniforms[name].location.Uniform1f(float32(t))
	case float32:
		s.currentProgram.uniforms[name].location.Uniform1f(t)

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

	case bool:
		if t {
			s.currentProgram.uniforms[name].location.Uniform1i(1)
		} else {
			s.currentProgram.uniforms[name].location.Uniform1i(0)
		}
	}

	return nil
}
