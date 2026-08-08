package main

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/sirkon/errors"
	"github.com/sirkon/message"
	"github.com/sirkon/varsize"

	"github.com/sirkon/cmd-tools/cmd/llmctx/internal/metadata"
)

func main() {
	var cli CLI

	parser := kong.Must(
		&cli,
		kong.Name(appName),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Summary: true,
			Compact: true,
		}),
	)
	ctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		parser.FatalIfErrorf(err)
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		message.Fatal(errors.Wrap(err, "failed to get cache dir"))
	}

	appCacheDir := filepath.Join(cacheDir, appName)
	logFilePath := filepath.Join(appCacheDir, "llmctx.bin")

	context := &runContext{
		cli: &cli,

		fileEdit: func(name string) error {
			ep := editor()
			opts := append(ep.Flags, name)
			cmd := exec.Command(ep.Binary, opts...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}

			return nil
		},
		interpreter: func() (_ *MetadataViewer, err error) {
			viw := &MetadataViewer{
				ops: make(map[string]ContextData),
			}

			src, err := os.Open(logFilePath)
			if err != nil {
				if os.IsNotExist(err) {
					return viw, nil
				}

				return nil, errors.Wrap(err, "open metadata file for reading")
			}
			defer func() {
				if err := src.Close(); err != nil {
					message.Warning(errors.Wrap(err, "failed to close metadata file"))
				}
			}()
			bufReader := bufio.NewReader(src)

			var buf []byte
			var offset int64
			defer func() {
				if err == nil {
					return
				}

				// Получили ошибку на позиции offset+
				// Отрезаем файл вплоть до этой позиции.
				message.Debug("failed to read metadata file, will try to fix it for further usage")
				if truncateError := os.Truncate(logFilePath, offset); truncateError != nil {
					message.Error(errors.Wrap(truncateError, "failed to truncate log file"))
				} else {
					message.Info("successfully fixed metadata file")
				}
			}()

			for {
				len64, err := binary.ReadUvarint(bufReader)
				if err != nil {
					if err == io.EOF {
						break
					}

					return nil, errors.Wrap(err, "read log record length")
				}

				length := int(len64)
				if length > cap(buf) {
					buf = make([]byte, length)
				}

				if _, err := io.ReadFull(bufReader, buf[:length]); err != nil {
					if err == io.EOF {
						return nil, io.ErrUnexpectedEOF // Потому что длина была - должна быть и запись.
					}

					return nil, errors.Wrap(err, "read log record")
				}

				if err := metadata.OperationsLoggerDispatch(viw, buf[:length]); err != nil {
					return nil, errors.Wrap(err, "unmarshal log record")
				}

				offset += int64(varsize.Uint(uint(length))) + int64(length)
			}

			return viw, nil
		},
		inserter: func(data ContextData) (err error) {
			var op metadata.OperationsLogger
			addOp := op.Add(data.Name, data.Desc, data.Path)

			if err := writeOperationData(logFilePath, addOp); err != nil {
				return err
			}

			return nil
		},
		deleter: func(name string) error {
			var op metadata.OperationsLogger
			delOp := op.Del(name)

			if err := writeOperationData(logFilePath, delOp); err != nil {
				return err
			}

			return nil
		},
		cacheRoot: appCacheDir,
	}
	if err := ctx.Run(context); err != nil {
		message.Fatal(errors.Wrap(err, "command failed"))
	}
}

func writeOperationData(path string, buf []byte) (err error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return errors.Wrap(err, "create data directory")
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(err, "open metadata file to update")
	}
	defer func() {
		if cErr := file.Close(); cErr != nil {
			if err == nil {
				err = cErr
			} else {
				message.Error(errors.Wrap(err, "failed to close metadata file after update"))
			}
		}
	}()

	if _, err := file.Write(buf); err != nil {
		return errors.Wrap(err, "write new context data")
	}

	return nil
}

func copyFile(srcPath string, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return errors.Wrap(err, "failed to open file context data file")
	}
	defer func() {
		if cErr := src.Close(); cErr != nil {
			if err == nil {
				err = cErr
			} else {
				message.Error(errors.Wrap(err, "failed to close context source file"))
			}
		}
	}()

	dst, err := os.Create(dstPath)
	if err != nil {
		return errors.Wrap(err, "failed to create destination file for context")
	}
	defer func() {
		if cErr := dst.Close(); cErr != nil {
			if err == nil {
				err = cErr
			} else {
				message.Error(errors.Wrap(err, "failed to close destination file for context"))
			}
		}
	}()

	if _, err := io.Copy(dst, src); err != nil {
		return errors.Wrap(err, "copy context data to destination file")
	}
	return nil
}
