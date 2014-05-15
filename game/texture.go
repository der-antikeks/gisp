package game

import (
	"github.com/go-gl/gl"
)

type Texture struct {
	File string // loading from entitymanager?

	texture gl.Texture
	w, h    int
}

//TODO: binding/unbinding from rendersystem
func (t Texture) Bind(slot int) {} // MainThread(func() {})
func (t Texture) Unbind()       {}
func (t Texture) Dispose()      {}
