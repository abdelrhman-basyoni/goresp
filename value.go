package goresp

import (
	"strconv"
)

type Value struct {
	Typ   string
	Str   string
	Num   int16
	Bulk  string
	Array []Value
}

// The function Marshal in the provided code is likely named after the concept of marshaling,
// which is a common term in computer science and programming.
// Marshaling refers to the process of converting data from one data structure into a format that can be easily stored, transmitted, or reconstructed in another data structure.
// In this context, the Marshal function is responsible for converting a "Value" object into a byte representation that adheres to the RESP (REdis Serialization Protocol) format.
func (v Value) Marshal() []byte {
	switch v.Typ {
	case "array":
		return v.marshalArray()
	case "bulk":
		return v.marshalBulk()
	case "string":
		return v.marshalString()
	case "int":
		return v.marshalNum()
	case "null":
		return v.marshallNull()
	case "error":
		return v.marshallError()
	default:
		return []byte{}
	}
}

func (v Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, STRING)
	bytes = append(bytes, v.Str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalNum() []byte {

	var bytes []byte

	bytes = append(bytes, INTEGER)
	bytes = append(bytes, strconv.FormatInt(int64(v.Num), 10)...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(len(v.Bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.Bulk...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalArray() []byte {
	len := len(v.Array)
	var bytes []byte
	bytes = append(bytes, ARRAY)
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')

	for i := 0; i < len; i++ {
		bytes = append(bytes, v.Array[i].Marshal()...)
	}

	return bytes
}

func (v Value) marshallError() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	bytes = append(bytes, v.Str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshallNull() []byte {
	return []byte("$-1\r\n")
}

func NewSetValue(key, value string) Value {
	arr := []Value{{Typ: "Bulk", Bulk: "set"}, {Typ: "Bulk", Bulk: key}, {Typ: "Bulk", Bulk: value}}
	val := Value{Typ: "Array", Array: arr}

	return val
}

func NewHsetValue(hash, key, value string) Value {
	arr := []Value{{Typ: "Bulk", Bulk: "hset"}, {Typ: "Bulk", Bulk: key}, {Typ: "Bulk", Bulk: key}, {Typ: "Bulk", Bulk: value}}
	val := Value{Typ: "Array", Array: arr}

	return val
}

func NewDelValue(keys []string) Value {
	arr := []Value{{Typ: "Bulk", Bulk: "del"}}

	for _, key := range keys {
		v := Value{Typ: "Bulk", Bulk: key}

		arr = append(arr, v)
	}
	val := Value{Typ: "Array", Array: arr}

	return val
}

func NewErrorValue(message string) Value {

	val := Value{Typ: "error", Str: message}

	return val
}

func NewNumberValue(number int16) Value {

	val := Value{Typ: "int", Num: number}

	return val
}
