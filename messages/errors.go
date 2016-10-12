package messages

import (
	"errors"
	"fmt"
)

var ErrInvalidMessage = errors.New("Invalid message")
var ErrInvalidCategory = errors.New("Invalid category")

type ErrInvalidKeyword struct {
	Keyword string
}

func (e *ErrInvalidKeyword) Error() string {
	return fmt.Sprintf("Invalid keyword: %s", e.Keyword)
}

func InvalidKeyword(keyword string) error {
	e := ErrInvalidKeyword{}
	e.Keyword = keyword
	return &e
}