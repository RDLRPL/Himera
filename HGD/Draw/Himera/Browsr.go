package himera

import (
	drawer "github.com/RDLxxx/Himera/HGD/Draw/Drawer"
	"github.com/RDLxxx/Himera/HGD/Draw/TextLIB"
	"github.com/RDLxxx/Himera/HGD/core"
	"github.com/RDLxxx/Himera/HGD/utils"

	"github.com/go-gl/gl/v4.1-core/gl"
)

func DrawInputBox(rectProgram uint32, textProgram uint32) {
	inputBoxWidth := float32(core.Browse.CurrentWidth)

	drawer.DrawRect(rectProgram, 0, 0, inputBoxWidth, core.Browse.InputBoxHeight,
		utils.RGBToFloat32(200, 200, 200))
	drawer.DrawRect(rectProgram, 0, 0+core.Browse.InputBoxHeight-2.0, inputBoxWidth, 2.0, utils.RGBToFloat32(0, 0, 0))

	gl.UseProgram(textProgram)

	textY := 0 + core.Browse.InputBoxHeight/2 - TextLIB.GetLineHeight(1.0)/2 + TextLIB.GetFontAscent(1.0)
	TextLIB.DrawText(textProgram, core.Browse.InputText, 0, textY, 1.0,
		utils.RGBToFloat32(0, 0, 0))

	if core.Browse.InputBoxFocused {
		core.Browse.BlinkTimer += 16.0
		if int(core.Browse.BlinkTimer/500)%2 == 0 {
			cursorText := core.Browse.InputText[:core.Browse.CursorPosition]
			cursorX, _ := TextLIB.GetTextDimensions(cursorText, 1.0)
			drawer.DrawRect(rectProgram, 0+cursorX, 0+5.0, 2.0, core.Browse.InputBoxHeight-10.0,
				[3]float32{0.0, 0.0, 0.0})
			gl.UseProgram(textProgram)
		}
	}

	if core.Browse.InputText == "" {
		TextLIB.DrawText(textProgram, "Url", 0, textY, 1.0,
			utils.RGBToFloat32(150, 150, 150))
	}
}
