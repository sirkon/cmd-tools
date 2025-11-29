package main

import (
	"bytes"
	"os/exec"
	"strings"

	"github.com/sirkon/errors"
	"github.com/sirkon/message"
)

func main() {
	// Get current branch
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		message.Fatal(errors.Wrap(err, "get current branch"))
	}
	branch := strings.TrimSpace(string(out))

	// Check if remote branch exists
	if branch != "HEAD" {
		check, _ := exec.Command("git", "ls-remote", "--heads", "origin", branch).Output()
		if len(check) == 0 {
			message.Fatal(errors.Newf("remote branch origin/%s does not exist", branch))
		}
	}

	// Fetch
	if out, err := exec.Command("git", "fetch").CombinedOutput(); err != nil {
		message.Fatal(errors.Wrap(err, "fetch remote changes: "+string(out)))
	}

	// Reset
	if out, err := exec.Command("git", "reset", "--hard", "origin/"+branch).CombinedOutput(); err != nil {
		message.Fatal(errors.Wrap(err, "reset failed: "+string(out)))
	}

	message.Info("done")
}
