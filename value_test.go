package goresp

import (
	"bytes"
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
