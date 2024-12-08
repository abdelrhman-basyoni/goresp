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

// reads and returns Value of type Bulk
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
