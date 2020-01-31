/*
The repo updater does the following

1) Fetches changes
2) Rebases upon origin/<repo name>
*/

package main

import (
	"bufio"
	"github.com/acarl005/stripansi"
	"github.com/alexflint/go-arg"
	"github.com/sirkon/cmd-tools/internal/git"
	"github.com/sirkon/message"
	"regexp"
)

var (
	branchExtractor = regexp.MustCompile(`^\s*(\*)?\s(.*)$`)
)

func main() {
	var args struct {
		Reset bool `arg:"-r,--reset" help:"hard reset to origin/<branch> instead of rebase"`
	}
	arg.MustParse(&args)

	if _, err := git.Do("fetch"); err != nil {
		message.Fatal(err)
	}

	var cmd string

	branches, err := git.Do("branch")
	if err != nil {
		message.Fatal("get branches", err)
	}
	scanner := bufio.NewScanner(branches)
	for scanner.Scan() {
		text := scanner.Text()
		data := branchExtractor.FindStringSubmatch(text)
		if len(data) == 0 {
			continue
		}
		if data[1] != "*" {

		}
		branch := stripansi.Strip(data[2])
		if len(branch) == 0 {
			continue
		}
		if branch == "master" {
			message.Debug("ignoring master branch")
			continue
		}

		var params []string
		if args.Reset {
			cmd = "reset"
			params = []string{"--hard"}
			message.Debug("resetting onto origin/" + branch)
		} else {
			cmd = "rebase"
			message.Debug("rebasing onto origin/" + branch)
		}

		params = append(params, "origin/"+branch)
		if _, err := git.Do(cmd, params...); err != nil {
			message.Error(err)
		} else {
			if args.Reset {
				message.Debugf("%s reseted onto origin", branch)
			} else {
				message.Debugf("%s rebased onto origin", branch)
			}
		}
		break
	}
}
