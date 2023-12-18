package abstractions

import "os"

type FileReader interface {
	ReadFile(name string) ([]byte, error)
}

func NewFileReader() FileReader {
	return &fileReader{}
}

type fileReader struct {
}

func (f *fileReader) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}
