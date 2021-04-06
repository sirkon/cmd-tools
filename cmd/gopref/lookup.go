package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirkon/errors"
)

func goSrcRoot() (string, error) {
	var gopath string

	gopath = os.Getenv("GOPATH")
	if gopath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.Wrap(err, "compute user home dir")
		}

		gopath = filepath.Join(home, "go")
	}

	return filepath.Join(gopath, "src"), nil
}

// CrossPredictor предиктор для автодополнения имён директорий с учётом пересечения названия — добавляет уникальный
// суффикс позволяющий уникально определить предиктор
type CrossPredictor interface {
	Identifier() string
	Predict(prefix string, res map[string][]CrossPredictor)
	Point(value string) (string, error)
}

// Predictor поиск имён только для директорий находящихся в заданном корне, без рекурсивности
type Predictor struct {
	root         string
	identifier   string
	prolongSlash bool
}

// Identifier для реализации CrossPredictor
func (j *Predictor) Identifier() string {
	return j.identifier
}

// Predict для реализации CrossPredictor
func (j *Predictor) Predict(prefix string, res map[string][]CrossPredictor) {
	parent, rest := filepath.Split(prefix)
	prjs, err := os.ReadDir(filepath.Join(j.root, parent))
	if err != nil {
		loggerr(errors.Wrap(err, "list directories in "+j.root))
	}

	for _, prj := range prjs {
		if !prj.IsDir() {
			continue
		}

		if strings.HasPrefix(prj.Name(), rest) {
			key := parent + prj.Name()
			if j.prolongSlash {
				key += "/"
			}
			res[key] = append(res[key], j)
		}
	}
}

// Point для реализации CrossPredictor
func (j *Predictor) Point(value string) (string, error) {
	path := filepath.Join(j.root, value)
	stat, err := os.Stat(path)
	if err != nil {
		return "", errors.Wrap(err, "check file system entry for "+value)
	}

	if !stat.IsDir() {
		return "", errors.Newf("entry %s refers not a directory", value)
	}

	return path, nil
}

// NewPredictor конструктор Predictor
func NewPredictor(root string, identifier string, prolongSlash bool) CrossPredictor {
	return &Predictor{
		root:         root,
		identifier:   identifier,
		prolongSlash: prolongSlash,
	}
}
