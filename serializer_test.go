package goresp

import (
	"strings"
	"testing"
)

func TestSerializeCommandSingleWord(t *testing.T) {
	command := "PING"
	expected := []byte("*1\r\n$4\r\nPING\r\n")

	result := SerializeCommand(command)

	if string(result) != string(expected) {
		t.Errorf("SerializeCommand(%q) = %q, want %q", command, result, expected)
	}
}

func TestSerializeCommandMultiWords(t *testing.T) {
	command := "SET key1 value1"
	expected := "*3\r\n$3\r\nSET\r\n$4\r\nkey1\r\n$6\r\nvalue1\r\n"

	result := SerializeCommand(command)

	if string(result) != expected {
		t.Errorf("SerializeCommand(%q) = %q, expected %q", command, string(result), expected)
	}
}

func TestSerializeCommandEmptyInput(t *testing.T) {
	result := SerializeCommand("")
	expected := []byte("*0\r\n")

	if string(result) != string(expected) {
		t.Errorf("SerializeCommand(\"\") = %v, want %v", result, expected)
	}
}

func TestSerializeCommandWithSpaces(t *testing.T) {
	command := "  SET key value  "
	expected := "*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n"

	result := SerializeCommand(command)

	if string(result) != expected {
		t.Errorf("SerializeCommand(%q) = %q, want %q", command, string(result), expected)
	}
}

func TestSerializeCommandWithUnicodeCharacters(t *testing.T) {
	command := "SET 你好 世界"
	expected := "*3\r\n$3\r\nSET\r\n$6\r\n你好\r\n$6\r\n世界\r\n"

	result := SerializeCommand(command)

	if string(result) != expected {
		t.Errorf("SerializeCommand(%q) = %q, want %q", command, string(result), expected)
	}
}

// SerializeValue Test

func TestSerializeValueBulk(t *testing.T) {
	input := Value{Typ: "bulk", Bulk: "test_bulk"}
	expected := "test_bulk"

	result := SerializeValue(input)

	if result != expected {
		t.Errorf("SerializeValue(%v) = %q, want %q", input, result, expected)
	}
}

func TestSerializeValueString(t *testing.T) {
	input := Value{Typ: "string", Str: "test_string"}
	expected := "test_string"

	result := SerializeValue(input)

	if result != expected {
		t.Errorf("SerializeValue(%v) = %q, want %q", input, result, expected)
	}
}

func TestSerializeValueInt(t *testing.T) {
	input := Value{Typ: "int", Num: 42}
	expected := "42"

	result := SerializeValue(input)

	if result != expected {
		t.Errorf("SerializeValue(%v) = %q, want %q", input, result, expected)
	}
}

func TestSerializeValueEmptyArray(t *testing.T) {
	input := Value{Typ: "array", Array: []Value{}}
	expected := ""

	result := SerializeValue(input)

	if result != expected {
		t.Errorf("SerializeValue(%v) = %q, want %q", input, result, expected)
	}
}

func TestSerializeValueUnknownType(t *testing.T) {
	input := Value{Typ: "unknown", Bulk: "test_unknown"}
	expected := "test_unknown"

	result := SerializeValue(input)

	if result != expected {
		t.Errorf("SerializeValue(%v) = %q, want %q", input, result, expected)
	}
}

func TestSerializeValueNestedArray(t *testing.T) {
	input := Value{
		Typ: "array",
		Array: []Value{
			{Typ: "bulk", Bulk: "SET"},
			{Typ: "bulk", Bulk: "key"},
			{Typ: "array", Array: []Value{
				{Typ: "int", Num: 1},
				{Typ: "string", Str: "two"},
				{Typ: "bulk", Bulk: "three"},
			}},
		},
	}
	expected := "SETkey1twothree"

	result := SerializeValue(input)

	if result != expected {
		t.Errorf("SerializeValue(%v) = %q, want %q", input, result, expected)
	}
}

func TestSerializeValueLargeInteger(t *testing.T) {
	input := Value{Typ: "int", Num: 9223372036854775807} // Max int64 value
	expected := "9223372036854775807"

	result := SerializeValue(input)

	if result != expected {
		t.Errorf("SerializeValue(%v) = %q, want %q", input, result, expected)
	}
}

func TestSerializeValueEmptyString(t *testing.T) {
	input := Value{Typ: "string", Str: ""}
	expected := ""

	result := SerializeValue(input)

	if result != expected {
		t.Errorf("SerializeValue(%v) = %q, want %q", input, result, expected)
	}
}

func TestSerializeValueVeryLongString(t *testing.T) {
	longString := strings.Repeat("a", 1000000)
	input := Value{Typ: "bulk", Bulk: longString}
	expected := longString

	result := SerializeValue(input)

	if result != expected {
		t.Errorf("SerializeValue(%v) = %q (length: %d), want %q (length: %d)", input, result[:100], len(result), expected[:100], len(expected))
	}
}
