package pokergame

import "errors"

var (
	ErrGame     = errors.New("game error")
	ErrInternal = errors.New("internal pokergame error")
)
