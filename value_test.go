package goresp

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestValue_Marshal_String(t *testing.T) {
	v := Value{
		Typ: "string",
		Str: "Hello",
	}
	expected := []byte("+Hello\r\n")
	result := v.Marshal()
	if !bytes.Equal(result, expected) {
		t.Errorf("Marshal() for string type returned %v, expected %v", result, expected)
	}
}

func TestValue_Marshal_Array(t *testing.T) {
	v := Value{
		Typ: "array",
		Array: []Value{
			{Typ: "string", Str: "foo"},
			{Typ: "int", Num: 42},
			{Typ: "bulk", Bulk: "bar"},
		},
	}

	expected := []byte("*3\r\n+foo\r\n:42\r\n$3\r\nbar\r\n")
	result := v.Marshal()

	if !bytes.Equal(result, expected) {
		t.Errorf("Marshal() for array type returned %v, want %v", result, expected)
	}
}

func TestValue_Marshal_BulkType(t *testing.T) {
	v := Value{
		Typ:  "bulk",
		Bulk: "hello",
	}
	expected := []byte("$5\r\nhello\r\n")
	result := v.Marshal()
	if !bytes.Equal(result, expected) {
		t.Errorf("Marshal() for bulk type failed, got: %v, want: %v", result, expected)
	}
}

func TestValue_Marshal_Null(t *testing.T) {
	v := Value{Typ: "null"}
	expected := []byte("$-1\r\n")
	result := v.Marshal()
	if !bytes.Equal(result, expected) {
		t.Errorf("Marshal() for null type returned %v, want %v", result, expected)
	}
}

func TestValue_Marshal_Integer(t *testing.T) {
	v := Value{
		Typ: "int",
		Num: 42,
	}
	expected := []byte(":42\r\n")
	result := v.Marshal()
	if string(result) != string(expected) {
		t.Errorf("Expected %q, but got %q", expected, result)
	}
}

func TestValue_Marshal_Error(t *testing.T) {
	v := Value{
		Typ: "error",
		Str: "Error message",
	}
	expected := []byte("-Error message\r\n")
	result := v.Marshal()
	if !bytes.Equal(result, expected) {
		t.Errorf("Marshal() for error type returned %v, want %v", result, expected)
	}
}

func TestValue_Marshal_UnknownType(t *testing.T) {
	v := Value{
		Typ: "unknown",
	}
	expected := []byte{}
	result := v.Marshal()
	if !bytes.Equal(result, expected) {
		t.Errorf("Marshal() for unknown type returned %v, want %v", result, expected)
	}
}

func TestValue_Marshal_EmptyArray(t *testing.T) {
	v := Value{
		Typ:   "array",
		Array: []Value{},
	}
	expected := []byte("*0\r\n")
	result := v.Marshal()
	if !bytes.Equal(result, expected) {
		t.Errorf("Marshal() for empty array returned %v, want %v", result, expected)
	}
}

func TestValue_Marshal_LargeInteger(t *testing.T) {
	v := Value{
		Typ: "int",
		Num: 32767, // Maximum value for int16
	}
	expected := []byte(":32767\r\n")
	result := v.Marshal()
	if !bytes.Equal(result, expected) {
		t.Errorf("Marshal() for large integer returned %v, want %v", result, expected)
	}
}

func TestValue_Marshal_Consistency(t *testing.T) {
	testCases := []struct {
		name     string
		value    Value
		expected []byte
	}{
		{
			name:     "Array",
			value:    Value{Typ: "array", Array: []Value{{Typ: "string", Str: "test"}}},
			expected: []byte("*1\r\n+test\r\n"),
		},
		{
			name:     "Bulk",
			value:    Value{Typ: "bulk", Bulk: "hello"},
			expected: []byte("$5\r\nhello\r\n"),
		},
		{
			name:     "String",
			value:    Value{Typ: "string", Str: "world"},
			expected: []byte("+world\r\n"),
		},
		{
			name:     "Integer",
			value:    Value{Typ: "int", Num: 42},
			expected: []byte(":42\r\n"),
		},
		{
			name:     "Null",
			value:    Value{Typ: "null"},
			expected: []byte("$-1\r\n"),
		},
		{
			name:     "Error",
			value:    Value{Typ: "error", Str: "Error message"},
			expected: []byte("-Error message\r\n"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.value.Marshal()
			if !bytes.Equal(result, tc.expected) {
				t.Errorf("Marshal() returned %v, want %v", result, tc.expected)
			}

			var specificResult []byte
			switch tc.value.Typ {
			case "array":
				specificResult = tc.value.marshalArray()
			case "bulk":
				specificResult = tc.value.marshalBulk()
			case "string":
				specificResult = tc.value.marshalString()
			case "int":
				specificResult = tc.value.marshalNum()
			case "null":
				specificResult = tc.value.marshallNull()
			case "error":
				specificResult = tc.value.marshallError()
			}

			if !bytes.Equal(result, specificResult) {
				t.Errorf("Marshal() and specific marshal function returned different results. Marshal: %v, Specific: %v", result, specificResult)
			}
		})
	}
}

// newSetValue
func TestNewSetValue_EmptyKeyValue(t *testing.T) {
	result := NewSetValue("", "")
	expected := Value{
		Typ: "array",
		Array: []Value{
			{Typ: "bulk", Bulk: "set"},
			{Typ: "bulk", Bulk: ""},
			{Typ: "bulk", Bulk: ""},
		},
	}

	if result.Typ != expected.Typ {
		t.Errorf("NewSetValue() returned incorrect type. Got %s, want %s", result.Typ, expected.Typ)
	}

	if len(result.Array) != len(expected.Array) {
		t.Errorf("NewSetValue() returned array of incorrect length. Got %d, want %d", len(result.Array), len(expected.Array))
	}

	for i, v := range result.Array {
		if v.Typ != expected.Array[i].Typ || v.Bulk != expected.Array[i].Bulk {
			t.Errorf("NewSetValue() returned incorrect value at index %d. Got {Typ: %s, Bulk: %s}, want {Typ: %s, Bulk: %s}",
				i, v.Typ, v.Bulk, expected.Array[i].Typ, expected.Array[i].Bulk)
		}
	}
}
func TestNewSetValue_WithUnicode(t *testing.T) {
	key := "你好"
	value := "世界"
	expected := Value{
		Typ: "array",
		Array: []Value{
			{Typ: "bulk", Bulk: "set"},
			{Typ: "bulk", Bulk: "你好"},
			{Typ: "bulk", Bulk: "世界"},
		},
	}

	result := NewSetValue(key, value)

	if result.Typ != expected.Typ {
		t.Errorf("NewSetValue() Typ = %v, want %v", result.Typ, expected.Typ)
	}

	if len(result.Array) != len(expected.Array) {
		t.Errorf("NewSetValue() Array length = %d, want %d", len(result.Array), len(expected.Array))
	}

	for i, v := range result.Array {
		if v.Typ != expected.Array[i].Typ || v.Bulk != expected.Array[i].Bulk {
			t.Errorf("NewSetValue() Array[%d] = {Typ: %v, Bulk: %v}, want {Typ: %v, Bulk: %v}",
				i, v.Typ, v.Bulk, expected.Array[i].Typ, expected.Array[i].Bulk)
		}
	}
}
func TestNewSetValue(t *testing.T) {
	key := "testKey"
	value := "testValue"

	result := NewSetValue(key, value)

	expectedArray := []Value{
		{Typ: "bulk", Bulk: "set"},
		{Typ: "bulk", Bulk: key},
		{Typ: "bulk", Bulk: value},
	}
	expected := Value{Typ: "array", Array: expectedArray}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("NewSetValue(%q, %q) = %v, want %v", key, value, result, expected)
	}
}
func TestNewSetValue_VeryLongStrings(t *testing.T) {
	longKey := strings.Repeat("a", 1000)
	longValue := strings.Repeat("b", 1000)

	result := NewSetValue(longKey, longValue)

	expectedArray := []Value{
		{Typ: "bulk", Bulk: "set"},
		{Typ: "bulk", Bulk: longKey},
		{Typ: "bulk", Bulk: longValue},
	}
	expected := Value{Typ: "array", Array: expectedArray}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("NewSetValue with long strings failed.\nGot: %+v\nWant: %+v", result, expected)
	}

	if len(result.Array[1].Bulk) != 1000 || len(result.Array[2].Bulk) != 1000 {
		t.Errorf("NewSetValue did not preserve long string lengths.\nKey length: %d, Value length: %d",
			len(result.Array[1].Bulk), len(result.Array[2].Bulk))
	}
}
func TestNewSetValue_ReturnsValueType(t *testing.T) {
	result := NewSetValue("testKey", "testValue")

	if reflect.TypeOf(result) != reflect.TypeOf(Value{}) {
		t.Errorf("NewSetValue() returned type %v, want Value", reflect.TypeOf(result))
	}
}

// test NewHset
func TestNewHsetValue_EmptyStrings(t *testing.T) {
	result := NewHsetValue("", "", "")
	expected := Value{
		Typ: "array",
		Array: []Value{
			{Typ: "bulk", Bulk: "hset"},
			{Typ: "bulk", Bulk: ""},
			{Typ: "bulk", Bulk: ""},
			{Typ: "bulk", Bulk: ""},
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("NewHsetValue(\"\", \"\", \"\") = %v, want %v", result, expected)
	}
}
func TestNewHsetValue_WithUnicode(t *testing.T) {
	hash := "哈希"
	key := "键"
	value := "值"
	expected := Value{
		Typ: "array",
		Array: []Value{
			{Typ: "bulk", Bulk: "hset"},
			{Typ: "bulk", Bulk: "哈希"},
			{Typ: "bulk", Bulk: "键"},
			{Typ: "bulk", Bulk: "值"},
		},
	}

	result := NewHsetValue(hash, key, value)
	fmt.Println(result)
	if result.Typ != expected.Typ {
		t.Errorf("NewHsetValue() Typ = %v, want %v", result.Typ, expected.Typ)
	}

	if len(result.Array) != len(expected.Array) {
		t.Errorf("NewHsetValue() Array length = %d, want %d", len(result.Array), len(expected.Array))
	}

	for i, v := range result.Array {
		if v.Typ != expected.Array[i].Typ || v.Bulk != expected.Array[i].Bulk {
			t.Errorf("NewHsetValue() Array[%d] = {Typ: %v, Bulk: %v}, want {Typ: %v, Bulk: %v}",
				i, v.Typ, v.Bulk, expected.Array[i].Typ, expected.Array[i].Bulk)
		}
	}
}
func TestNewHsetValue_ReturnsCorrectValueType(t *testing.T) {
	result := NewHsetValue("myhash", "mykey", "myvalue")

	if result.Typ != "array" {
		t.Errorf("NewHsetValue() returned Value with Typ = %s, want 'array'", result.Typ)
	}

	expectedArray := []Value{
		{Typ: "bulk", Bulk: "hset"},
		{Typ: "bulk", Bulk: "myhash"},
		{Typ: "bulk", Bulk: "mykey"},
		{Typ: "bulk", Bulk: "myvalue"},
	}

	if len(result.Array) != len(expectedArray) {
		t.Errorf("NewHsetValue() returned Value with Array length = %d, want %d", len(result.Array), len(expectedArray))
	}

	for i, v := range result.Array {
		if v.Typ != expectedArray[i].Typ || v.Bulk != expectedArray[i].Bulk {
			t.Errorf("NewHsetValue() returned Value with Array[%d] = {Typ: %s, Bulk: %s}, want {Typ: %s, Bulk: %s}",
				i, v.Typ, v.Bulk, expectedArray[i].Typ, expectedArray[i].Bulk)
		}
	}
}

// test NewDelValue
func TestNewDelValue_EmptyKeys(t *testing.T) {
	result := NewDelValue([]string{})
	expected := Value{
		Typ: "array",
		Array: []Value{
			{Typ: "bulk", Bulk: "del"},
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("NewDelValue([]string{}) = %v, want %v", result, expected)
	}

	if len(result.Array) != 1 {
		t.Errorf("NewDelValue([]string{}) returned array of length %d, want 1", len(result.Array))
	}

	if result.Array[0].Typ != "bulk" || result.Array[0].Bulk != "del" {
		t.Errorf("NewDelValue([]string{}) first element = {Typ: %s, Bulk: %s}, want {Typ: bulk, Bulk: del}",
			result.Array[0].Typ, result.Array[0].Bulk)
	}
}
func TestNewDelValue_SingleKey(t *testing.T) {
	keys := []string{"testkey"}
	result := NewDelValue(keys)

	expected := Value{
		Typ: "array",
		Array: []Value{
			{Typ: "bulk", Bulk: "del"},
			{Typ: "bulk", Bulk: "testkey"},
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("NewDelValue(%v) = %v, want %v", keys, result, expected)
	}
}
func TestNewDelValue_MultipleKeys(t *testing.T) {
	keys := []string{"key1", "key2", "key3"}
	result := NewDelValue(keys)

	expected := Value{
		Typ: "array",
		Array: []Value{
			{Typ: "bulk", Bulk: "del"},
			{Typ: "bulk", Bulk: "key1"},
			{Typ: "bulk", Bulk: "key2"},
			{Typ: "bulk", Bulk: "key3"},
		},
	}

	if result.Typ != expected.Typ {
		t.Errorf("NewDelValue() returned incorrect type. Got %s, want %s", result.Typ, expected.Typ)
	}

	if len(result.Array) != len(expected.Array) {
		t.Errorf("NewDelValue() returned array of incorrect length. Got %d, want %d", len(result.Array), len(expected.Array))
	}

	for i, v := range result.Array {
		if v.Typ != expected.Array[i].Typ || v.Bulk != expected.Array[i].Bulk {
			t.Errorf("NewDelValue() returned incorrect value at index %d. Got {Typ: %s, Bulk: %s}, want {Typ: %s, Bulk: %s}",
				i, v.Typ, v.Bulk, expected.Array[i].Typ, expected.Array[i].Bulk)
		}
	}
}
func TestNewDelValue_SpecialCharacters(t *testing.T) {
	keys := []string{"key@1", "key#2", "key$3", "key%4", "key^5"}
	result := NewDelValue(keys)

	expected := Value{
		Typ: "array",
		Array: []Value{
			{Typ: "bulk", Bulk: "del"},
			{Typ: "bulk", Bulk: "key@1"},
			{Typ: "bulk", Bulk: "key#2"},
			{Typ: "bulk", Bulk: "key$3"},
			{Typ: "bulk", Bulk: "key%4"},
			{Typ: "bulk", Bulk: "key^5"},
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("NewDelValue() with special characters = %v, want %v", result, expected)
	}

	if len(result.Array) != len(keys)+1 {
		t.Errorf("NewDelValue() returned array of incorrect length. Got %d, want %d", len(result.Array), len(keys)+1)
	}

	for i, key := range keys {
		if result.Array[i+1].Typ != "bulk" || result.Array[i+1].Bulk != key {
			t.Errorf("NewDelValue() returned incorrect value at index %d. Got {Typ: %s, Bulk: %s}, want {Typ: bulk, Bulk: %s}",
				i+1, result.Array[i+1].Typ, result.Array[i+1].Bulk, key)
		}
	}
}
func TestNewDelValue_VeryLongKeys(t *testing.T) {
	longKey1 := strings.Repeat("a", 1000)
	longKey2 := strings.Repeat("b", 2000)
	keys := []string{longKey1, longKey2}

	result := NewDelValue(keys)

	expected := Value{
		Typ: "array",
		Array: []Value{
			{Typ: "bulk", Bulk: "del"},
			{Typ: "bulk", Bulk: longKey1},
			{Typ: "bulk", Bulk: longKey2},
		},
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("NewDelValue() with very long keys = %v, want %v", result, expected)
	}

	if len(result.Array) != len(keys)+1 {
		t.Errorf("NewDelValue() returned array of incorrect length. Got %d, want %d", len(result.Array), len(keys)+1)
	}

	if len(result.Array[1].Bulk) != 1000 || len(result.Array[2].Bulk) != 2000 {
		t.Errorf("NewDelValue() did not preserve long key lengths. Got %d and %d, want 1000 and 2000",
			len(result.Array[1].Bulk), len(result.Array[2].Bulk))
	}
}
func TestNewDelValue_VariousInputs(t *testing.T) {
	testCases := []struct {
		name     string
		keys     []string
		expected Value
	}{
		{
			name: "Empty input",
			keys: []string{},
			expected: Value{
				Typ:   "array",
				Array: []Value{{Typ: "bulk", Bulk: "del"}},
			},
		},
		{
			name: "Single key",
			keys: []string{"key1"},
			expected: Value{
				Typ: "array",
				Array: []Value{
					{Typ: "bulk", Bulk: "del"},
					{Typ: "bulk", Bulk: "key1"},
				},
			},
		},
		{
			name: "Multiple keys",
			keys: []string{"key1", "key2", "key3"},
			expected: Value{
				Typ: "array",
				Array: []Value{
					{Typ: "bulk", Bulk: "del"},
					{Typ: "bulk", Bulk: "key1"},
					{Typ: "bulk", Bulk: "key2"},
					{Typ: "bulk", Bulk: "key3"},
				},
			},
		},
		{
			name: "Keys with special characters",
			keys: []string{"key@1", "key#2", "key$3"},
			expected: Value{
				Typ: "array",
				Array: []Value{
					{Typ: "bulk", Bulk: "del"},
					{Typ: "bulk", Bulk: "key@1"},
					{Typ: "bulk", Bulk: "key#2"},
					{Typ: "bulk", Bulk: "key$3"},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NewDelValue(tc.keys)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("NewDelValue(%v) = %v, want %v", tc.keys, result, tc.expected)
			}
		})
	}
}
func TestNewDelValue_CompareWithNewdelvalue(t *testing.T) {
	keys := []string{"key1", "key2", "key3"}

	result := NewDelValue(keys)
	expected := newdelvalue(keys)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("NewDelValue(%v) = %v, want %v", keys, result, expected)
	}
}

// Helper function to simulate the hypothetical newdelvalue function
func newdelvalue(keys []string) Value {
	arr := make([]Value, len(keys)+1)
	arr[0] = Value{Typ: "bulk", Bulk: "del"}
	for i, key := range keys {
		arr[i+1] = Value{Typ: "bulk", Bulk: key}
	}
	return Value{Typ: "array", Array: arr}
}
func TestNewDelValue_LargeNumberOfKeys(t *testing.T) {
	// Generate 1000 keys
	keys := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		keys[i] = fmt.Sprintf("key%d", i)
	}

	result := NewDelValue(keys)

	// Check if the result is of type "array"
	if result.Typ != "array" {
		t.Errorf("NewDelValue() returned incorrect type. Got %s, want array", result.Typ)
	}

	// Check if the array has 1001 elements (1000 keys + "del")
	if len(result.Array) != 1001 {
		t.Errorf("NewDelValue() returned array of incorrect length. Got %d, want 1001", len(result.Array))
	}

	// Check if the first element is "del"
	if result.Array[0].Typ != "bulk" || result.Array[0].Bulk != "del" {
		t.Errorf("First element is incorrect. Got {Typ: %s, Bulk: %s}, want {Typ: bulk, Bulk: del}",
			result.Array[0].Typ, result.Array[0].Bulk)
	}

	// Check a sample of keys
	sampleIndices := []int{1, 500, 999}
	for _, i := range sampleIndices {
		if result.Array[i].Typ != "bulk" || result.Array[i].Bulk != fmt.Sprintf("key%d", i-1) {
			t.Errorf("Key at index %d is incorrect. Got {Typ: %s, Bulk: %s}, want {Typ: bulk, Bulk: key%d}",
				i, result.Array[i].Typ, result.Array[i].Bulk, i-1)
		}
	}
}
func TestNewDelValue_WithUnicodeKeys(t *testing.T) {
	keys := []string{"키", "कुंजी", "مفتاح", "ключ", "钥匙"}
	result := NewDelValue(keys)

	expected := Value{
		Typ: "array",
		Array: []Value{
			{Typ: "bulk", Bulk: "del"},
			{Typ: "bulk", Bulk: "키"},
			{Typ: "bulk", Bulk: "कुंजी"},
			{Typ: "bulk", Bulk: "مفتاح"},
			{Typ: "bulk", Bulk: "ключ"},
			{Typ: "bulk", Bulk: "钥匙"},
		},
	}

	if result.Typ != expected.Typ {
		t.Errorf("NewDelValue() returned incorrect type. Got %s, want %s", result.Typ, expected.Typ)
	}

	if len(result.Array) != len(expected.Array) {
		t.Errorf("NewDelValue() returned array of incorrect length. Got %d, want %d", len(result.Array), len(expected.Array))
	}

	for i, v := range result.Array {
		if v.Typ != expected.Array[i].Typ || v.Bulk != expected.Array[i].Bulk {
			t.Errorf("NewDelValue() returned incorrect value at index %d. Got {Typ: %s, Bulk: %s}, want {Typ: %s, Bulk: %s}",
				i, v.Typ, v.Bulk, expected.Array[i].Typ, expected.Array[i].Bulk)
		}
	}
}
func TestNewDelValue_AlwaysReturnsArrayType(t *testing.T) {
	testCases := []struct {
		name string
		keys []string
	}{
		{"Empty keys", []string{}},
		{"Single key", []string{"key1"}},
		{"Multiple keys", []string{"key1", "key2", "key3"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := NewDelValue(tc.keys)
			if result.Typ != "array" {
				t.Errorf("NewDelValue(%v) returned Value with Typ = %s, want 'array'", tc.keys, result.Typ)
			}
		})
	}
}
