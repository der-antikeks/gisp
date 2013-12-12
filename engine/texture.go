package engine

import (
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/go-gl/gl"
)

type Texture interface {
	Bind(slot int)
	Unbind()
	Dispose()
}

type ImageTexture struct {
	Texture

	buffer      gl.Texture
	initialized bool

	image                []*image.RGBA
	wrapS, wrapT         int
	magFilter, minFilter int
	needsUpdate          bool
}

func LoadTexture(path string) (*ImageTexture, error) {
	// load file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// decode png
	im, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	t := &ImageTexture{
		image: make([]*image.RGBA, 1),

		wrapS:     gl.REPEAT,
		wrapT:     gl.REPEAT,
		magFilter: gl.LINEAR,
		minFilter: gl.LINEAR_MIPMAP_LINEAR,

		needsUpdate: true,
	}
	t.image[0] = image.NewRGBA(im.Bounds())

	// convert to rgba
	draw.Draw(t.image[0], im.Bounds(), im, image.Pt(0, 0), draw.Src)

	return t, nil
}

func LoadCompressedTexture(path string) (*ImageTexture, error) { return nil, nil }

// init texture buffers
func (t *ImageTexture) init() {
	t.buffer = gl.GenTexture()

	t.initialized = true
}

// update image and gl parameters
func (t *ImageTexture) update() {
	if !t.initialized {
		t.init()
	}

	// bind buffer
	t.buffer.Bind(gl.TEXTURE_2D)

	// set texture parameters
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, t.wrapS)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, t.wrapT)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, t.magFilter)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, t.minFilter)

	// give image(s) to opengl
	for level, img := range t.image {
		gl.TexImage2D(gl.TEXTURE_2D, level, gl.RGBA,
			img.Bounds().Dx(), img.Bounds().Dy(),
			0, gl.RGBA, gl.UNSIGNED_BYTE, img.Pix)
	}

	// generate mipmaps
	if len(t.image) == 1 {
		gl.GenerateMipmap(gl.TEXTURE_2D)
	}

	t.needsUpdate = false
}

// cleanup
func (t *ImageTexture) Dispose() {
	if t.buffer != 0 {
		t.buffer.Delete()
	}
}

// bind texture in Texture Unit slot
func (t *ImageTexture) Bind(slot int) {
	if t.needsUpdate {
		t.update()
	} else {
		t.buffer.Bind(gl.TEXTURE_2D)
	}

	gl.ActiveTexture(gl.TEXTURE0 + gl.GLenum(slot))
}

func (t *ImageTexture) Unbind() {
	if t.initialized {
		t.buffer.Unbind(gl.TEXTURE_2D)
	}
}
