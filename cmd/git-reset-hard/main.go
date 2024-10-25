package main

import (
	"bytes"
	"github.com/sirkon/errors"
	"github.com/sirkon/message"
	"os/exec"
	"strings"
)

func main() {
	cmd := exec.Command("git", "branch", "--show-current")
	var tmpBuffer bytes.Buffer
	cmd.Stdout = &tmpBuffer
	if err := cmd.Run(); err != nil {
		message.Fatal(errors.Wrap(err, "get current branch"))
	}
	branch := strings.TrimSpace(tmpBuffer.String())

	if err := exec.Command("git", "fetch").Run(); err != nil {
		message.Fatal(errors.Wrap(err, "fetch remote changes"))
	}

	if err := exec.Command("git", "reset", "--hard", "origin/"+branch).Run(); err != nil {
		message.Fatal(errors.Wrap(err, "move HEAD to the state of the remote"))
	}

	message.Info("done")
}
