package metadata

type Operations interface {
	Add(name string, description string, path string) error
	Del(name string) error
}

type OperationsLogger struct {
}

func (w *OperationsLogger) allocateBuffer(n int) []byte {
	return make([]byte, 0, n)
}
