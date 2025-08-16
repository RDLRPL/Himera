package TextLIB

import (
	"github.com/go-gl/gl/v4.1-core/gl"
	"golang.org/x/image/font"
)

var FontMetrics font.Metrics

const (
	FontSize = 21.0
	Dpi      = 72.0
)

type Character struct {
	TextureID uint32
	Size      [2]int32
	Bearing   [2]int32
	Advance   int32
}

var Characters map[rune]*Character
var fontFace font.Face

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

	currentX := x

	for _, ch := range text {
		char := Characters[ch]
		if char == nil {
			char = Characters[0x00D8]
			if char == nil {
				continue
			}
		}

		xpos := currentX + float32(char.Bearing[0])*scale
		ypos := y - float32(char.Size[1])*scale + float32(char.Bearing[1])*scale

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

		currentX += float32(char.Advance) * scale
	}

	gl.BindVertexArray(0)
	gl.BindTexture(gl.TEXTURE_2D, 0)
	gl.DeleteVertexArrays(1, &vao)
	gl.DeleteBuffers(1, &vbo)
}
