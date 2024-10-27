package common

import "errors"

var (
	ErrNotFound       = errors.New("resource not found")
	ErrUnauthorized   = errors.New("unauthorized access")
	ErrForbidden      = errors.New("forbidden access")
	ErrInvalidInput   = errors.New("invalid input provided")
	ErrInternalServer = errors.New("internal server error")
	ErrConflict       = errors.New("resource already exists")
)
