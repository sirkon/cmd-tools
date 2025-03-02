package portallog

import (
	"os"
	"strings"

	"github.com/sirkon/errors"
	"github.com/sirkon/message"
)

// NewLog создание лога который будет писать операции в файл с заданным именем.
func NewLog(dstPath string) *LogRecord {
	return &LogRecord{dstPath: dstPath}
}

// LogRecord сущность для записи логов.
type LogRecord struct {
	dstPath string
}

func (r *LogRecord) appendOperation(buf []byte) (err error) {
	file, err := os.OpenFile(r.dstPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrap(err, "open file to append an op")
	}
	defer func() {
		if cErr := file.Close(); cErr != nil {
			if err == nil {
				err = errors.Wrap(cErr, "close op log file write")
			} else {
				message.Error(errors.Wrap(cErr, "close op log file write"))
			}
		}
	}()

	if _, err := file.Write(buf); err != nil {
		return errors.Wrap(err, "write operation data")
	}

	return nil
}

// AddPortal для реализации PortalLogger.
func (r *LogRecord) AddPortal(name string, path string) error {
	var e Encoding
	buf := e.AddPortal(name, path)
	if err := r.appendOperation(buf); err != nil {
		return err
	}

	return nil
}

// DeletePortal для реализации PortalLogger.
func (r *LogRecord) DeletePortal(name string) error {
	var e Encoding
	buf := e.DeletePortal(name)
	if err := r.appendOperation(buf); err != nil {
		return err
	}

	return nil
}

type logConsume struct {
	data map[string]string
}

// AddPortal для реализации PortalLogger.
func (l *logConsume) AddPortal(name string, path string) error {
	if l.data == nil {
		l.data = map[string]string{}
	}

	l.data[name] = path
	return nil
}

// DeletePortal для реализации PortalLogger.
func (l *logConsume) DeletePortal(name string) error {
	if l.data == nil {
		l.data = map[string]string{}
	}

	delete(l.data, name)
	return nil
}

type logFind struct {
	name string
	path string
}

// AddPortal для реализации PortalLogger.
func (l *logFind) AddPortal(name string, path string) error {
	if name != l.name {
		return nil
	}

	l.path = path
	return nil
}

// DeletePortal для реализации PortalLogger.
func (l *logFind) DeletePortal(name string) error {
	if name != l.name {
		return nil
	}

	l.path = ""
	return nil
}

type logFilter struct {
	data   map[string]struct{}
	prefix string
}

// AddPortal для реализации PortalLogger.
func (l *logFilter) AddPortal(name string, path string) error {
	if !strings.HasPrefix(name, l.prefix) {
		return nil
	}

	l.data[name] = struct{}{}
	return nil
}

// DeletePortal для реализации PortalLogger.
func (l *logFilter) DeletePortal(name string) error {
	if !strings.HasPrefix(name, l.prefix) {
		return nil
	}

	delete(l.data, name)
	return nil
}
