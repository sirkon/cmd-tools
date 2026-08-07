package main

import (
	"os/exec"
	"strings"

	"github.com/sirkon/errors"
)

func getGitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")

	// CombinedOutput собирает все, что выдала команда (и результат, и ошибку)
	output, err := cmd.CombinedOutput()
	cleanOutput := strings.TrimSpace(string(output))

	if err != nil {
		if cleanOutput != "" {
			// Если команда упала, но что-то написала в консоль — возвращаем этот текст
			// Обычно там: "fatal: not a git repository..."
			return "", errors.Newf("git error: %s", cleanOutput)
		}

		// Если упало совсем глухо (например, самого git нет в системе)
		return "", errors.Wrap(err, "execute git command")
	}

	return cleanOutput, nil
}
