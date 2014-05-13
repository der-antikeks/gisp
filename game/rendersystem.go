package game

import (
	"fmt"
	"log"
	"time"

	"github.com/der-antikeks/gisp/ecs"
	"github.com/der-antikeks/gisp/math"

	"github.com/go-gl/gl"
)

type RenderSystem struct {
	engine *ecs.Engine
	prio   ecs.SystemPriority
	wm     *WindowManager

	drawChan, camChan, updChan chan ecs.Message

	drawable []ecs.Entity
	camera   ecs.Entity
}

func NewRenderSystem(engine *ecs.Engine, wm *WindowManager) *RenderSystem {
	s := &RenderSystem{
		engine: engine,
		prio:   PriorityRender,
		wm:     wm,

		drawChan: make(chan ecs.Message),
		camChan:  make(chan ecs.Message),
		updChan:  make(chan ecs.Message),

		camera: -1,
	}

	go func() {
		s.Restart()

		for {
			select {
			case event := <-s.drawChan:
				switch e := event.(type) {
				case ecs.MessageEntityAdd:
					s.drawable = append(s.drawable, e.Added)
				case ecs.MessageEntityRemove:
					for i, f := range s.drawable {
						if f == e.Removed {
							s.drawable = append(s.drawable[:i], s.drawable[i+1:]...)
							break
						}
					}
				}

			case event := <-s.camChan:
				switch e := event.(type) {
				case ecs.MessageEntityAdd:
					s.camera = e.Added
				case ecs.MessageEntityRemove:
					if s.camera == e.Removed {
						s.camera = -1
					}
				}

			case event := <-s.updChan:
				switch e := event.(type) {
				case ecs.MessageUpdate:
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
	s.engine.Subscribe(ecs.Filter{
		Types: []ecs.MessageType{ecs.UpdateMessageType},
	}, s.prio, s.updChan)

	s.engine.Subscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{TransformationType, GeometryType, MaterialType},
	}, s.prio, s.drawChan)

	s.engine.Subscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{TransformationType, ProjectionType},
	}, s.prio, s.camChan)
}

func (s *RenderSystem) Stop() {
	s.engine.Unsubscribe(ecs.Filter{
		Types: []ecs.MessageType{ecs.UpdateMessageType},
	}, s.updChan)

	s.engine.Unsubscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{TransformationType, GeometryType, MaterialType},
	}, s.drawChan)

	s.engine.Unsubscribe(ecs.Filter{
		Types:  []ecs.MessageType{ecs.EntityAddMessageType, ecs.EntityRemoveMessageType},
		Aspect: []ecs.ComponentType{TransformationType, ProjectionType},
	}, s.camChan)

	s.drawable = []ecs.Entity{}
	s.camera = -1
}

func (s *RenderSystem) Update(delta time.Duration) error {
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
	if s.camera == -1 {
		return fmt.Errorf("no camera found for RenderSystem")
	}
	// update scene matrix (all objects)
	// update camera matrix if not child of scene
	// calculate frustum of camera
	ec, err := s.engine.Get(s.camera, ProjectionType)
	if err != nil {
		return err
	}
	p := ec.(Projection)

	ec, err = s.engine.Get(s.camera, TransformationType)
	if err != nil {
		return err
	}
	t := ec.(Transformation)

	projScreenMatrix := p.ProjectionMatrix().Mul(t.MatrixWorld().Inverse())
	frustum := math.FrustumFromMatrix(projScreenMatrix)
	// fetch all objects visible in frustum
	opaque, transparent := s.visibleEntities(frustum)

	// opaque pass (front-to-back order)
	MainThread(func() {
		gl.Disable(gl.BLEND)

		for _, e := range opaque {
			s.renderEntity(e, s.camera)
		}
	})

	// transparent pass (back-to-front order)
	MainThread(func() {
		gl.Enable(gl.BLEND)
		gl.BlendEquationSeparate(gl.FUNC_ADD, gl.FUNC_ADD)
		gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

		for _, e := range transparent {
			s.renderEntity(e, s.camera)
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

func (s *RenderSystem) visibleEntities(frustum math.Frustum) (opaque, transparent []ecs.Entity) {
	opaque = make([]ecs.Entity, len(s.drawable))
	transparent = make([]ecs.Entity, len(s.drawable))
	var cntOp, cntTr int
	var err error
	var ec ecs.Component

	for _, e := range s.drawable {
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

		if frustum.IntersectsSphere(t.MatrixWorld().Transform(c), r*t.MatrixWorld().MaxScaleOnAxis()) {
			if m.Opaque() {
				opaque[cntOp] = e
				cntOp++
			} else {
				transparent[cntTr] = e
				cntTr++
			}
		}
	}

	return opaque[:cntOp], transparent[:cntTr]
}

func (s *RenderSystem) renderEntity(object, camera ecs.Entity) error {
	// ### bind material
	ec, err := s.engine.Get(object, MaterialType)
	if err != nil {
		return err
	}
	material := ec.(Material)
	// unbind old material if not equals
	// enable shader program if not already by previous material (could have same program but different uniforms (texture))
	if !material.Shader.enabled {
		material.Shader.program.Use()
		material.Shader.enabled = true
	}

	// ### bind geometry
	ec, err = s.engine.Get(object, GeometryType)
	if err != nil {
		return err
	}
	geometry := ec.(Geometry)
	// if old geometry not equals, disable all material buffers
	// bind each geometry buffer
	// enable material attributes for each
	geometry.init()

	// disable old attributes
	material.Shader.DisableAttributes()
	geometry.VertexArrayObject.Bind()

	// vertices
	geometry.PositionBuffer.Bind(gl.ARRAY_BUFFER)
	//program.EnableAttribute("vertexPosition")
	//program.Attribute("vertexPosition").AttribPointer(3, gl.FLOAT, false, 0, nil)
	material.Shader.EnableAttribute("vertexPosition")
	//geometry.positionBuffer.Unbind(gl.ARRAY_BUFFER)

	// normal
	geometry.NormalBuffer.Bind(gl.ARRAY_BUFFER)
	//program.EnableAttribute("vertexNormal")
	//program.Attribute("vertexNormal").AttribPointer(3, gl.FLOAT, false, 0, nil)
	material.Shader.EnableAttribute("vertexNormal")

	// uv
	geometry.UvBuffer.Bind(gl.ARRAY_BUFFER)
	//program.EnableAttribute("vertexUV")
	//program.Attribute("vertexUV").AttribPointer(2, gl.FLOAT, false, 0, nil)
	material.Shader.EnableAttribute("vertexUV")

	// ### set matrices

	// update projection uniform
	ec, err = s.engine.Get(camera, ProjectionType)
	if err != nil {
		return err
	}
	projection := ec.(Projection)
	material.SetUniform("projectionMatrix", projection.ProjectionMatrix().Float32())

	// viewMatrix
	ec, err = s.engine.Get(camera, TransformationType)
	if err != nil {
		return err
	}
	cameratransform := ec.(Transformation)
	viewMatrix := cameratransform.MatrixWorld().Inverse()
	//program.Uniform("viewMatrix").UniformMatrix4fv(false, viewMatrix.Float32())
	material.SetUniform("viewMatrix", viewMatrix.Float32())

	ec, err = s.engine.Get(object, TransformationType)
	if err != nil {
		return err
	}
	transform := ec.(Transformation)
	// material update uniforms model/view/normal/projection-matrix

	// Model matrix : an identity matrix (model will be at the origin)
	//program.Uniform("modelMatrix").UniformMatrix4fv(false, m.MatrixWorld().Float32())
	material.SetUniform("modelMatrix", transform.MatrixWorld().Float32())

	// modelViewMatrix
	modelViewMatrix := viewMatrix.Mul(transform.MatrixWorld())
	//program.Uniform("modelViewMatrix").UniformMatrix4fv(false, modelViewMatrix.Float32())
	material.SetUniform("modelViewMatrix", modelViewMatrix.Float32())

	// normalMatrix
	normalMatrix := modelViewMatrix.Normal()
	//program.Uniform("normalMatrix").UniformMatrix3fv(false, normalMatrix.Matrix3Float32())
	material.SetUniform("normalMatrix", normalMatrix.Matrix3Float32())

	// update uniforms
	material.UpdateUniforms()

	// ### draw
	geometry.FaceBuffer.Bind(gl.ELEMENT_ARRAY_BUFFER)
	gl.DrawElements(gl.TRIANGLES, len(geometry.Faces)*3, gl.UNSIGNED_SHORT, nil /* uintptr(start) */) // gl.UNSIGNED_INT, UNSIGNED_SHORT

	return nil
}
