package main

import (
	"fmt"
	"github.com/shirou/gopsutil/process"
	"github.com/sirkon/message"
	"os"
)

type extractor interface {
	Extract(input string) (ok bool, err error)
}

// checkers список фильтров на приложеньки, которые неплохо было бы убить
var checkers = map[string]extractor{
	"War Thunder":          &WarThunderAces{},
	"War Thunder launcher": &WarThunderLauncher{},
	"The Witcher 3":        &Witcher3{},
}

func main() {
	ps, err := process.Processes()
	if err != nil {
		message.Fatalf("failed to get processes list: %s", err)
	}

	for _, p := range ps {
		cmd, err := p.Cmdline()
		if err != nil {
			message.Errorf("failed to get cmd line for process %d: %s", p.Pid, err)
			continue
		}

		for app, checker := range checkers {
			if ok, _ := checker.Extract(cmd); ok {
				_, _ = fmt.Fprintf(os.Stderr, "killing %s (pid=%d, cmdline=\"%s\"): ", app, p.Pid, cmd)
				if err := p.Kill(); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "\033[31mfailed to kill: %s\n", err)
				} else {
					_, _ = os.Stderr.WriteString("done\n")
				}
			}
		}
	}
}
