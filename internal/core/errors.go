package core

import "errors"

var (
	ErrOutOfBounds = errors.New("shift out of bounds")
	ErrEmptyStore  = errors.New("no images in store")
	ErrNotFound    = errors.New("image not found")
)
