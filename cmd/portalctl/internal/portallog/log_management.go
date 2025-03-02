package portallog

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
	"path/filepath"

	"github.com/sirkon/errors"
	"github.com/sirkon/message"
	"github.com/sirkon/varsize"
)

// LogRead читает лог операций и возвращает вычитанный словарь.
// При необходимости осуществляет "починку" файла, отрезая всё, что не удалось
// вычитать или декодировать.
func LogRead(src string) (map[string]string, error) {
	consumer := logConsume{
		data: map[string]string{},
	}

	if err := processOpLogFile(src, &consumer); err != nil {
		return nil, errors.Wrap(err, "read portals data")
	}

	return consumer.data, nil
}

// LogShowPortalPath поиск пути портала с заданным именем.
// При необходимости осуществляет "починку" файла, отрезая всё, что не удалось
// вычитать или декодировать.
func LogShowPortalPath(src string, portal string) (res string, _ error) {
	consumer := logFind{
		name: portal,
		path: "",
	}
	if err := processOpLogFile(src, &consumer); err != nil {
		return res, errors.Wrap(err, "read portals data")
	}

	if consumer.path == "" {
		return res, errors.New("not found")
	}

	return consumer.path, nil
}

// LogFilter поиск порталов с заданным префиксом в имени.
// При необходимости осуществляет "починку" файла, отрезая всё, что не удалось
// вычитать или декодировать.
func LogFilter(src string, prefix string) ([]string, error) {
	consumer := logFilter{
		data:   map[string]struct{}{},
		prefix: prefix,
	}

	if err := processOpLogFile(src, &consumer); err != nil {
		return nil, errors.Wrap(err, "read portals data")
	}

	res := make([]string, 0, len(consumer.data))
	for portal := range consumer.data {
		res = append(res, portal)
	}

	return res, nil
}

func processOpLogFile(src string, consumer PortalLogger) (err error) {
	var pos int

	defer func() {
		if err == nil {
			return
		}

		if err != nil && pos < 0 {
			return
		}

		if err := os.Truncate(src, int64(pos)); err != nil {
			message.Warning(errors.Wrap(err, "truncate malformed op log file"))
		}

		err = nil
	}()

	file, err := os.Open(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return errors.Wrap(err, "open op log file")
	}
	defer func() {
		if cErr := file.Close(); cErr != nil {
			pos = -1
			cErr = errors.Wrap(cErr, "close op log read file")
			if err == nil {
				err = cErr
				return
			}
			message.Warning(cErr)
		}
	}()

	reader := bufio.NewReader(file)
	var buf []byte
	for {
		l, err := binary.ReadUvarint(reader)
		if err == io.EOF {
			return nil
		}

		if cap(buf) >= int(l) {
			buf = buf[:l]
		} else {
			buf = make([]byte, l)
		}

		if _, err := io.ReadFull(reader, buf); err != nil {
			return errors.Wrap(err, "read op record")
		}

		if err := EncodingDispatch(consumer, buf); err != nil {
			return errors.Wrap(err, "dispatch log operation")
		}

		pos += varsize.Uint(l) + len(buf)
	}
}

// LogDump компактизация существующего лога операций за счёт изъятия данных об удалённых записях.
func LogDump(dst string, cfgRoot string, data map[string]string) error {
	tmpOpLog := filepath.Join(cfgRoot, ".tmp")
	tmp, err := os.Create(tmpOpLog)
	if err != nil {
		return errors.Wrap(err, "create temporary op log file")
	}
	defer func() {
		if tmp == nil {
			return
		}

		if err := tmp.Close(); err != nil {
			message.Error(errors.Wrap(err, "close temporary op log file after an error write"))
		}

		_ = os.RemoveAll(tmpOpLog)
	}()

	writer := bufio.NewWriter(tmp)
	var enc Encoding
	for name, path := range data {
		buf := enc.AddPortal(name, path)
		if _, err := writer.Write(buf); err != nil {
			return errors.Wrap(err, "append add operation")
		}
	}

	if err := writer.Flush(); err != nil {
		return errors.Wrap(err, "flush encoded data")
	}

	ttmp := tmp
	tmp = nil
	if err := ttmp.Close(); err != nil {
		return errors.Wrap(err, "close temporary op log file")
	}

	if err := os.Rename(tmpOpLog, dst); err != nil {
		return errors.Wrap(err, "replace current op log file with the freshly built data")
	}

	return nil
}
