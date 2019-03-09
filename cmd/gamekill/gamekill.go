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

	var gamesCount int
	var gamesKilled int
	for _, p := range ps {
		cmd, err := p.Cmdline()
		if err != nil {
			message.Errorf("failed to get cmd line for process %d: %s", p.Pid, err)
			continue
		}

		for app, checker := range checkers {
			if ok, _ := checker.Extract(cmd); ok {
				gamesCount++
				_, _ = fmt.Fprintf(os.Stderr, "killing %s (pid=%d, cmdline=\"%s\"): ", app, p.Pid, cmd)
				if err := p.Kill(); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "\033[31mfailed to kill: %s\n", err)
				} else {
					gamesKilled++
					_, _ = os.Stderr.WriteString("done\n")
				}
			}
		}
	}

	switch gamesCount {
	case 0:
		message.Info("summary: no games detected")
	case 1:
		if gamesKilled == gamesCount {
			message.Info("summary: only 1 game detected, killed")
		} else {
			message.Info("summary: only 1 game detected, failed to kill it")
		}
	default:
		if gamesKilled == gamesCount {
			message.Infof("summary: %d games detected, all killed", gamesCount)
		} else {
			message.Infof("summary: %d games detected, killed %d of them", gamesCount, gamesKilled)
		}
	}
}
