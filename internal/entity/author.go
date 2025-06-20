package entity

import (
	"github.com/pkg/errors"
)

type Author struct {
	ID   string
	Name string
}

var (
	ErrAuthorNotFound = errors.New("author not found")
)
