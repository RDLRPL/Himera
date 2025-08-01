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

var FontMetrics font.Metrics

const (
	FontSize = 14.0
	Dpi      = 72.0
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

	// Включаем блендинг для прозрачности
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	// Базовая линия для правильного выравнивания текста
	baseline := y + float32(FontMetrics.Ascent>>6)*scale

	for _, ch := range text {
		char := Characters[ch]
		if char == nil {
			// Используем символ замещения для неизвестных символов
			char = Characters[rune('?')]
			if char == nil {
				continue
			}
		}

		xpos := x + float32(char.Bearing[0])*scale
		ypos := baseline - float32(char.Bearing[1])*scale // Правильный расчет Y от базовой линии

		w := float32(char.Size[0]) * scale
		h := float32(char.Size[1]) * scale

		// Правильная ориентация вершин и текстурных координат
		vertices := []float32{
			// Position      // Texture
			xpos, ypos, 0.0, 1.0, // Bottom-left
			xpos + w, ypos, 1.0, 1.0, // Bottom-right
			xpos, ypos + h, 0.0, 0.0, // Top-left

			xpos + w, ypos, 1.0, 1.0, // Bottom-right
			xpos, ypos + h, 0.0, 0.0, // Top-left
			xpos + w, ypos + h, 1.0, 0.0, // Top-right
		}
		gl.BindTexture(gl.TEXTURE_2D, char.TextureID)
		gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, len(vertices)*4, gl.Ptr(vertices))

		gl.DrawArrays(gl.TRIANGLES, 0, 6)

		x += float32(char.Advance>>6) * scale
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

	// Добавляем расширенный набор символов включая кириллицу
	ranges := [][2]rune{
		{32, 126},    // ASCII
		{160, 255},   // Расширенная латиница
		{1040, 1103}, // Кириллица (А-Я, а-я)
		{1025, 1025}, // Ё
		{1105, 1105}, // ё
	}

	for _, r := range ranges {
		for ch := r[0]; ch <= r[1]; ch++ {
			if err := createCharacterTexture(face, ch); err != nil {
				fmt.Printf("Warning: failed to create texture for character %c: %v\n", ch, err)
			}
		}
	}

	// Создаем символ замещения для неизвестных символов
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

	// Вычисляем размеры символа с учетом padding
	w := int((bounds.Max.X - bounds.Min.X) >> 6)
	h := int((bounds.Max.Y - bounds.Min.Y) >> 6)

	// Минимальные размеры
	if w <= 0 {
		w = int(FontSize / 2)
	}
	if h <= 0 {
		h = int(FontSize)
	}

	// Добавляем padding для лучшего качества
	padding := 2
	imgW := w + padding*2
	imgH := h + padding*2

	// Создаем изображение с альфа-каналом
	img := image.NewRGBA(image.Rect(0, 0, imgW, imgH))

	// Заполняем прозрачным цветом
	draw.Draw(img, img.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 0}}, image.Point{}, draw.Src)

	// Создаем drawer для рендеринга символа
	drawer := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{color.RGBA{255, 255, 255, 255}}, // Белый цвет
		Face: face,
		Dot: fixed.Point26_6{
			X: -bounds.Min.X + fixed.I(padding),
			Y: -bounds.Min.Y + fixed.I(padding),
		},
	}

	drawer.DrawString(string(ch))

	stride := img.Stride
	tmp := make([]byte, stride)
	for y := 0; y < imgH/2; y++ {
		topOff := y * stride
		botOff := (imgH - 1 - y) * stride

		// сохраняем строку y
		copy(tmp, img.Pix[topOff:topOff+stride])
		// сверху копируем строку снизу
		copy(img.Pix[topOff:topOff+stride], img.Pix[botOff:botOff+stride])
		// снизу кладём сохранённую
		copy(img.Pix[botOff:botOff+stride], tmp)
	}

	// Создаем OpenGL текстуру
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)

	// Загружаем данные изображения в текстуру
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(imgW),
		int32(imgH),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(img.Pix),
	)

	// Настраиваем параметры текстуры
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	Characters[ch] = &Character{
		TextureID: texture,
		Size:      [2]int32{int32(imgW), int32(imgH)},
		Bearing:   [2]int32{int32(bounds.Min.X>>6) - int32(padding), int32(bounds.Max.Y>>6) + int32(padding)},
		Advance:   int32(advance),
	}

	return nil
}

// Дополнительная функция для получения размеров текста
func GetTextDimensions(text string, scale float32) (width, height float32) {
	if FontMetrics.Height == 0 {
		return 0, 0
	}

	width = 0
	height = float32(FontMetrics.Height>>6) * scale

	for _, ch := range text {
		char := Characters[ch]
		if char != nil {
			width += float32(char.Advance>>6) * scale
		}
	}

	return width, height
}
