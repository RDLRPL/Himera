package TextLIB

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

func GetLineHeight(scale float32) float32 {
	return float32(FontMetrics.Height>>6) * scale
}

func GetFontAscent(scale float32) float32 {
	return float32(FontMetrics.Ascent>>6) * scale
}

func GetFontDescent(scale float32) float32 {
	return float32(FontMetrics.Descent>>6) * scale
}

func GetBaselineY(y, height, scale float32) float32 {
	ascent := GetFontAscent(scale)
	return y + height/2 - ascent/2
}
