package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirkon/message"

	"github.com/sirkon/cmd-tools/internal/git"
)

func main() {
	gopath := os.Getenv("GOPATH")
	if len(gopath) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			message.Fatalf("failed to get user home directory path: %s", err)
		}
		gopath = filepath.Join(home, "go")
	}

	ja := jobAction{
		gosrc:  filepath.Join(gopath, "src") + string(filepath.Separator),
		branch: "UCS-6294",
	}

	message.Info("start looking for branches desired")
	err := filepath.Walk(ja.lookupPath(), ja.action)
	if err != nil {
		message.Fatal("premature exit: %s", err)
	}
}

type jobAction struct {
	gosrc  string
	branch string
}

func (a *jobAction) lookupPath() string {
	return a.gosrc
}

func (a *jobAction) action(path string, info os.FileInfo, err error) error {
	if !info.IsDir() {
		return nil
	} else if !strings.HasPrefix(path, a.lookupPath()) {
		return filepath.SkipDir
	}

	prjName := path[len(a.lookupPath()):]
	pathItems := filepath.SplitList(prjName)
	if len(pathItems) > 3 {
		// находимся внутри проекта, выходим
		return filepath.SkipDir
	}
	if len(pathItems) < 3 {
		// ещё не дошли до каталогов с проектами, переходим к следующим элементам
		return nil
	}

	if err := a.processProject(path, prjName); err != nil {
		message.Errorf("failed to process `%s`: %s", prjName, err)
	}
	return filepath.SkipDir
}

func (a *jobAction) processProject(prjPath, prjName string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current dir: %s", err)
	}

	defer func() {
		err = os.Chdir(cwd)
		if err != nil {
			message.Fatalf("failed to return into `%s`: %s", cwd, err)
		}
	}()

	if err := os.Chdir(prjPath); err != nil {
		return fmt.Errorf("failed to switch into %s: %s", prjName, err)
	}

	res, err := git.Do("branch", "-a")
	if err != nil {
		return fmt.Errorf("git error: %s", err)
	}

	scanner := bufio.NewScanner(res)
	var be Branch
	var branchLocated bool
	var curBranch string

	for scanner.Scan() {
		if ok, _ := be.Extract(scanner.Bytes()); !ok {
			message.Fatalf("internal error, failed to parse input `%s`", scanner.Text())
		}
		if be.Branch == a.branch {
			branchLocated = true
		}
		if be.Current.Valid {
			curBranch = be.Branch
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan git branch command output: %s", err)
	}

	if !branchLocated {
		return nil
	}

	return nil
}
