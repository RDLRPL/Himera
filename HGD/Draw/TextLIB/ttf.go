package TextLIB

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func InitFont() error {
	fontBytes, err := ioutil.ReadFile("HGD/ttf/Hasklig.ttf")
	if err != nil {
		return fmt.Errorf("read font ? %v", err)
	}

	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return fmt.Errorf("parse font ? %v", err)
	}

	Characters = make(map[rune]*Character)

	fontFace = truetype.NewFace(f, &truetype.Options{
		Size:    FontSize,
		DPI:     Dpi,
		Hinting: font.HintingFull,
	})

	FontMetrics = fontFace.Metrics()

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
			if err := CreateCharacterTexture(fontFace, ch); err != nil {
				fmt.Printf("Char img ? %c: %v\n", ch, err)
			}
		}
	}

	if Characters[rune('?')] == nil {
		CreateCharacterTexture(fontFace, '?')
	}

	return nil
}

func CreateCharacterTexture(face font.Face, ch rune) error {
	bounds, advance, ok := face.GlyphBounds(ch)
	if !ok {
		return fmt.Errorf("glyph ? %c", ch)
	}

	glyphWidth := int(bounds.Max.X-bounds.Min.X) >> 6
	glyphHeight := int(bounds.Max.Y-bounds.Min.Y) >> 6

	bearingX := int(bounds.Min.X >> 6)
	bearingY := int(bounds.Max.Y >> 6)

	if glyphWidth <= 0 {
		glyphWidth = 1
	}
	if glyphHeight <= 0 {
		glyphHeight = 1
	}

	img := image.NewRGBA(image.Rect(0, 0, glyphWidth, glyphHeight))

	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 0}}, image.Point{}, draw.Src)

	drawer := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{color.RGBA{255, 255, 255, 255}},
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
		gl.RGBA,
		int32(glyphWidth),
		int32(glyphHeight),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(img.Pix),
	)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	Characters[ch] = &Character{
		TextureID: texture,
		Size:      [2]int32{int32(glyphWidth), int32(glyphHeight)},
		Bearing:   [2]int32{int32(bearingX), int32(bearingY)},
		Advance:   int32(advance >> 6),
	}

	return nil
}
