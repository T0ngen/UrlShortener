package storage

import "errors"



var (
	ErrURLNotFound = errors.New("url not Found")
	ErrURLExists = errors.New("url exists")
)