package main

import "github.com/sirkon/errors"

// PortalName имя портала.
type PortalName string

// UnmarshalText для реализации decoder.TextUnmarshaler.
func (s *PortalName) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		return errors.New("portal name must not be empty")
	}

	*s = PortalName(data)
	return nil
}
