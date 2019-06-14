package main

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/sirkon/message"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
)

func main() {
	curDir, err := os.Getwd()
	if err != nil {
		message.Fatalf("failed to get current directory: %s", err)
	}

	repo, err := git.PlainOpen(curDir)
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			message.Fatal("%s is not in a repository", curDir)
		}
		message.Fatal(err)
	}

	remote := getRepoOrigin(repo)

	// нужно быть именно в каком-то из репозиториев от gitlab.stageoffice.ru
	var detected bool
	var remoteURL string
	for _, remoteURL = range remote.URLs {
		if strings.Contains(remoteURL, "gitlab.stageoffice.ru") {
			detected = true
			break
		}
	}
	if !detected {
		return
	}

	if len(os.Args) < 2 {
		message.Fatal("at least 1 argument expected")
	}
	rawCommitMsg, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		message.Fatalf("failed to get commit message: %s", err)
	}
	commitMsg := strings.TrimSpace(string(rawCommitMsg))

	ref, err := repo.Head()
	if err != nil {
		message.Fatalf("failed to get a HEAD: %s", err)
	}
	branch, ok := getBranchName(ref.String())
	if !ok {
		message.Fatal("failed to retrieve branch name from %s", ref.String())
	}

	if !strings.HasPrefix(commitMsg, branch) {
		message.Fatalf("commit message must look like `%s | <TEXT>`, got `%s` instead", branch, commitMsg)
	}

	var cmChecker CommitMsg
	if ok, err := cmChecker.Extract(commitMsg); !ok {
		if err != nil {
			message.Fatal("invalid branch name to commmit: %s", err)
		}
		message.Fatalf(`cannot commit such a branch name, it must be <PREFIX>-<NUM>, got %s instead`, branch)
	}

	var requiredPrefix string
	switch {
	case strings.Contains(remoteURL, "gitlab.stageoffice.ru/ucs/bazaar"):
		requiredPrefix = "CAT"
	case strings.Contains(remoteURL, "gitlab.stageoffice.ru/UCS-CALENDAR/"):
		requiredPrefix = "CC"
	case strings.Contains(remoteURL, "gitlab.stageoffice.ru/UCS-"):
		requiredPrefix = "UCS"
	default:
		message.Fatalf("%s: unsupported gitlab.stageoffice.ru/* kind of repository", remoteURL)
	}
	if cmChecker.Prefix != requiredPrefix {
		message.Fatal("commit message must be %s-<NUM> | <TEXT>, got `%s` instead", requiredPrefix, commitMsg)
	}
}

func getRepoOrigin(rep *git.Repository) *config.RemoteConfig {
	cfg, err := rep.Config()
	if err != nil {
		message.Fatal(err)
	}
	remote, ok := cfg.Remotes["origin"]
	if !ok {
		message.Fatal("no `origin` remote repo, cannot continue")
	}
	return remote
}

func getBranchName(ref string) (string, bool) {
	pos := strings.LastIndex(ref, "/")
	if pos < 0 {
		return "", false
	}
	return ref[pos+1:], true
}
