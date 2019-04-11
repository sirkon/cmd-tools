package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/sirkon/message"

	"github.com/sirkon/cmd-tools/internal/git"
)

var args struct {
	Branch  string `arg:"positional,required"`
	Verbose bool   `arg:"--verbose" help:"diagnose output"`
}

func main() {
	arg.MustParse(&args)

	gopath := os.Getenv("GOPATH")
	if len(gopath) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			message.Fatal(err)
		}
		gopath = filepath.Join(home, "go")
	}

	gosrc := filepath.Join(gopath, "src")

	action := action{
		branch:  args.Branch,
		gopath:  gopath,
		gosrc:   gosrc,
		prefix:  gosrc + string(filepath.Separator),
		verbose: args.Verbose,
	}
	if err := filepath.Walk(gosrc, action.action); err != nil {
		message.Fatal(err)
	}
}

type action struct {
	branch  string
	gopath  string
	gosrc   string
	prefix  string
	verbose bool
}

func (a *action) action(path string, info os.FileInfo, err error) error {
	if !info.IsDir() {
		return nil
	}
	if path == a.gosrc {
		return nil
	}
	if !strings.HasPrefix(path, a.prefix) {
		message.Warningf("unexpected path `%s` for given go path `%s`", path, a.gopath)
		return filepath.SkipDir
	}

	project := path[len(a.prefix):]
	items := strings.Split(project, string(filepath.Separator))

	if len(items) < 3 {
		return nil
	}
	if len(items) > 3 {
		return filepath.SkipDir
	}

	if err := os.Chdir(path); err != nil {
		message.Fatalf("failed to chdir into %s: %s", path, err)
	}

	res, err := git.Do("branch")
	if err != nil {
		if a.verbose {
			message.Warningf("warning: project `%s` is not under git control", project)
		}
		return filepath.SkipDir
	}

	scanner := bufio.NewScanner(res)
	var be Branch
	for scanner.Scan() {
		if ok, _ := be.Extract(scanner.Bytes()); !ok {
			message.Fatalf("failed to process branch line `%s`", scanner.Text())
		}
		if a.branch == be.Name {
			message.Info(project)
		}
	}
	return nil
}
