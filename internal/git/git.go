package git

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// Do does git given git command with params
func Do(command string, params ...string) (io.Reader, error) {
	commandLine := strings.Join(append([]string{"git", command}, params...), " ")
	cmd := exec.Command("git", append([]string{command}, params...)...)
	output := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = output
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("%s:\n%s", commandLine, stderr.String())
	}
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("%s: \n%s", commandLine, stderr.String())
	}
	return output, nil
}
