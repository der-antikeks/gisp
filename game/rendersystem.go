package game

import (
	"fmt"
	"time"

	"github.com/der-antikeks/gisp/ecs"
	"github.com/der-antikeks/gisp/math"

	"github.com/go-gl/gl"
)

type RenderSystem struct {
	wm       *WindowManager
	drawable *ecs.Collection
	camera   *ecs.Collection
}

func NewRenderSystem(wm *WindowManager) ecs.System {
	return &RenderSystem{
		wm: wm,
	}
}

func (s *RenderSystem) AddedToEngine(e *ecs.Engine) error {
	s.drawable = e.Collection(TransformationType, GeometryType, MaterialType)
	s.camera = e.Collection(TransformationType, ProjectionType)
	return nil
}
func (s *RenderSystem) RemovedFromEngine(e *ecs.Engine) error {
	return nil
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
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	}

	// TODO: move rendertarget to camera? or camera like interface (Viewport(w,h), Projection, etc.)
	camera := s.camera.First()
	if camera == nil {
		return fmt.Errorf("no camera found for RenderSystem")
	}
	// update scene matrix (all objects)
	// update camera matrix if not child of scene
	// calculate frustum of camera
	p := camera.Get(ProjectionType).(*Projection)
	t := camera.Get(TransformationType).(*Transformation)
	projScreenMatrix := p.ProjectionMatrix().Mul(t.MatrixWorld().Inverse())
	frustum := math.FrustumFromMatrix(projScreenMatrix)
	// fetch all objects visible in frustum
	opaque, transparent := s.visibleEntities(frustum)

	// opaque pass (front-to-back order)
	gl.Disable(gl.BLEND)
	for _, e := range opaque {
		s.renderEntity(e, camera)
	}

	// transparent pass (back-to-front order)
	gl.Enable(gl.BLEND)
	gl.BlendEquationSeparate(gl.FUNC_ADD, gl.FUNC_ADD)
	gl.BlendFuncSeparate(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA, gl.ONE, gl.ONE_MINUS_SRC_ALPHA)

	for _, e := range transparent {
		s.renderEntity(e, camera)
	}

	// swap buffers
	s.wm.Update()
	return nil
}

func (s *RenderSystem) setClearColor(color math.Color, alpha float64) {
	gl.ClearColor(gl.GLclampf(color.R), gl.GLclampf(color.G), gl.GLclampf(color.B), gl.GLclampf(alpha))
}

func (s *RenderSystem) visibleEntities(frustum math.Frustum) (opaque, transparent []*ecs.Entity) {
	entities := s.drawable.Entities()

	opaque = make([]*ecs.Entity, len(entities))
	transparent = make([]*ecs.Entity, len(entities))
	var cntOp, cntTr int

	for _, e := range entities {
		t := e.Get(TransformationType).(*Transformation)
		g := e.Get(GeometryType).(*Geometry)
		m := e.Get(MaterialType).(*Material)
		c, r := g.Bounding.Sphere()

		if frustum.IntersectsSphere(t.MatrixWorld().Transform(c), r*t.MatrixWorld().MaxScaleOnAxis()) {
			if m.Opaque {
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

func (s *RenderSystem) renderEntity(object, camera *ecs.Entity) {
	// bind material
	m := object.Get(MaterialType).(*Material)
	// unbind old material if not equals
	// enable shader program if not already by previous material (could have same program but different uniforms (texture))
	// update uniforms

	// bind geometry
	g := object.Get(GeometryType).(*Geometry)
	// if old geometry not equals, disable all material buffers
	// bind each geometry buffer
	// enable material attributes for each

	// set matrices
	t := object.Get(TransformationType).(*Transformation)
	// material update uniforms model/view/normal/projection-matrix

	// draw
	// bind geometry face buffer
	faceCount := 0 // geometry.FaceCount()

	gl.DrawElements(gl.TRIANGLES, faceCount, gl.UNSIGNED_SHORT, nil /* uintptr(start) */) // gl.UNSIGNED_INT, UNSIGNED_SHORT

	fmt.Println(m, g, t)
}
