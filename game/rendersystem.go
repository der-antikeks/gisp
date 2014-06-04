package game

import (
	"fmt"
	"log"
	"sort"
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
type RenderSystem struct {
	context *GlContextSystem
	spatial *SpatialSystem
	state   *GameStateSystem
	ents    *EntitySystem // temporary

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

func NewRenderSystem(context *GlContextSystem, spatial *SpatialSystem, state *GameStateSystem, ents *EntitySystem) *RenderSystem {
	s := &RenderSystem{
		context: context,
		spatial: spatial,
		state:   state,
		ents:    ents,

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
		s.Restart()

		for {
			select {
			case event := <-s.drawChan:
				switch e := event.(type) {
				case MessageEntityAdd:
					sn := s.getScene(e.Added)
					sc := s.scenes[sn]
					sc.drawable = append(sc.drawable, e.Added)
					s.scenes[sn] = sc

				case MessageEntityRemove:
					sn := s.getScene(e.Removed)
					sc := s.scenes[sn]
					for i, f := range sc.drawable {
						if f == e.Removed {
							sc.drawable = append(sc.drawable[:i], sc.drawable[i+1:]...)
							break
						}
					}
					s.scenes[sn] = sc
				}

			case event := <-s.camChan:
				switch e := event.(type) {
				case MessageEntityAdd:
					sn := s.getScene(e.Added)
					sc := s.scenes[sn]
					sc.camera = e.Added
					s.scenes[sn] = sc
				case MessageEntityRemove:
					sn := s.getScene(e.Removed)
					sc := s.scenes[sn]
					if sc.camera == e.Removed {
						sc.camera = NoEntity
					}
					s.scenes[sn] = sc
				}

			case event := <-s.updChan:
				switch e := event.(type) {
				case MessageUpdate:
					if err := s.Update(e.Delta); err != nil {
						log.Println("could not render:", err)
					}
				}
			}
		}
	}()

	return s
}

func (s *RenderSystem) Restart() {
	s.state.OnUpdate().Subscribe(s.updChan, PriorityRender)

	s.ents.OnAdd(TransformationType, GeometryType, MaterialType, SceneType).Subscribe(s.drawChan, PriorityRender)
	s.ents.OnRemove(TransformationType, GeometryType, MaterialType, SceneType).Subscribe(s.drawChan, PriorityRender)

	s.ents.OnAdd(TransformationType, ProjectionType, SceneType).Subscribe(s.camChan, PriorityRender)
	s.ents.OnRemove(TransformationType, ProjectionType, SceneType).Subscribe(s.camChan, PriorityRender)
}

func (s *RenderSystem) Stop() {
	s.state.OnUpdate().Unsubscribe(s.updChan)

	s.ents.OnAdd(TransformationType, GeometryType, MaterialType, SceneType).Unsubscribe(s.drawChan)
	s.ents.OnRemove(TransformationType, GeometryType, MaterialType, SceneType).Unsubscribe(s.drawChan)

	s.ents.OnAdd(TransformationType, ProjectionType, SceneType).Unsubscribe(s.camChan)
	s.ents.OnRemove(TransformationType, ProjectionType, SceneType).Unsubscribe(s.camChan)

	//s.drawable = []Entity{}
	//s.camera = NoEntity
	// TODO: empty scenes?
}

func (s *RenderSystem) AddRenderPass(camera Entity, priority int) {
	// TODO
}

func (s *RenderSystem) Update(delta time.Duration) error {
	for n := range s.scenes {
		if err := s.updateScene(delta, n); err != nil {
			return err
		}
	}
	return nil
}

func (s *RenderSystem) getScene(e Entity) string {
	ec, err := s.ents.Get(e, SceneType)
	if err != nil {
		return ""
	}
	return ec.(Scene).Name
}

func (s *RenderSystem) updateScene(delta time.Duration, sc string) error {
	color := mgl32.Vec3{0, 0, 0}
	alpha := 1.0
	s.setClearColor(color, alpha)

	// w, h := s.context.Size()
	// gl.Viewport(0, 0, w, h) TODO: already set in WindowManager onResize(), must be changed with frambuffer?

	// TODO: clearing should depend on rendertarget
	clear := true
	if clear {
		s.context.MainThread(func() {
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		})
	}

	// TODO: move rendertarget to camera? or camera like interface (Viewport(w,h), Projection, etc.)
	if s.scenes[sc].camera == NoEntity {
		return fmt.Errorf("no camera found for RenderSystem")
	}
	// update scene matrix (all objects)
	// update camera matrix if not child of scene
	// calculate frustum of camera
	ec, err := s.ents.Get(s.scenes[sc].camera, ProjectionType)
	if err != nil {
		return err
	}
	p := ec.(Projection)

	ec, err = s.ents.Get(s.scenes[sc].camera, TransformationType)
	if err != nil {
		return err
	}
	t := ec.(Transformation)

	projScreenMatrix := p.Matrix.Mul4(t.MatrixWorld().Inv())
	frustum := Mat4ToFrustum(projScreenMatrix)
	// fetch all objects visible in frustum
	opaque, transparent := s.visibleEntities(frustum, t.Position, s.scenes[sc].drawable)

	// opaque pass (front-to-back order)
	s.context.MainThread(func() {
		gl.Disable(gl.BLEND)

		for _, e := range opaque {
			s.renderEntity(e, s.scenes[sc].camera)
		}
	})

	// transparent pass (back-to-front order)
	s.context.MainThread(func() {
		gl.Enable(gl.BLEND)
		gl.BlendEquationSeparate(gl.FUNC_ADD, gl.FUNC_ADD)
		gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

		for _, e := range transparent {
			s.renderEntity(e, s.scenes[sc].camera)
		}
	})

	// swap buffers
	s.context.Update()
	return nil
}

func (s *RenderSystem) setClearColor(color mgl32.Vec3, alpha float64) {
	s.context.MainThread(func() {
		gl.ClearColor(gl.GLclampf(color[0]), gl.GLclampf(color[1]), gl.GLclampf(color[2]), gl.GLclampf(alpha))
	})
}

// TODO: replace with spatial system
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

func (s *RenderSystem) visibleEntities(frustum Frustum, cp mgl32.Vec3, drawable []Entity) (opaque, transparent []Entity) {
	opaque = make([]Entity, 0)
	transparent = make([]Entity, 0)
	var err error
	var ec Component

	zorder := map[Entity]float32{}
	cp4 := mgl32.Vec4{cp[0], cp[1], cp[2], 0} // TODO

	for _, e := range drawable {
		ec, err = s.ents.Get(e, TransformationType)
		if err != nil {
			continue
		}
		t := ec.(Transformation)

		ec, err = s.ents.Get(e, GeometryType)
		if err != nil {
			continue
		}
		g := ec.(Geometry)

		ec, err = s.ents.Get(e, MaterialType)
		if err != nil {
			continue
		}
		m := ec.(Material)

		c, r := g.Bounding.Sphere()
		c = t.MatrixWorld().Mul4x1(c)
		r *= t.MatrixWorld().MaxScale()

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

func (s *RenderSystem) renderEntity(object, camera Entity) error {
	ec, err := s.ents.Get(object, MaterialType)
	if err != nil {
		return err
	}
	material := ec.(Material)

	ec, err = s.ents.Get(object, GeometryType)
	if err != nil {
		return err
	}
	geometry := ec.(Geometry)

	ec, err = s.ents.Get(camera, ProjectionType)
	if err != nil {
		return err
	}
	projection := ec.(Projection)

	ec, err = s.ents.Get(camera, TransformationType)
	if err != nil {
		return err
	}
	cameratransform := ec.(Transformation)

	ec, err = s.ents.Get(object, TransformationType)
	if err != nil {
		return err
	}
	objecttransform := ec.(Transformation)

	return s.render(
		objecttransform, material, geometry,
		cameratransform, projection)
}

func (s *RenderSystem) render(
	objecttransform Transformation,
	material Material,
	geometry Geometry,

	cameratransform Transformation,
	projection Projection) error {

	// ### bind material
	var updateAttributes bool
	if material.program != s.currentProgram {
		// unbind old material (Textures)

		s.currentProgram = material.program
		s.currentProgram.program.Use()

		updateAttributes = true
	}

	// TODO: caching, unbind only if not needed by new material
	s.unbindTextures()

	// ### bind geometry
	if geometry.mesh != s.currentGeometry || updateAttributes {
		s.currentGeometry = geometry.mesh

		// disable old attributes
		s.currentProgram.DisableAttributes()
		s.currentGeometry.VertexArrayObject.Bind()

		// vertices
		s.currentGeometry.PositionBuffer.Bind(gl.ARRAY_BUFFER)
		//program.EnableAttribute("vertexPosition")
		//program.Attribute("vertexPosition").AttribPointer(3, gl.FLOAT, false, 0, nil)
		s.currentProgram.EnableAttribute("vertexPosition")
		//geometry.positionBuffer.Unbind(gl.ARRAY_BUFFER)

		// normal
		s.currentGeometry.NormalBuffer.Bind(gl.ARRAY_BUFFER)
		//program.EnableAttribute("vertexNormal")
		//program.Attribute("vertexNormal").AttribPointer(3, gl.FLOAT, false, 0, nil)
		s.currentProgram.EnableAttribute("vertexNormal")

		// uv
		s.currentGeometry.UvBuffer.Bind(gl.ARRAY_BUFFER)
		//program.EnableAttribute("vertexUV")
		//program.Attribute("vertexUV").AttribPointer(2, gl.FLOAT, false, 0, nil)
		s.currentProgram.EnableAttribute("vertexUV")
	}

	// ### set matrices

	// update projection uniform

	s.UpdateUniform("projectionMatrix", projection.Matrix /*.Float32()*/)

	// viewMatrix

	viewMatrix := cameratransform.MatrixWorld().Inv()
	//program.Uniform("viewMatrix").UniformMatrix4fv(false, viewMatrix.Float32())
	s.UpdateUniform("viewMatrix", viewMatrix /*.Float32()*/)

	// material update uniforms model/view/normal/projection-matrix

	// Model matrix : an identity matrix (model will be at the origin)
	//program.Uniform("modelMatrix").UniformMatrix4fv(false, m.MatrixWorld().Float32())
	s.UpdateUniform("modelMatrix", objecttransform.MatrixWorld() /*.Float32()*/)

	// modelViewMatrix
	modelViewMatrix := viewMatrix.Mul4(objecttransform.MatrixWorld())
	//program.Uniform("modelViewMatrix").UniformMatrix4fv(false, modelViewMatrix.Float32())
	s.UpdateUniform("modelViewMatrix", modelViewMatrix /*.Float32()*/)

	// normalMatrix
	normalMatrix := modelViewMatrix.Normal()
	//program.Uniform("normalMatrix").UniformMatrix3fv(false, normalMatrix.Matrix3Float32())
	s.UpdateUniform("normalMatrix", normalMatrix /*.Matrix3Float32()*/)

	// update material uniforms
	for n, v := range material.uniforms {
		if err := s.UpdateUniform(n, v); err != nil {
			return err
		}
	}

	// ### draw
	s.currentGeometry.FaceBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
	gl.DrawElements(gl.TRIANGLES, s.currentGeometry.FaceCount*3, gl.UNSIGNED_SHORT, nil /* uintptr(start) */) // gl.UNSIGNED_INT, UNSIGNED_SHORT

	return nil
}

func (s *RenderSystem) bindTexture(buffer gl.Texture) int {
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

func (s *RenderSystem) unbindTextures() {
	for _, buffer := range s.currentTextures {
		buffer.Unbind(gl.TEXTURE_2D)
	}
	s.currentTextures = s.currentTextures[:0]
}

func (s *RenderSystem) UpdateUniform(name string, value interface{}) error {
	if _, found := s.currentProgram.uniforms[name]; !found {
		return fmt.Errorf("unsupported uniform: %v", name)
	}

	switch t := value.(type) {
	default:
		return fmt.Errorf("%v has unknown type: %T", name, t)

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
