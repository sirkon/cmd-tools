package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/posener/complete"
	"github.com/sirkon/errors"
	"github.com/sirkon/message"
	"github.com/willabides/kongplete"
)

var cliParams struct {
	EntryID            string `arg:"" name:"entry-id" help:"Directory identifier" predictor:"pred"`
	InstallCompletions bool   `short:"i" help:"Setup completions and exit."`
}

func main() {
	parser := kong.Must(
		&cliParams,
		kong.Name("gj"),
		kong.Description("Show directory by directory id."),
		kong.UsageOnError(),
	)
	kongplete.Complete(
		parser,
		kongplete.WithPredictor(
			"pred",
			complete.PredictFunc(func(args complete.Args) []string {
				res, err := predictionJob(args.Last)
				if err != nil {
					loggerr(errors.Wrap(err, "predict for "+args.Last))
					return nil
				}

				return res
			}),
		),
	)

	ctx, err := parser.Parse(os.Args[1:])
	parser.FatalIfErrorf(err)

	if cliParams.InstallCompletions {
		if err := kongplete.InstallShellCompletions(ctx); err != nil {
			message.Fatal(errors.Wrap(err, "install completions"))
		}

		return
	}

	id, predID := decomposeCompl(cliParams.EntryID)

	predictors, err := getPredictors()
	if err != nil {
		message.Fatal(errors.Wrap(err, "get predictors"))
	}

	res := map[string][]CrossPredictor{}
	for _, p := range predictors {
		if !strings.HasPrefix(p.Identifier(), predID) {
			continue
		}
		p.Predict(id, res)
	}

	preds := res[id]
	switch len(preds) {
	case 0:
		gosrc, err := goSrcRoot()
		if err == nil {
			respath := filepath.Join(gosrc, id)
			if _, err := os.Stat(respath); err == nil {
				fmt.Println(respath)
				return
			}
		}
		message.Fatalf("%s: no directory for this identifier", id)
	case 1:
		point, err := preds[0].Point(id)
		if err != nil {
			message.Fatal(errors.Wrap(err, "get path for "+id))
		}
		fmt.Println(point)
	default:
		for _, p := range preds {
			if p.Identifier() == predID {
				point, err := p.Point(id)
				if err != nil {
					message.Fatal(errors.Wrap(err, "get path for "+id))
				}
				fmt.Println(point)
				return
			}
		}

		message.Fatalf("unknown entry point " + cliParams.EntryID)
	}
}

func decomposeCompl(prefix string) (string, string) {
	parts := strings.SplitN(prefix, ":", 2)

	if len(parts) == 1 {
		return parts[0], ""
	} else {
		return parts[0], parts[1]
	}
}

func predictionJob(prefix string) ([]string, error) {
	predictors, err := getPredictors()
	if err != nil {
		return nil, errors.Wrap(err, "get predictors")
	}

	res := map[string][]CrossPredictor{}
	prefix, predID := decomposeCompl(prefix)

	for _, p := range predictors {
		if !strings.HasPrefix(p.Identifier(), predID) {
			continue
		}

		if strings.HasSuffix(prefix, "/") {
			p.Predict(strings.TrimRight(prefix, "/"), res)
		}
		p.Predict(prefix, res)
	}

	keys := make([]string, 0, len(res))
	for k := range res {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result []string
	for _, k := range keys {
		v := res[k]
		if len(v) == 1 && predID == "" {
			result = append(result, k)
		} else {
			for _, vv := range v {
				result = append(result, k+":"+vv.Identifier())
			}
		}
	}

	return result, nil
}

func getPredictors() ([]CrossPredictor, error) {
	src, err := goSrcRoot()
	if err != nil {
		return nil, errors.Wrap(err, "compute go sources root dir")
	}

	var predictors []CrossPredictor

	ucsRoot := filepath.Join(src, "gitlab.stageoffice.ru")
	prjs, err := os.ReadDir(ucsRoot)
	if err != nil {
		return nil, errors.Wrap(err, "get groups in UCS dir")
	}

	for _, prj := range prjs {
		p := NewPredictor(filepath.Join(ucsRoot, prj.Name()), prj.Name(), false)
		predictors = append(predictors, p)
	}

	p := NewPredictor(filepath.Join(src, "github.com/sirkon"), "home", false)
	predictors = append(predictors, p)

	p = NewPredictor(src, "all", true)
	predictors = append(predictors, p)
	return predictors, nil
}
