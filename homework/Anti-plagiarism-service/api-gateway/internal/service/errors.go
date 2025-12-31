package service

import (
	"errors"
)

var (
	ErrWorkNotFound      = errors.New("work not found")
	ErrWorkAlreadyExists = errors.New("work already exists")
)
