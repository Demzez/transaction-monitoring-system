package custom_error

import "errors"

var (
	ErrFunc = errors.New("some function error")
)
