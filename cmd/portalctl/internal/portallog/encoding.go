package portallog

// Encoding кодирование данных лога.
type Encoding struct {
	buf []byte
}

func (l *Encoding) allocateBuffer(n int) []byte {
	if len(l.buf) < n {
		l.buf = make([]byte, 0, n)
	}

	return l.buf[:0]
}
