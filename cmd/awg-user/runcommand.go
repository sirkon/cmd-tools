package main

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/sirkon/errors"
)

func runCommand(input string, cmd string, params ...string) (string, error) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	command := exec.Command(cmd, params...)
	command.Stdout = &stdout
	command.Stderr = &stderr

	if input != "" {
		command.Stdin = strings.NewReader(input)
	}

	if err := command.Run(); err != nil {
		msg := strings.TrimSpace(err.Error())
		return "", errors.Wrap(err, msg)
	}

	return strings.TrimSpace(stdout.String()), nil
}
