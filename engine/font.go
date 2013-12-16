package engine

import (
	"image"
	"image/color"
	"io/ioutil"
	m "math"

	"github.com/der-antikeks/gisp/math"

	"code.google.com/p/freetype-go/freetype"
	"github.com/go-gl/gl"
)

type Glyph struct {
	x, y    int
	w, h    int
	advance int
}

// http://www.valvesoftware.com/publications/2007/SIGGRAPH2007_AlphaTestedMagnification.pdf
type Font struct {
	material *Material
	charset  map[rune]Glyph
}

func LoadFont(fontfile string /*, size, low, high int*/) (*Font, error) {
	dpi := 72.0           // screen resolution in dots per inch
	size := 32.0          // font size in points
	low, high := 32, 127  // lower, upper rune limits
	spread := 4           // signed distance radius
	padding := spread + 1 // padding between glyphs

	// load truetype into image
	glyphs, img, err := loadTruetype(fontfile, dpi, size, low, high, padding)
	if err != nil {
		return nil, err
	}

	// generate distance field
	img = generateDistanceField(img, glyphs, spread)

	// generate texture
	tex := &ImageTexture{
		image: []*image.RGBA{img},

		wrapS:     gl.REPEAT,
		wrapT:     gl.REPEAT,
		magFilter: gl.LINEAR,
		minFilter: gl.LINEAR_MIPMAP_LINEAR,

		needsUpdate: true,
	}

	// load material
	mat, err := NewMaterial("font")
	if err != nil {
		return nil, err
	}
	mat.SetUniform("distanceFieldMap", tex)
	mat.SetUniform("diffuse", math.Color{1, 0, 1})

	return &Font{
		material: mat,
		charset:  glyphs,
	}, nil
}

func (f *Font) Dispose() {
	if f.material != nil {
		f.material.Dispose()
	}
}

func (f *Font) Printf(format string, a ...interface{}) *Mesh {
	// TODO
	var geo *Geometry
	geo = NewPlaneGeometry(50, 50)

	return &Mesh{
		geometry: geo,
		material: f.material,

		up:    math.Vector{0, 1, 0},
		scale: math.Vector{1, 1, 1},

		matrixNeedsUpdate:      true,
		matrixWorldNeedsUpdate: true,
	}
}

func loadTruetype(fontfile string, dpi, size float64, low, high, padding int) (map[rune]Glyph, *image.RGBA, error) {
	// read font data
	fontBytes, err := ioutil.ReadFile(fontfile)
	if err != nil {
		return nil, nil, err
	}

	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		return nil, nil, err
	}

	// create image
	glyphCounts := high - low + 1
	glyphsPerRow := int(m.Ceil(m.Sqrt(float64(glyphCounts))))
	glyphsPerCol := (glyphCounts / glyphsPerRow) + 1

	glyphBounds := font.Bounds(int32(size))
	glyphWidth := int(glyphBounds.XMax-glyphBounds.XMin) + padding*2
	glyphHeight := int(glyphBounds.YMax-glyphBounds.YMin) + padding*2

	rgba := image.NewRGBA(image.Rect(0, 0,
		math.NextHighestPowerOfTwo(glyphWidth*glyphsPerRow),
		math.NextHighestPowerOfTwo(glyphHeight*glyphsPerCol)))
	//draw.Draw(rgba, rgba.Bounds(), image.Black, image.ZP, draw.Src)

	// initialize context
	c := freetype.NewContext()
	c.SetDPI(dpi)
	c.SetFont(font)
	c.SetFontSize(size)
	c.SetClip(rgba.Bounds())
	c.SetDst(rgba)
	c.SetSrc(image.White)

	// draw runes
	glyphs := make(map[rune]Glyph)
	var glyphNum, glyphX, glyphY int

	for ch := low; ch <= high; ch++ {
		glyphX = (glyphNum % glyphsPerRow) * glyphWidth
		glyphY = (glyphNum / glyphsPerRow) * glyphHeight

		metric := font.HMetric(int32(size), font.Index(rune(ch)))
		advance := int(metric.AdvanceWidth) + padding

		glyphs[rune(ch)] = Glyph{glyphX, glyphY, glyphWidth, glyphHeight, advance}

		pt := freetype.Pt(glyphX+padding, glyphY+padding+int(c.PointToFix32(size)>>8))
		if _, err = c.DrawString(string(ch), pt); err != nil {
			return nil, nil, err
		}

		glyphNum++
	}

	return glyphs, rgba, nil
}

func generateDistanceField(in *image.RGBA, glyphs map[rune]Glyph, spread int) *image.RGBA {
	out := image.NewRGBA(in.Bounds())

	for _, glyph := range glyphs {
		// create mask
		mask := make([][]bool, glyph.h)

		for y := 0; y < glyph.h; y++ {
			if mask[y] == nil {
				mask[y] = make([]bool, glyph.w)
			}

			for x := 0; x < glyph.w; x++ {
				r, g, b, a := in.At(glyph.x+x, glyph.y+y).RGBA()
				mask[y][x] = (r >= 0x7fff || g >= 0x7fff || b >= 0x7fff) && (a >= 0x7fff)
			}
		}

		// find signed distance
		for y := 0; y < glyph.h; y++ {
			for x := 0; x < glyph.w; x++ {
				out.Set(
					glyph.x+x, glyph.y+y,
					distanceToColor(
						findSignedDistance(x, y, spread, mask),
						spread))
			}
		}
	}

	return out
}

func findSignedDistance(centerX, centerY, delta int, mask [][]bool) float64 {
	width := len(mask[0])
	height := len(mask)
	base := mask[centerY][centerX]

	startX, endX := int(m.Max(0, float64(centerX-delta))), int(m.Min(float64(centerX+delta), float64(width-1)))
	startY, endY := int(m.Max(0, float64(centerY-delta))), int(m.Min(float64(centerY+delta), float64(height-1)))

	closestSquareDist := delta * delta

	for y := startY; y <= endY; y++ {
		for x := startX; x <= endX; x++ {
			if base != mask[y][x] {
				squareDist := (centerX-x)*(centerX-x) + (centerY-y)*(centerY-y)
				if squareDist < closestSquareDist {
					closestSquareDist = squareDist
				}
			}
		}
	}

	closestDist := m.Sqrt(float64(closestSquareDist))
	if base {
		return m.Min(closestDist, float64(delta))
	}
	return -m.Min(closestDist, float64(delta))
}

func distanceToColor(distance float64, spread int) color.Color {
	c := uint8(m.Min(1, m.Max(0, 0.5+0.5*(distance/float64(spread)))) * 0xff)
	return color.RGBA{c, c, c, c} // premultiplied alpha white
}
