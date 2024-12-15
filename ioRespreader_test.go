package goresp

import (
	"strconv"
	"strings"
	"testing"
)

// TODO: fix the problems with the test
func TestRespIo_readArray_EmptyArray(t *testing.T) {
	input := "*0\r\n"
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
	input := "*1000\r\n"
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
	input := "*2\r\n*3\r\n:1\r\n:2\r\n:3\r\n*2\r\n+Hello\r\n+World\r\n"
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
		if v.Typ != "bulk" || v.Num != int16(i+1) {
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
	input := "*-5\r\n"
	reader := NewRespIo(strings.NewReader(input))

	_, err := reader.readArray()

	if err == nil {
		t.Error("Expected an error for negative array length, but got nil")
	}

	expectedError := "strconv.ParseInt: parsing \"-5\": value out of range"
	if err.Error() != expectedError {
		t.Errorf("Expected error message '%s', but got '%s'", expectedError, err.Error())
	}
}

func TestRespIo_readArray_MixedDataTypes(t *testing.T) {
	input := "*5\r\n:42\r\n$5\r\nhello\r\n+world\r\n*2\r\n:1\r\n:2\r\n-Error message\r\n"
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
	if result.Array[0].Typ != "bulk" || result.Array[0].Num != 42 {
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
	if result.Array[4].Typ != "string" || result.Array[4].Str != "Error message" {
		t.Errorf("Expected fifth element to be error 'Error message', got type %s and value '%s'", result.Array[4].Typ, result.Array[4].Str)
	}
}

func TestRespIo_readArray_MaxLength(t *testing.T) {
	// Create an array with MaxInt32 elements (2147483647)
	input := "*2147483647\r\n"
	for i := 0; i < 2147483647; i++ {
		input += ":1\r\n"
	}

	reader := NewRespIo(strings.NewReader(input))

	result, err := reader.readArray()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Typ != "array" {
		t.Errorf("Expected type to be 'array', got %s", result.Typ)
	}

	if len(result.Array) != 2147483647 {
		t.Errorf("Expected array length 2147483647, got %d", len(result.Array))
	}

	// Check a few random elements to ensure they're correct
	checkIndices := []int{0, 1000000, 2147483646}
	for _, idx := range checkIndices {
		if result.Array[idx].Typ != "bulk" || result.Array[idx].Num != 1 {
			t.Errorf("Expected element at index %d to be number 1, got type %s and value %d", idx, result.Array[idx].Typ, result.Array[idx].Num)
		}
	}
}

func TestRespIo_readArray_UnexpectedEOF(t *testing.T) {
	input := "*3\r\n:1\r\n:2\r\n"
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
		if v.Typ != "bulk" || v.Num != expectedValues[i] {
			t.Errorf("Expected element %d to be number %d, got type %s and value %d", i, expectedValues[i], v.Typ, v.Num)
		}
	}
}

func TestRespIo_readArray_WithNullValues(t *testing.T) {
	input := "*3\r\n$4\r\ntest\r\n_\r\n$5\r\nhello\r\n"
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
	input := "*3\r\n$4\r\n🍕\r\n$6\r\n世界\r\n$7\r\nπ≈3.14\r\n"
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
