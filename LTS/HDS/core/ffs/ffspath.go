package ffs

import (
	"errors"
)

type Ffs struct {
	Dir string
}

func NewFFST(dir string) *Ffs {
	return &Ffs{
		Dir: dir,
	}
}

var ErrTemplateNotFound = errors.New("template not found")
