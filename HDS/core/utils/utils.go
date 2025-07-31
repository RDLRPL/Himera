package utils

func RGBToFloat32(r, g, b uint8) [3]float32 {
	return [3]float32{
		float32(r) / 255.0,
		float32(g) / 255.0,
		float32(b) / 255.0,
	}
}
