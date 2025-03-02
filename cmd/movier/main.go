package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/sirkon/errors"
	"github.com/sirkon/message"
)

func main() {
	if len(os.Args) != 2 {
		message.Fatalf("Usage: %s <dir>", os.Args[0])
	}

	wdir := os.Args[1]
	if err := job(wdir); err != nil {
		message.Fatal(
			errors.Wrap(err, "move files in the working directory").
				Str("working-directory-path", wdir),
		)
	}
}

func job(wdir string) error {
	stat, err := os.Stat(wdir)
	if err != nil {
		return errors.Wrap(err, "get working directory stat")
	}
	if !stat.IsDir() {
		return errors.New("is not a directory")
	}

	infos, err := os.ReadDir(wdir)
	if err != nil {
		return errors.Wrap(err, "get files from the working directory")
	}
	var files []string
	for _, info := range infos {
		if info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			continue
		}
		files = append(files, info.Name())
	}

	fullName := func(name string) string {
		return filepath.Join(wdir, name)
	}

	orderMatcher := regexp.MustCompile(`^\d+\.\s*(.*)$`)

	sort.Slice(files, func(i, j int) bool {
		name1 := files[i]
		name2 := files[j]
		ordered1 := orderMatcher.MatchString(name1)
		ordered2 := orderMatcher.MatchString(name2)
		if ordered1 && ordered2 {
			return name1 < name2
		}
		if ordered1 {
			return false
		}
		if ordered2 {
			return false
		}
		return name1 < name2
	})

	zeroes := len(fmt.Sprintf("%d", len(files)))
	if zeroes < 3 {
		zeroes = 3
	}
	mask := fmt.Sprintf("%%0%dd. %%s", zeroes)

	for i, file := range files {
		var moveTo string
		if orderMatcher.MatchString(file) {
			data := orderMatcher.FindStringSubmatch(file)
			moveTo = data[1]
		} else {
			moveTo = file
		}
		moveTo = fmt.Sprintf(mask, i+1, moveTo)
		if file == moveTo {
			message.Infof("will not move %s as it keeps its old name", file)
			continue
		}

		message.Infof("moving %s → %s", file, moveTo)
		if err := os.Rename(fullName(file), fullName(moveTo)); err != nil {
			return errors.Wrap(err, "move file").
				Str("from-file-path", file).
				Str("to-file-path", moveTo)
		}
	}

	return nil
}
