package drawer

import (
	"github.com/RDLxxx/Himera/HGD/core"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type GLResources struct {
	rectVAO, rectVBO uint32
	initialized      bool
}

var glResources = &GLResources{}

func InitGLResources() {
	if glResources.initialized {
		return
	}

	gl.GenVertexArrays(1, &glResources.rectVAO)
	gl.GenBuffers(1, &glResources.rectVBO)

	glResources.initialized = true
}

func CleanupGLResources() {
	if glResources.initialized {
		gl.DeleteVertexArrays(1, &glResources.rectVAO)
		gl.DeleteBuffers(1, &glResources.rectVBO)
		glResources.initialized = false
	}
}

func DrawRect(program uint32, x, y, width, height float32, color [3]float32) {
	if !glResources.initialized {
		InitGLResources()
	}

	vertices := []float32{
		x, y,
		x, y + height,
		x + width, y + height,
		x, y,
		x + width, y + height,
		x + width, y,
	}

	gl.BindVertexArray(glResources.rectVAO)
	gl.BindBuffer(gl.ARRAY_BUFFER, glResources.rectVBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STREAM_DRAW)

	gl.VertexAttribPointer(0, 2, gl.FLOAT, false, 2*4, nil)
	gl.EnableVertexAttribArray(0)

	gl.UseProgram(program)

	projectionLoc := gl.GetUniformLocation(program, gl.Str("projection\x00"))
	if projectionLoc >= 0 {
		projection := [16]float32{
			2.0 / float32(core.Browse.CurrentWidth), 0, 0, 0,
			0, -2.0 / float32(core.Browse.CurrentHeight), 0, 0,
			0, 0, -1, 0,
			-1, 1, 0, 1,
		}
		gl.UniformMatrix4fv(projectionLoc, 1, false, &projection[0])
	}

	colorLoc := gl.GetUniformLocation(program, gl.Str("fillColor\x00"))
	if colorLoc >= 0 {
		gl.Uniform3f(colorLoc, color[0], color[1], color[2])
	}

	if width > 0 && height > 0 {
		gl.DrawArrays(gl.TRIANGLES, 0, 6)
	}

	gl.BindVertexArray(0)
}
