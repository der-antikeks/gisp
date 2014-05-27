package game

import (
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/der-antikeks/gisp/math"

	"github.com/go-gl/gl"
)

type RenderSystem struct {
	engine *Engine
	prio   SystemPriority
	wm     *WindowManager

	drawChan, camChan, updChan chan Message

	scenes map[string]struct {
		drawable []Entity
		lights   []Entity
		camera   Entity
	}

	currentGeometry *meshbuffer
	currentProgram  *shaderprogram
	currentTextures []gl.Texture // usedTextureUnits
}

func NewRenderSystem(engine *Engine, wm *WindowManager) *RenderSystem {
	s := &RenderSystem{
		engine: engine,
		prio:   PriorityRender,
		wm:     wm,

		drawChan: make(chan Message),
		camChan:  make(chan Message),
		updChan:  make(chan Message),

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
						sc.camera = -1
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
	s.engine.Subscribe(Filter{
		Types: []MessageType{UpdateMessageType},
	}, s.prio, s.updChan)

	s.engine.Subscribe(Filter{
		Types:  []MessageType{EntityAddMessageType, EntityRemoveMessageType},
		Aspect: []ComponentType{TransformationType, GeometryType, MaterialType, SceneTreeType},
	}, s.prio, s.drawChan)

	s.engine.Subscribe(Filter{
		Types:  []MessageType{EntityAddMessageType, EntityRemoveMessageType},
		Aspect: []ComponentType{TransformationType, ProjectionType, SceneTreeType},
	}, s.prio, s.camChan)
}

func (s *RenderSystem) Stop() {
	s.engine.Unsubscribe(Filter{
		Types: []MessageType{UpdateMessageType},
	}, s.updChan)

	s.engine.Unsubscribe(Filter{
		Types:  []MessageType{EntityAddMessageType, EntityRemoveMessageType},
		Aspect: []ComponentType{TransformationType, GeometryType, MaterialType, SceneTreeType},
	}, s.drawChan)

	s.engine.Unsubscribe(Filter{
		Types:  []MessageType{EntityAddMessageType, EntityRemoveMessageType},
		Aspect: []ComponentType{TransformationType, ProjectionType, SceneTreeType},
	}, s.camChan)

	//s.drawable = []Entity{}
	//s.camera = -1
	// TODO: empty scenes?
}

func (s *RenderSystem) getScene(e Entity) string {
	ec, err := s.engine.Get(e, SceneTreeType)
	if err != nil {
		return ""
	}
	return ec.(SceneTree).Name
}

func (s *RenderSystem) Update(delta time.Duration) error {
	for n := range s.scenes {
		if err := s.updateScene(delta, n); err != nil {
			return err
		}
	}
	return nil
}

func (s *RenderSystem) updateScene(delta time.Duration, sc string) error {
	color := math.Color{0, 0, 0}
	alpha := 1.0
	s.setClearColor(color, alpha)

	// w, h := s.wm.Size()
	// gl.Viewport(0, 0, w, h) TODO: already set in WindowManager onResize(), must be changed with frambuffer?

	// TODO: clearing should depend on rendertarget
	clear := true
	if clear {
		MainThread(func() {
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		})
	}

	// TODO: move rendertarget to camera? or camera like interface (Viewport(w,h), Projection, etc.)
	if s.scenes[sc].camera == -1 {
		return fmt.Errorf("no camera found for RenderSystem")
	}
	// update scene matrix (all objects)
	// update camera matrix if not child of scene
	// calculate frustum of camera
	ec, err := s.engine.Get(s.scenes[sc].camera, ProjectionType)
	if err != nil {
		return err
	}
	p := ec.(Projection)

	ec, err = s.engine.Get(s.scenes[sc].camera, TransformationType)
	if err != nil {
		return err
	}
	t := ec.(Transformation)

	projScreenMatrix := p.Matrix.Mul(t.MatrixWorld().Inverse())
	frustum := math.FrustumFromMatrix(projScreenMatrix)
	// fetch all objects visible in frustum
	opaque, transparent := s.visibleEntities(frustum, t.Position, s.scenes[sc].drawable)

	// opaque pass (front-to-back order)
	MainThread(func() {
		gl.Disable(gl.BLEND)

		for _, e := range opaque {
			s.renderEntity(e, s.scenes[sc].camera)
		}
	})

	// transparent pass (back-to-front order)
	MainThread(func() {
		gl.Enable(gl.BLEND)
		gl.BlendEquationSeparate(gl.FUNC_ADD, gl.FUNC_ADD)
		gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

		for _, e := range transparent {
			s.renderEntity(e, s.scenes[sc].camera)
		}
	})

	// swap buffers
	s.wm.Update()
	return nil
}

func (s *RenderSystem) setClearColor(color math.Color, alpha float64) {
	MainThread(func() {
		gl.ClearColor(gl.GLclampf(color.R), gl.GLclampf(color.G), gl.GLclampf(color.B), gl.GLclampf(alpha))
	})
}

type byZ struct {
	entities []Entity
	zorder   map[Entity]float64
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

func (s *RenderSystem) visibleEntities(frustum math.Frustum, cp math.Vector, drawable []Entity) (opaque, transparent []Entity) {
	opaque = make([]Entity, 0)
	transparent = make([]Entity, 0)
	var err error
	var ec Component

	zorder := map[Entity]float64{}

	for _, e := range drawable {
		ec, err = s.engine.Get(e, TransformationType)
		if err != nil {
			continue
		}
		t := ec.(Transformation)

		ec, err = s.engine.Get(e, GeometryType)
		if err != nil {
			continue
		}
		g := ec.(Geometry)

		ec, err = s.engine.Get(e, MaterialType)
		if err != nil {
			continue
		}
		m := ec.(Material)

		c, r := g.Bounding.Sphere()
		c = t.MatrixWorld().Transform(c)
		r *= t.MatrixWorld().MaxScaleOnAxis()

		if frustum.IntersectsSphere(c, r) {
			zorder[e] = c.Sub(cp).Length()

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
	ec, err := s.engine.Get(object, MaterialType)
	if err != nil {
		return err
	}
	material := ec.(Material)

	ec, err = s.engine.Get(object, GeometryType)
	if err != nil {
		return err
	}
	geometry := ec.(Geometry)

	ec, err = s.engine.Get(camera, ProjectionType)
	if err != nil {
		return err
	}
	projection := ec.(Projection)

	ec, err = s.engine.Get(camera, TransformationType)
	if err != nil {
		return err
	}
	cameratransform := ec.(Transformation)

	ec, err = s.engine.Get(object, TransformationType)
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

	s.UpdateUniform("projectionMatrix", projection.Matrix.Float32())

	// viewMatrix

	viewMatrix := cameratransform.MatrixWorld().Inverse()
	//program.Uniform("viewMatrix").UniformMatrix4fv(false, viewMatrix.Float32())
	s.UpdateUniform("viewMatrix", viewMatrix.Float32())

	// material update uniforms model/view/normal/projection-matrix

	// Model matrix : an identity matrix (model will be at the origin)
	//program.Uniform("modelMatrix").UniformMatrix4fv(false, m.MatrixWorld().Float32())
	s.UpdateUniform("modelMatrix", objecttransform.MatrixWorld().Float32())

	// modelViewMatrix
	modelViewMatrix := viewMatrix.Mul(objecttransform.MatrixWorld())
	//program.Uniform("modelViewMatrix").UniformMatrix4fv(false, modelViewMatrix.Float32())
	s.UpdateUniform("modelViewMatrix", modelViewMatrix.Float32())

	// normalMatrix
	normalMatrix := modelViewMatrix.Normal()
	//program.Uniform("normalMatrix").UniformMatrix3fv(false, normalMatrix.Matrix3Float32())
	s.UpdateUniform("normalMatrix", normalMatrix.Matrix3Float32())

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

	case Texture:
		s.currentProgram.uniforms[name].location.Uniform1i(s.bindTexture(t.buffer))

	case int:
		s.currentProgram.uniforms[name].location.Uniform1i(t)
	case float64:
		s.currentProgram.uniforms[name].location.Uniform1f(float32(t))
	case float32:
		s.currentProgram.uniforms[name].location.Uniform1f(t)

	case [16]float32:
		s.currentProgram.uniforms[name].location.UniformMatrix4fv(false, t)
	case [9]float32:
		s.currentProgram.uniforms[name].location.UniformMatrix3fv(false, t)

	case math.Color:
		s.currentProgram.uniforms[name].location.Uniform3f(float32(t.R), float32(t.G), float32(t.B))

	case math.Vector:
		s.currentProgram.uniforms[name].location.Uniform3f(float32(t[0]), float32(t[1]), float32(t[2]))

	case bool:
		if t {
			s.currentProgram.uniforms[name].location.Uniform1i(1)
		} else {
			s.currentProgram.uniforms[name].location.Uniform1i(0)
		}
	}

	return nil
}
