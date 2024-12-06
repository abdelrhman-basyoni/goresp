package goresp

import "io"

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

type RespReader interface {
	Read() (Value, error)
}

func NewRespReader(reader io.Reader) RespReader {

	return NewRespIo(reader)
}

// writer
type RespWriter interface {
	Write(v Value) error
}

type BasicWriter struct {
	writer io.Writer
}

func NewBasicWriter(w io.Writer) *BasicWriter {
	return &BasicWriter{writer: w}
}

func (w *BasicWriter) Write(bytes []byte) error {

	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}
