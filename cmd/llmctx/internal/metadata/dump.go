package metadata

type Dump interface {
	Deleted(name string, description string, data []byte) error
}

type DumpLogger struct {
}

func (w *DumpLogger) allocateBuffer(n int) []byte {
	return make([]byte, 0, n)
}
