package goresp

import (
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
