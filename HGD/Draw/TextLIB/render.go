package TextLIB

import (
	"fmt"
	"image"
	"image/color"
	"io/ioutil"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

var FontMetrics font.Metrics

const (
	FontSize = 14.0
	Dpi      = 96.0
)

type Character struct {
	TextureID uint32
	Size      [2]int32
	Bearing   [2]int32
	Advance   int32
}

var Characters map[rune]*Character

func DrawText(program uint32, text string, x, y float32, scale float32, color [3]float32) {
	gl.UseProgram(program)
	gl.ActiveTexture(gl.TEXTURE0)

	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.GenBuffers(1, &vbo)
	gl.BindVertexArray(vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 24*4, nil, gl.DYNAMIC_DRAW)

	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 0, nil)
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	colorLoc := gl.GetUniformLocation(program, gl.Str("textColor\x00"))
	gl.Uniform3f(colorLoc, color[0], color[1], color[2])

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	baseline := y

	for _, ch := range text {
		char := Characters[ch]
		if char == nil {
			char = Characters[rune('?')]
			if char == nil {
				continue
			}
		}

		xpos := x + float32(char.Bearing[0])*scale
		ypos := baseline - float32(char.Bearing[1])*scale

		w := float32(char.Size[0]) * scale
		h := float32(char.Size[1]) * scale

		vertices := []float32{
			xpos, ypos + h, 0.0, 1.0,
			xpos, ypos, 0.0, 0.0,
			xpos + w, ypos, 1.0, 0.0,

			xpos, ypos + h, 0.0, 1.0,
			xpos + w, ypos, 1.0, 0.0,
			xpos + w, ypos + h, 1.0, 1.0,
		}

		gl.BindTexture(gl.TEXTURE_2D, char.TextureID)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(vertices)*4, gl.Ptr(vertices))

		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		x += float32(char.Advance) * scale
	}

	gl.BindVertexArray(0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.DeleteVertexArrays(1, &vao)
	gl.DeleteBuffers(1, &vbo)
}

func InitFont() error {
	fontBytes, err := ioutil.ReadFile("HGD/ttf/arial.ttf")
	if err != nil {
		return fmt.Errorf("failed to read font file: %v", err)
	}

	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return fmt.Errorf("failed to parse font: %v", err)
	}

	Characters = make(map[rune]*Character)

	face := truetype.NewFace(f, &truetype.Options{
		Size:    FontSize,
		DPI:     Dpi,
		Hinting: font.HintingFull,
	})

	FontMetrics = face.Metrics()

	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)

	ranges := [][2]rune{
		{32, 126},
		{160, 255},
		{1040, 1103},
		{1025, 1025},
		{1105, 1105},
	}

	for _, r := range ranges {
		for ch := r[0]; ch <= r[1]; ch++ {
			if err := createCharacterTexture(face, ch); err != nil {
				fmt.Printf("Warning: failed to create texture for character %c: %v\n", ch, err)
			}
		}
	}

	if Characters[rune('?')] == nil {
		createCharacterTexture(face, '?')
	}

	return nil
}

func createCharacterTexture(face font.Face, ch rune) error {
	bounds, advance, ok := face.GlyphBounds(ch)
	if !ok {
		return fmt.Errorf("glyph not found for character %c", ch)
	}

	w := int((bounds.Max.X - bounds.Min.X) >> 6)
	h := int((bounds.Max.Y - bounds.Min.Y) >> 6)

	if w <= 0 {
		w = int(FontSize / 2)
	}
	if h <= 0 {
		h = int(FontSize)
	}

	img := image.NewGray(image.Rect(0, 0, w, h))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetGray(x, y, color.Gray{0})
		}
	}

	drawer := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{color.Gray{255}},
		Face: face,
		Dot: fixed.Point26_6{
			X: -bounds.Min.X,
			Y: -bounds.Min.Y,
		},
	}

	drawer.DrawString(string(ch))

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RED,
		int32(w),
		int32(h),
		0,
		gl.RED,
		gl.UNSIGNED_BYTE,
		gl.Ptr(img.Pix),
	)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	Characters[ch] = &Character{
		TextureID: texture,
		Size:      [2]int32{int32(w), int32(h)},
		Bearing:   [2]int32{int32(bounds.Min.X >> 6), int32(bounds.Max.Y >> 6)},
		Advance:   int32(advance >> 6),
	}

	return nil
}

func GetTextDimensions(text string, scale float32) (width, height float32) {
	if FontMetrics.Height == 0 {
		return 0, 0
	}

	width = 0
	height = float32(FontMetrics.Height>>6) * scale

	for _, ch := range text {
		char := Characters[ch]
		if char != nil {
			width += float32(char.Advance) * scale
		}
	}

	return width, height
}
