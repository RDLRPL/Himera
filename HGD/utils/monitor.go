package utils

import (
	"github.com/go-gl/glfw/v3.3/glfw"
)

type Monitor struct {
	Width  int
	Height int
}

func GetPrimaryMonitor() (*Monitor, error) {
	if err := glfw.Init(); err != nil {
		return nil, err
	}
	defer glfw.Terminate()

	monitor := glfw.GetPrimaryMonitor()
	if monitor == nil {
		return nil, &glfw.Error{}
	}

	videoMode := monitor.GetVideoMode()
	if videoMode == nil {
		return nil, &glfw.Error{}
	}

	return &Monitor{
		Width:  videoMode.Width,
		Height: videoMode.Height,
	}, nil
}
