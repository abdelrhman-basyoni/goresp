package goresp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type RespIo struct {
	reader *bufio.Reader
}

func NewRespIo(rd io.Reader) *RespIo {
	return &RespIo{reader: bufio.NewReader(rd)}
}

// it reads aline from the underlying reader, stooping at '\r and returning the line without the  trailing '\r\n'
// it returns the line m number of bytes read,
func (r *RespIo) readLine() (line []byte, numBytes int, err error) {
	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		numBytes += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}

	return line[:len(line)-2], numBytes, nil
}

// it reads  and parses the number value  used to parse the length of the bulk string $<length>\r\n<data>\r\n
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

// the main function the triggers the reading process on the io   resp and it returns a Value object
func (r *RespIo) Read() (Value, error) {
	_type, err := r.reader.ReadByte()

	if err != nil {
		return Value{}, err
	}
	switch _type {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	case INTEGER:
		return r.readNumber()
	case STRING:
		return r.readString()
	case ERROR:
		return r.readError()
	case NULL:
		return r.readNull()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}
}

// reads and returns Value of type Array
func (r *RespIo) readArray() (Value, error) {

	v := Value{}
	v.Typ = "array"

	// read length of Array
	length, _, err := r.readInteger()
	if err != nil {
		return v, err
	}

	// foreach line, parse and read the value
	if length < 0 {
		return v, fmt.Errorf("Array length cant be negative")
	}

	v.Array = make([]Value, 0)
	for i := 0; i < length; i++ {
		val, err := r.Read()
		if err != nil {
			return v, err
		}

		// append parsed value to Array
		v.Array = append(v.Array, val)
	}

	return v, nil
}

// reads and returns Value of type Bulk
func (r *RespIo) readBulk() (Value, error) {
	v := Value{}

	v.Typ = "bulk"

	length, _, err := r.readInteger()
	if err != nil {
		return v, err
	}
	if length < 0 {

		return v, fmt.Errorf("Bulk length cant be negative")

	}

	Bulk := make([]byte, length)

	// to handle the extreme case if the length is bigger than 4096 which  the internal buffer for the bufio read
	totalRead := 0
	for totalRead < length {
		n, err := r.reader.Read(Bulk[totalRead:])
		if err != nil {
			return v, fmt.Errorf("Error reading bulk data: %v", err)
		}
		if n == 0 {
			return v, fmt.Errorf("Unexpected EOF: read %d bytes, expected %d bytes", totalRead, length)
		}
		totalRead += n
	}

	if err != nil {

		return v, err
	}

	v.Bulk = string(Bulk)

	// Read the trailing CRLF
	line, _, err := r.readLine()
	if err != nil {
		return v, fmt.Errorf("Error reading trailing CRLF: %v", err)
	}
	if string(line) != "" {
		return v, fmt.Errorf("Expected CRLF after bulk data, but got '%s'", line)
	}

	return v, nil
}

// reads and returns Value of type simple string
func (r *RespIo) readString() (Value, error) {
	line, _, err := r.readLine()
	v := Value{}
	v.Typ = "string"
	v.Str = string(line)

	if err != nil {
		return v, err
	}
	return v, nil
}

// reads and returns Value of type number
func (r *RespIo) readNumber() (Value, error) {
	v := Value{}

	v.Typ = "integer"

	line, _, err := r.readLine()

	n, err := strconv.Atoi(strings.TrimSpace(string(line)))
	if err != nil {
		return v, err
	}

	v.Num = int64(n)
	return v, nil
}

func (r *RespIo) readError() (Value, error) {
	v := Value{}

	v.Typ = "error"

	line, _, err := r.readLine()

	if err != nil {
		return v, err
	}

	v.Str = string(line)
	return v, nil

}

func (r *RespIo) readNull() (Value, error) {
	v := Value{}
	r.readLine()
	v.Typ = "null"

	return v, nil
}
