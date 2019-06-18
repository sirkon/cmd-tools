package main

import (
	"fmt"
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
	if branch == "master" {
		message.Fatal("commiting to master is not allowed at gitlab.stageoffice.ru")
	}
	var bnv BranchNameValidator
	if ok, _ := bnv.Extract(branch); !ok {
		message.Fatalf("you can only commit in branches like <PREFIX>-<NUM> (<PREFIX> = UCS, CAT, CC, etc) , got %s instead", branch)
	}

	if !strings.HasPrefix(commitMsg, branch) {
		message.Fatalf("commit message must look like `%s | <TEXT>`, got `%s` instead", branch, commitMsg)
	}

	var acceptablePrefixes []string
	const (
		catalog      = "CAT"
		calendar     = "CC"
		wholeproject = "UCS"
	)
	switch {
	case strings.Contains(remoteURL, "gitlab.stageoffice.ru/ucs/bazaar"):
		acceptablePrefixes = append(acceptablePrefixes, catalog)
	case strings.Contains(remoteURL, "gitlab.stageoffice.ru:ucs/bazaar"):
		acceptablePrefixes = append(acceptablePrefixes, catalog)
	case strings.Contains(remoteURL, "gitlab.stageoffice.ru/UCS-CALENDAR/"):
		acceptablePrefixes = append(acceptablePrefixes, calendar)
	case strings.Contains(remoteURL, "gitlab.stageoffice.ru:UCS-CALENDAR/"):
		acceptablePrefixes = append(acceptablePrefixes, calendar)
	case strings.Contains(remoteURL, "gitlab.stageoffice.ru/UCS-CATALOG/"):
		acceptablePrefixes = append(acceptablePrefixes, catalog)
	case strings.Contains(remoteURL, "gitlab.stageoffice.ru:UCS-CATALOG/"):
		acceptablePrefixes = append(acceptablePrefixes, catalog)
	case strings.Contains(remoteURL, "gitlab.stageoffice.ru/UCS-COMMON/schema"):
		acceptablePrefixes = append(acceptablePrefixes, catalog, calendar, wholeproject)
	case strings.Contains(remoteURL, "gitlab.stageoffice.ru:UCS-COMMON/schema"):
		acceptablePrefixes = append(acceptablePrefixes, catalog, calendar, wholeproject)
	case strings.Contains(remoteURL, "gitlab.stageoffice.ru/UCS-"):
		acceptablePrefixes = append(acceptablePrefixes, wholeproject)
	case strings.Contains(remoteURL, "gitlab.stageoffice.ru:UCS-"):
		acceptablePrefixes = append(acceptablePrefixes, wholeproject)
	default:
		message.Warningf("%s: unsupported gitlab.stageoffice.ru/* kind of repository", remoteURL)
		return
	}

	var cmChecker CommitMsg
	if ok, err := cmChecker.Extract(commitMsg); !ok {
		if err != nil {
			message.Fatal("invalid branch name to commmit: %s", err)
		}
		message.Fatalf(`cannot commit such a branch name, it must be %s-<NUM>, got %s instead`, acceptablePrefixes, branch)
	}

	if !isInStringArray(cmChecker.Prefix, acceptablePrefixes) {
		switch len(acceptablePrefixes) {
		case 0:
			message.Fatal("internal error – no prefix found for remote path %s", remoteURL)
		case 1:
			message.Fatal("commit message must be %s-<NUM> | <TEXT>, got `%s` instead", acceptablePrefixes[0], commitMsg)
		case 2:
			message.Fatal("commit message must be either %s-<NUM> or %s-<NUM> | <TEXT>, got `%s` instead",
				acceptablePrefixes[0],
				acceptablePrefixes[1],
				commitMsg,
			)
		default:
			var values []string
			for _, prefix := range acceptablePrefixes {
				values = append(values, fmt.Sprintf("%s-<NUM> | <TEXT>", prefix))
			}
			finalValue := fmt.Sprintf("%s or %s", strings.Join(values[:len(values)-1], ", "), values[len(values)-1])
			message.Fatal("commit message must be one of %s, got `%s` instead", finalValue, commitMsg)
		}

	}
}

func isInStringArray(value string, array []string) bool {
	for _, item := range array {
		if item == value {
			return true
		}
	}
	return false
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
