package game

import (
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/go-gl/gl"
)

type Texture struct {
	File string // loading from entitymanager?

	buffer gl.Texture
	w, h   int
}

// TODO: binding/unbinding from rendersystem
// MainThread(func() {})

// bind texture in Texture Unit slot
func (t Texture) Bind(slot int) {
	t.buffer.Bind(gl.TEXTURE_2D)
	gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(slot))
}

func (t Texture) Unbind() {
	t.buffer.Unbind(gl.TEXTURE_2D)
}

// cleanup
func (t Texture) Dispose() {
	if t.buffer != 0 {
		t.buffer.Delete()
	}
}

func LoadTexture(path string) (Texture, error) {
	// load file
	file, err := os.Open(path)
	if err != nil {
		return Texture{}, err
	}
	defer file.Close()

	// decode image
	im, _, err := image.Decode(file)
	if err != nil {
		return Texture{}, err
	}
	bounds := im.Bounds()

	// convert to rgba
	rgba, ok := im.(*image.RGBA)
	if !ok {
		rgba = image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, im, image.Pt(0, 0), draw.Src)
	}

	// create texture
	t := Texture{
		File: path,
		w:    bounds.Dx(),
		h:    bounds.Dy(),
	}

	MainThread(func() {
		t.buffer = gl.GenTexture()
		t.buffer.Bind(gl.TEXTURE_2D)

		// set texture parameters
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE) // gl.REPEAT
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE) // gl.REPEAT
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)

		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR) // gl.LINEAR_MIPMAP_LINEAR
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

		// give image(s) to opengl
		gl.TexImage2D(gl.TEXTURE_2D, 0 /*level*/, gl.RGBA,
			rgba.Bounds().Dx(), rgba.Bounds().Dy(),
			0, gl.RGBA, gl.UNSIGNED_BYTE, rgba.Pix)

		// generate mipmaps
		gl.GenerateMipmap(gl.TEXTURE_2D)

		t.buffer.Unbind(gl.TEXTURE_2D)
	})

	return t, nil
}

// TODO:
func NewFramebuffer(w, h int) Texture {
	t := Texture{
		buffer: gl.GenTexture(),
		w:      w,
		h:      h,
	}

	t.buffer.Bind(gl.TEXTURE_2D)
	{
		// set texture parameters
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE) // gl.REPEAT
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE) // gl.REPEAT
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_R, gl.CLAMP_TO_EDGE)

		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR) // gl.NEAREST
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR) // gl.NEAREST

		// create storage
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA,
			t.w, t.h,
			0, gl.RGBA, gl.UNSIGNED_BYTE, nil)

		// generate mipmaps
		gl.GenerateMipmap(gl.TEXTURE_2D)
	}
	t.buffer.Unbind(gl.TEXTURE_2D)

	return t
}
