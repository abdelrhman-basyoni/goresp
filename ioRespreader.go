package goresp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

type RespIo struct {
	reader *bufio.Reader
}

func NewRespIo(rd io.Reader) *RespIo {
	return &RespIo{reader: bufio.NewReader(rd)}
}

func (r *RespIo) readLine() (line []byte, n int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}
	return line[:len(line)-2], n, nil
}

func (r *RespIo) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}
	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, n, err
	}
	return int(i64), n, nil
}

func (r *RespIo) Read() (Value, error) {
	_type, err := r.reader.ReadByte()

	if err != nil {
		return Value{}, err
	}
	fmt.Println(string(_type))
	switch _type {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	case INTEGER:
		return r.readNumber()
	case STRING:
		return r.readString()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

func (r *RespIo) readArray() (Value, error) {
	v := Value{}
	v.Typ = "array"

	// read length of Array
	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	// foreach line, parse and read the value
	v.Array = make([]Value, 0)
	for i := 0; i < len; i++ {
		val, err := r.Read()
		if err != nil {
			return v, err
		}

		// append parsed value to Array
		v.Array = append(v.Array, val)
	}

	return v, nil
}

func (r *RespIo) readBulk() (Value, error) {
	v := Value{}

	v.Typ = "bulk"

	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	Bulk := make([]byte, len)

	r.reader.Read(Bulk)

	v.Bulk = string(Bulk)

	// Read the trailing CRLF
	r.readLine()

	return v, nil
}
func (r *RespIo) readString() (Value, error) {
	v := Value{}

	v.Typ = "string"

	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	Str := make([]byte, len)

	r.reader.Read(Str)

	v.Str = string(Str)

	// Read the trailing CRLF
	r.readLine()

	return v, nil
}

func (r *RespIo) readNumber() (Value, error) {
	v := Value{}

	v.Typ = "bulk"

	len, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	Num := make([]byte, len)

	r.reader.Read(Num)

	n, err := strconv.Atoi(string(Num))
	if err != nil {
		return v, err
	}

	v.Num = int16(n)
	return v, nil
}

// Writer

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}
}

func (w *Writer) Write(v Value) error {
	var bytes = v.Marshal()

	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}
