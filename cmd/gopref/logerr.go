package main

import (
	"os"

	"github.com/sirkon/message"
)

func loggerr(err error) {
	if os.Getenv("COMPLETELOG") == "" {
		return
	}

	message.Error(err)
}
