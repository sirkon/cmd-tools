/*
The repo cleaner does the following

1) Switches to master
2) Fetches changes
3) Rebases upon the origin/master
4) Remove all branches except master
*/

package main

import (
	"bufio"
	"github.com/acarl005/stripansi"
	"github.com/sirkon/cmd-tools/internal/git"
	"github.com/sirkon/message"
	"regexp"
)

var (
	branchExtractor = regexp.MustCompile(`^\s*\*?\s(.*)$`)
)

func main() {
	if _, err := git.Do("checkout", "master"); err != nil {
		message.Fatal(err)
	}

	if _, err := git.Do("fetch"); err != nil {
		message.Fatal(err)
	}
	if _, err := git.Do("rebase", "origin/master"); err != nil {
		message.Fatal(err)
	}

	branches, err := git.Do("branch")
	if err != nil {
		message.Fatal(err)
	}
	scanner := bufio.NewScanner(branches)
	for scanner.Scan() {
		text := scanner.Text()
		data := branchExtractor.FindStringSubmatch(text)
		if len(data) == 0 {
			continue
		}
		branch := stripansi.Strip(data[1])
		if len(branch) == 0 {
			continue
		}
		if branch == "master" {
			message.Debug("ignoring master branch")
			continue
		}
		if _, err := git.Do("branch", "-D", branch); err != nil {
			message.Error(err)
		} else {
			message.Debugf("%s deleted", branch)
		}
	}
}
