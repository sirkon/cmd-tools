package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/sirkon/errors"
	"github.com/sirkon/message"
)

func main() {
	var cli struct {
		Out    outFileName `short:"o" help:"Output file name." default:"out.pdf"`
		Except string      `short:"e" help:"Exclude files that have this substring in their path. Empty value means no filter." default:""`
		Src    []string    `arg:"" help:"Paths of files to insert into pdf." required:""`
	}

	parser := kong.Must(
		&cli,
		kong.Name("pdfy"),
		kong.Description("Creates PDF with text from given files."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Summary: true,
			Compact: true,
		}),
	)
	_, err := parser.Parse(os.Args[1:])
	if err != nil {
		parser.FatalIfErrorf(err)
	}

	var paths []string
	for _, path := range cli.Src {
		if path == "" {
			continue
		}

		if cli.Except != "" && strings.Contains(path, cli.Except) {
			continue
		}

		paths = append(paths, path)
	}

	if len(paths) == 0 {
		message.Fatalf("missing files")
	}

	if err := job(cli.Out, paths); err != nil {
		message.Fatal(err)
	}
}

func job(out outFileName, paths []string) (err error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "get user home dir")
	}

	desktop := filepath.Join(home, "Desktop")
	ir := filepath.Join(desktop, strings.TrimSuffix(string(out), ".pdf")+".md")

	if err := collectData(ir, paths); err != nil {
		return errors.Wrap(err, "collect data")
	}

	if err := os.Chdir(desktop); err != nil {
		return errors.Wrap(err, "chdir into "+desktop)
	}

	cmd := exec.Command("panpdf", ir)
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "convert intermediate .md file into pdf")
	}

	return nil
}

func collectData(ir string, paths []string) error {
	dest, err := os.Create(ir)
	if err != nil {
		return errors.Wrap(err, "create output intermediate file")
	}
	defer func() {
		if cErr := dest.Close(); cErr != nil {
			if err == nil {
				err = errors.Wrap(cErr, "close intermediate output file")
			}
			message.Error(errors.Wrap(cErr, "failed to close intermediate output file"))
		}
	}()

	buf := bufio.NewWriterSize(dest, 512*1024)
	defer func() {
		if fErr := buf.Flush(); fErr != nil {
			if err == nil {
				err = errors.Wrap(fErr, "flush output file")
				return
			}
			message.Error(errors.Wrap(fErr, "failed to flush into intermediate output file"))
		}
	}()
	for _, path := range paths {
		if err := pushFile(buf, path); err != nil {
			return errors.Wrap(err, "process "+path)
		}
	}

	return nil
}

func pushFile(file io.Writer, path string) error {
	var buf bytes.Buffer
	buf.WriteString("# ")
	buf.WriteString(filepath.Base(path))
	buf.WriteByte('\n')
	buf.WriteString("```")

	var cbType string
	switch filepath.Ext(path) {
	case ".go":
		cbType = "go"
	case ".rs":
		cbType = "rust"
	case ".sh":
		cbType = "shell"
	case ".py":
		cbType = "python"
	case ".cpp":
		cbType = "cpp"
	case ".c":
		cbType = "c"
	case ".h":
		cbType = "c"
	case ".pl":
		cbType = "perl"
	case ".rb":
		cbType = "ruby"
	case ".proto":
		cbType = "proto"
	}
	buf.WriteString(cbType)
	buf.WriteByte('\n')
	if _, err := io.Copy(file, &buf); err != nil {
		return errors.Wrap(err, "write header")
	}

	src, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "open file")
	}
	defer func() {
		if err := src.Close(); err != nil {
			message.Error(errors.Wrap(err, "failed to close "+path))
		}
	}()
	if _, err := io.Copy(file, src); err != nil {
		return errors.Wrap(err, "copy file data")
	}
	if _, err := io.WriteString(file, "```\n\n"); err != nil {
		return errors.Wrap(err, "write footer")
	}

	return nil
}

type outFileName string

func (o *outFileName) MarshalText() (text []byte, err error) {
	return []byte(*o), nil
}

func (o *outFileName) UnmarshalText(rawText []byte) error {
	text := string(rawText)
	if !strings.HasSuffix(text, ".pdf") {
		return errors.New("filename must end with .pdf")
	}

	*o = outFileName(text)
	return nil
}
