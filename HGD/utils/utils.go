package utils

import (
	"path/filepath"
	"runtime"
)

func RGBToFloat32(r, g, b uint8) [3]float32 {
	return [3]float32{
		float32(r) / 255.0,
		float32(g) / 255.0,
		float32(b) / 255.0,
	}
}

func GetExecPath() string {
	_, filename, _, _ := runtime.Caller(1)

	return filepath.Dir(filename)
}

func CheckErrors(e error) {
	if e != nil {
		panic(e)
	}
}
