package engine

import (
	"fmt"

	"github.com/go-gl/gl"
)

type Uniform struct {
	location gl.UniformLocation
}

type Attribute struct {
	location gl.AttribLocation
	enabled  bool
}

type Program struct {
	program    gl.Program
	attributes map[string]*Attribute
	uniforms   map[string]*Uniform
}

func NewProgram(vertex, fragment string, attributes, uniforms []string) (*Program, error) {
	// vertex shader
	vshader := gl.CreateShader(gl.VERTEX_SHADER)
	vshader.Source(vertex)
	vshader.Compile()
	if vshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		return nil, fmt.Errorf("vertex shader error: %v", vshader.GetInfoLog())
	}
	defer vshader.Delete()

	// fragment shader
	fshader := gl.CreateShader(gl.FRAGMENT_SHADER)
	fshader.Source(fragment)
	fshader.Compile()
	if fshader.Get(gl.COMPILE_STATUS) != gl.TRUE {
		return nil, fmt.Errorf("fragment shader error: %v", fshader.GetInfoLog())
	}
	defer fshader.Delete()

	// program
	prg := &Program{
		program:    gl.CreateProgram(),
		attributes: make(map[string]*Attribute),
		uniforms:   make(map[string]*Uniform),
	}

	prg.program.AttachShader(vshader)
	prg.program.AttachShader(fshader)
	prg.program.Link()
	if prg.program.Get(gl.LINK_STATUS) != gl.TRUE {
		return nil, fmt.Errorf("linker error: %v", prg.program.GetInfoLog())
	}

	// locations
	for _, a := range attributes {
		prg.attributes[a] = &Attribute{
			location: prg.program.GetAttribLocation(a),
		}
	}

	for _, u := range uniforms {
		prg.uniforms[u] = &Uniform{
			location: prg.program.GetUniformLocation(u),
		}
	}

	return prg, nil
}

func (p *Program) Use() {
	p.program.Use()
}

func (p *Program) Dispose() {
	p.program.Delete()
}

// enable before changing
func (p *Program) EnableAttribute(name string) {
	if a, ok := p.attributes[name]; ok && !a.enabled {
		a.location.EnableArray()
		a.enabled = true
	} else {
		panic("unknown attribute: " + name)
	}
}

// disable before enabling others
func (p *Program) DisableAttributes() {
	for _, a := range p.attributes {
		if a.enabled {
			a.location.DisableArray()
			a.enabled = false
		}
	}
}

// set after changing active program
func (p *Program) Uniform(name string) gl.UniformLocation {
	return p.uniforms[name].location
}

func (p *Program) Attribute(name string) gl.AttribLocation {
	return p.attributes[name].location
}
