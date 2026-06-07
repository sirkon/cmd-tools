package main

import (
	"github.com/sirkon/errors"
)

type Description string

func (d *Description) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return errors.New("description must not be empty")
	}

	*d = Description(text)
	return nil
}
