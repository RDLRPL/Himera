package goutils

import (
	"path/filepath"
	"runtime"
)

func GetExecPath() string {
	_, filename, _, _ := runtime.Caller(1)

	return filepath.Dir(filename)
}
