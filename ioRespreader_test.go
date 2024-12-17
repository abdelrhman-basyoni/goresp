package goresp

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

// TODO: fix the problems with the test
func TestRespIo_readArray_EmptyArray(t *testing.T) {
	// we removed the prefix "*" because readArray starts after the Read function remove the message type first
	input := "0\r\n"
	reader := NewRespIo(strings.NewReader(input))

	result, err := reader.readArray()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Typ != "array" {
		t.Errorf("Expected type to be 'array', got %s", result.Typ)
	}

	if len(result.Array) != 0 {
		t.Errorf("Expected empty array, got array of length %d", len(result.Array))
	}
}

func TestRespIo_readArray_LargeArray(t *testing.T) {
	// Create a large array with 1000 elements
	// we removed the prefix "*" because readArray starts after the Read function remove the message type first
	input := "1000\r\n"
	for i := 0; i < 1000; i++ {
		input += "$1\r\n" + strconv.Itoa(i%10) + "\r\n"
	}

	reader := strings.NewReader(input)
	respIo := NewRespIo(reader)

	value, err := respIo.readArray()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if value.Typ != "array" {
		t.Errorf("Expected type 'array', got '%s'", value.Typ)
	}

	if len(value.Array) != 1000 {
		t.Errorf("Expected array length 1000, got %d", len(value.Array))
	}

	for i, v := range value.Array {
		expected := strconv.Itoa(i % 10)
		if v.Bulk != expected {
			t.Errorf("Expected element %d to be '%s', got '%s'", i, expected, v.Bulk)
		}
	}
}

func TestRespIo_readArray_NestedArrays(t *testing.T) {
	// we removed the prefix "*" because readArray starts after the Read function remove the message type first
	input := "2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Hello\r\n+World\r\n"
	reader := NewRespIo(strings.NewReader(input))

	result, err := reader.readArray()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Typ != "array" {
		t.Errorf("Expected type to be 'array', got %s", result.Typ)
	}

	if len(result.Array) != 2 {
		t.Errorf("Expected outer array length 2, got %d", len(result.Array))
	}

	// Check first nested array
	firstNestedArray := result.Array[0]
	if firstNestedArray.Typ != "array" {
		t.Errorf("Expected first nested element to be 'array', got %s", firstNestedArray.Typ)
	}
	if len(firstNestedArray.Array) != 3 {
		t.Errorf("Expected first nested array length 3, got %d", len(firstNestedArray.Array))
	}
	for i, v := range firstNestedArray.Array {
		if v.Typ != "integer" || v.Num != int16(i+1) {
			t.Errorf("Expected element %d to be number %d, got type %s and value %d", i, i+1, v.Typ, v.Num)
		}
	}

	// Check second nested array
	secondNestedArray := result.Array[1]
	if secondNestedArray.Typ != "array" {
		t.Errorf("Expected second nested element to be 'array', got %s", secondNestedArray.Typ)
	}
	if len(secondNestedArray.Array) != 2 {
		t.Errorf("Expected second nested array length 2, got %d", len(secondNestedArray.Array))
	}
	expectedStrings := []string{"Hello", "World"}
	for i, v := range secondNestedArray.Array {
		if v.Typ != "string" || v.Str != expectedStrings[i] {
			t.Errorf("Expected element %d to be string '%s', got type %s and value '%s'", i, expectedStrings[i], v.Typ, v.Str)
		}
	}
}

func TestRespIo_readArray_NegativeLength(t *testing.T) {
	// we removed the prefix "*" because readArray starts after the Read function remove the message type first
	input := "-5\r\n"
	reader := NewRespIo(strings.NewReader(input))

	_, err := reader.readArray()

	if err == nil {
		t.Error("Expected an error for negative array length, but got nil")
	}

	expectedError := "Array length cant be negative"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', but got '%s'", expectedError, err.Error())
	}
}

func TestRespIo_readArray_MixedDataTypes(t *testing.T) {
	// we removed the prefix "*" because readArray starts after the Read function remove the message type first
	input := "5\r\n:42\r\n$5\r\nhello\r\n+world\r\n*2\r\n:1\r\n:2\r\n-Error message\r\n"
	reader := NewRespIo(strings.NewReader(input))

	result, err := reader.readArray()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Typ != "array" {
		t.Errorf("Expected type to be 'array', got %s", result.Typ)
	}

	if len(result.Array) != 5 {
		t.Errorf("Expected array length 5, got %d", len(result.Array))
	}

	// Check integer
	if result.Array[0].Typ != "integer" || result.Array[0].Num != 42 {
		t.Errorf("Expected first element to be integer 42, got type %s and value %d", result.Array[0].Typ, result.Array[0].Num)
	}

	// Check bulk string
	if result.Array[1].Typ != "bulk" || result.Array[1].Bulk != "hello" {
		t.Errorf("Expected second element to be bulk string 'hello', got type %s and value '%s'", result.Array[1].Typ, result.Array[1].Bulk)
	}

	// Check simple string
	if result.Array[2].Typ != "string" || result.Array[2].Str != "world" {
		t.Errorf("Expected third element to be simple string 'world', got type %s and value '%s'", result.Array[2].Typ, result.Array[2].Str)
	}

	// Check nested array
	if result.Array[3].Typ != "array" || len(result.Array[3].Array) != 2 {
		t.Errorf("Expected fourth element to be array of length 2, got type %s and length %d", result.Array[3].Typ, len(result.Array[3].Array))
	}

	// Check error
	if result.Array[4].Typ != "error" || result.Array[4].Str != "Error message" {
		t.Errorf("Expected fifth element to be error 'Error message', got type %s and value '%s'", result.Array[4].Typ, result.Array[4].Str)
	}
}

func TestRespIo_readArray_UnexpectedEOF(t *testing.T) {
	input := "3\r\n:1\r\n:2\r\n"
	reader := NewRespIo(strings.NewReader(input))

	result, err := reader.readArray()

	if err == nil {
		t.Errorf("Expected an error, got nil")
	}

	if !strings.Contains(err.Error(), "EOF") {
		t.Errorf("Expected EOF error, got: %v", err)
	}

	if result.Typ != "array" {
		t.Errorf("Expected type to be 'array', got %s", result.Typ)
	}

	if len(result.Array) != 2 {
		t.Errorf("Expected array length 2, got %d", len(result.Array))
	}

	expectedValues := []int16{1, 2}
	for i, v := range result.Array {
		if v.Typ != "integer" || v.Num != expectedValues[i] {
			t.Errorf("Expected element %d to be number %d, got type %s and value %d", i, expectedValues[i], v.Typ, v.Num)
		}
	}
}

func TestRespIo_readArray_WithNullValues(t *testing.T) {
	input := "3\r\n$4\r\ntest\r\n_\r\n$5\r\nhello\r\n"
	reader := NewRespIo(strings.NewReader(input))

	result, err := reader.readArray()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Typ != "array" {
		t.Errorf("Expected type to be 'array', got %s", result.Typ)
	}

	if len(result.Array) != 3 {
		t.Errorf("Expected array length 3, got %d", len(result.Array))
	}

	expectedValues := []struct {
		typ  string
		bulk string
	}{
		{"bulk", "test"},
		{"null", ""},
		{"bulk", "hello"},
	}

	for i, expected := range expectedValues {
		if result.Array[i].Typ != expected.typ {
			t.Errorf("Element %d: expected type %s, got %s", i, expected.typ, result.Array[i].Typ)
		}
		if expected.typ == "bulk" && result.Array[i].Bulk != expected.bulk {
			t.Errorf("Element %d: expected bulk value %s, got %s", i, expected.bulk, result.Array[i].Bulk)
		}
	}
}

func TestRespIo_readArray_WithUnicodeCharacters(t *testing.T) {
	input := "3\r\n$4\r\n🍕\r\n$6\r\n世界\r\n$9\r\nπ≈3.14\r\n"
	reader := NewRespIo(strings.NewReader(input))

	result, err := reader.readArray()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Typ != "array" {
		t.Errorf("Expected type to be 'array', got %s", result.Typ)
	}

	if len(result.Array) != 3 {
		t.Errorf("Expected array length 3, got %d", len(result.Array))
	}

	expectedValues := []string{"🍕", "世界", "π≈3.14"}
	for i, expected := range expectedValues {
		if result.Array[i].Typ != "bulk" {
			t.Errorf("Element %d: expected type 'bulk', got %s", i, result.Array[i].Typ)
		}
		if result.Array[i].Bulk != expected {
			t.Errorf("Element %d: expected value '%s', got '%s'", i, expected, result.Array[i].Bulk)
		}
	}
}

// testing bulks

func TestRespIo_readBulk_ZeroLength(t *testing.T) {
	input := "0\r\n\r\n"
	reader := NewRespIo(strings.NewReader(input))

	result, err := reader.readBulk()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Typ != "bulk" {
		t.Errorf("Expected type to be 'bulk', got %s", result.Typ)
	}

	if result.Bulk != "" {
		t.Errorf("Expected empty bulk string, got '%s'", result.Bulk)
	}
}

func TestRespIo_readBulk_NegativeLength(t *testing.T) {
	input := "-5\r\nHello\r\n"
	reader := NewRespIo(strings.NewReader(input))

	_, err := reader.readBulk()

	if err == nil {
		t.Error("Expected an error for negative bulk string length, but got nil")
	}

	expectedError := "Bulk length cant be negative"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', but got '%s'", expectedError, err.Error())
	}
}

func TestRespIo_readBulk_WhitespaceOnly(t *testing.T) {
	input := "5\r\n     \r\n"
	reader := NewRespIo(strings.NewReader(input))

	result, err := reader.readBulk()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Typ != "bulk" {
		t.Errorf("Expected type to be 'bulk', got %s", result.Typ)
	}

	expectedBulk := "     "
	if result.Bulk != expectedBulk {
		t.Errorf("Expected bulk value '%s', got '%s'", expectedBulk, result.Bulk)
	}

	if len(result.Bulk) != 5 {
		t.Errorf("Expected bulk length 5, got %d", len(result.Bulk))
	}
}

func TestRespIo_readBulk_NonASCIICharacters(t *testing.T) {
	input := "14\r\nHello, 世界!\r\n"

	reader := NewRespIo(strings.NewReader(input))

	result, err := reader.readBulk()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Typ != "bulk" {
		t.Errorf("Expected type to be 'bulk', got %s", result.Typ)
	}

	expectedBulk := "Hello, 世界!"
	if result.Bulk != expectedBulk {
		t.Errorf("Expected bulk value '%s', got '%s'", expectedBulk, result.Bulk)
	}

	if len(result.Bulk) != 14 {
		t.Errorf("Expected bulk length 14, got %d", len(result.Bulk))
	}
}

func TestRespIo_readBulk_LengthMismatch(t *testing.T) {
	input := "6\r\nHello\r\n"
	reader := NewRespIo(strings.NewReader(input))

	_, err := reader.readBulk()

	if err == nil {
		t.Error("Expected an error for length mismatch, but got nil")
	}

	expectedError := "Error reading trailing CRLF: EOF"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error message to contain '%s', but got '%s'", expectedError, err.Error())
	}
}

func TestRespIo_readBulk_MaxLength(t *testing.T) {
	// Maximum allowed length for a bulk string (512MB)
	maxLength := 512 * 1024 * 1024
	input := fmt.Sprintf("%d\r\n%s\r\n", maxLength, strings.Repeat("a", maxLength))
	reader := NewRespIo(strings.NewReader(input))
	result, err := reader.readBulk()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Typ != "bulk" {
		t.Errorf("Expected type to be 'bulk', got %s", result.Typ)
	}

	if len(result.Bulk) != maxLength {
		t.Errorf("Expected bulk length %d, got %d", maxLength, len(result.Bulk))
	}

	expectedBulk := strings.Repeat("a", maxLength)
	if result.Bulk != expectedBulk {
		t.Errorf("Expected bulk value of %d 'a' characters, got different content", maxLength)
	}
}

func TestRespIo_readBulk_WithNullBytes(t *testing.T) {
	input := "8\r\nHello\x00\x00!\r\n"
	reader := NewRespIo(strings.NewReader(input))

	result, err := reader.readBulk()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Typ != "bulk" {
		t.Errorf("Expected type to be 'bulk', got %s", result.Typ)
	}

	expectedBulk := "Hello\x00\x00!"
	if result.Bulk != expectedBulk {
		t.Errorf("Expected bulk value %q, got %q", expectedBulk, result.Bulk)
	}

	if len(result.Bulk) != 8 {
		t.Errorf("Expected bulk length 8, got %d", len(result.Bulk))
	}
}
func TestRespIo_readBulk_EOFBeforeCompleteRead(t *testing.T) {
	input := "10\r\nHello"
	reader := NewRespIo(strings.NewReader(input))

	_, err := reader.readBulk()

	if err == nil {
		t.Error("Expected an error when EOF is reached before reading the entire bulk string, but got nil")
	}

	if err == errors.New("Error reading bulk data: EOF") {
		t.Errorf("Expected EOF error, got: %v", err)
	}
}

func TestRespIo_readBulk_ConsecutiveReads(t *testing.T) {
	input := "5\r\nHello\r\n6\r\nWorld!\r\n"
	reader := NewRespIo(strings.NewReader(input))

	// First read
	result1, err := reader.readBulk()
	if err != nil {
		t.Errorf("First read: Expected no error, got %v", err)
	}
	if result1.Typ != "bulk" {
		t.Errorf("First read: Expected type to be 'bulk', got %s", result1.Typ)
	}
	if result1.Bulk != "Hello" {
		t.Errorf("First read: Expected bulk value 'Hello', got '%s'", result1.Bulk)
	}

	// Second read
	result2, err := reader.readBulk()
	if err != nil {
		t.Errorf("Second read: Expected no error, got %v", err)
	}
	if result2.Typ != "bulk" {
		t.Errorf("Second read: Expected type to be 'bulk', got %s", result2.Typ)
	}
	if result2.Bulk != "World!" {
		t.Errorf("Second read: Expected bulk value 'World!', got '%s'", result2.Bulk)
	}
}

func TestRespIo_readBulk_TrailingCRLF(t *testing.T) {
	input := "5\r\nHello\r\nExtra"
	reader := NewRespIo(strings.NewReader(input))

	result, err := reader.readBulk()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result.Typ != "bulk" {
		t.Errorf("Expected type to be 'bulk', got %s", result.Typ)
	}

	expectedBulk := "Hello"
	if result.Bulk != expectedBulk {
		t.Errorf("Expected bulk value '%s', got '%s'", expectedBulk, result.Bulk)
	}

	// Check if the trailing CRLF was read correctly
	nextByte, err := reader.reader.ReadByte()
	if err != nil {
		t.Errorf("Expected no error reading next byte, got %v", err)
	}
	if nextByte != 'E' {
		t.Errorf("Expected next byte to be 'E', got '%c'", nextByte)
	}
}
