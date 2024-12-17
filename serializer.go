package goresp

import (
	"strconv"
	"strings"
)

// serializeCommand converts the given string to RESP formated bytes
func SerializeCommand(command string) []byte {
	parts := strings.Fields(command)

	var test []Value

	for _, part := range parts {

		bulk := Value{Typ: "bulk", Bulk: part}
		test = append(test, bulk)
	}

	newCommand := Value{Typ: "array", Array: test}

	return newCommand.Marshal()
}

// SerializeValue takes resp Value and converts it to string
func SerializeValue(res Value) string {
	//Todo: handle null and error
	switch res.Typ {
	case "bulk":
		{
			return res.Bulk
		}
	case "string":
		{
			return res.Str
		}

	case "int":
		{
			return strconv.Itoa(int(res.Num))
		}

	case "array":
		{
			var fullStr string
			for _, value := range res.Array {
				fullStr += SerializeValue(value)
			}

			return fullStr
		}
	default:
		{
			return res.Bulk
		}
	}
}
