package grohl

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

var timeExample = time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC)
var errExample = fmt.Errorf("error message")

type ExampleStruct struct {
	Value interface{}
}

var actuals = []Data{
	{"fn": "string", "test": "hi"},
	{"fn": "stringspace", "test": "a b"},
	{"fn": "stringline", "test": "a b\nc"},
	{"fn": "stringslasher", "test": `slasher \\`},
	{"fn": "stringeqspace", "test": "x=4, y=10"},
	{"fn": "stringeq", "test": "x=4,y=10"},
	{"fn": "stringspace", "test": "hello world"},
	{"fn": "stringbothquotes", "test": `echo 'hello' "world"`},
	{"fn": "stringsinglequotes", "test": `a 'a'`},
	{"fn": "stringdoublequotes", "test": `echo "hello"`},
	{"fn": "stringbothquotesnospace", "test": `'a"`},
	{"fn": "emptystring", "test": ""},
	{"fn": "int", "test": int(1)},
	{"fn": "int8", "test": int8(1)},
	{"fn": "int16", "test": int16(1)},
	{"fn": "int32", "test": int32(1)},
	{"fn": "int64", "test": int64(1)},
	{"fn": "uint", "test": uint(1)},
	{"fn": "uint8", "test": uint8(1)},
	{"fn": "uint16", "test": uint16(1)},
	{"fn": "uint32", "test": uint32(1)},
	{"fn": "uint64", "test": uint64(1)},
	{"fn": "float", "test": float32(1.0)},
	{"fn": "bool", "test": true},
	{"fn": "nil", "test": nil},
	{"fn": "time", "test": timeExample},
	{"fn": "error", "test": errExample},
	{"fn": "slice", "test": []byte{86, 87, 88}},
	{"fn": "struct", "test": ExampleStruct{Value: "testing123"}},
}

var expectations = [][]string{
	{"fn=string", "test=hi"},
	{"fn=stringspace", `test="a b"`},
	{"fn=stringline", `test="a b|c"`},
	{`fn=stringslasher`, `test="slasher \\\\"`},
	{`fn=stringeqspace`, `test="x=4, y=10"`},
	{`fn=stringeq`, `test="x=4,y=10"`},
	{`fn=stringspace`, `test="hello world"`},
	{`fn=stringbothquotes`, `test="echo 'hello' \"world\""`},
	{`fn=stringsinglequotes`, `test="a 'a'"`},
	{`fn=stringdoublequotes`, `test='echo "hello"'`},
	{`fn=stringbothquotesnospace`, `test='a"`},
	{"fn=emptystring", "test=nil"},
	{"fn=int", "test=1"},
	{"fn=int8", "test=1"},
	{"fn=int16", "test=1"},
	{"fn=int32", "test=1"},
	{"fn=int64", "test=1"},
	{"fn=uint", "test=1"},
	{"fn=uint8", "test=1"},
	{"fn=uint16", "test=1"},
	{"fn=uint32", "test=1"},
	{"fn=uint64", "test=1"},
	{"fn=float", "test=1.000"},
	{"fn=bool", "test=true"},
	{"fn=nil", "test=nil"},
	{"fn=time", "test=2000-01-02T03:04:05+0000"},
	{`fn=error`, `test="error message"`},
	{`fn=slice`, `test="[86 87 88]"`},
	{`fn=struct`, `test={Value:testing123}`},
}

func TestFormat(t *testing.T) {
	for i, actual := range actuals {
		AssertData(t, actual, expectations[i]...)
	}
}

func TestFormatWithTime(t *testing.T) {
	data := Data{"fn": "time", "test": 1}
	m := make(map[string]bool)
	parts := BuildLogParts(data, true)
	for _, pair := range parts {
		m[pair] = true
	}
	line := builtLogLine{m, strings.Join(parts, space)}

	if !strings.HasPrefix(line.full, "now=") {
		t.Errorf("Invalid prefix: %s", line.full)
	}

	AssertBuiltLine(t, line, "fn=time", "test=1", "~now=")
}

func AssertLog(t *testing.T, ctx *Context, expected ...string) {
	AssertData(t, ctx.Merge(nil), expected...)
}

func AssertData(t *testing.T, data Data, expected ...string) {
	AssertBuiltLine(t, buildLogLine(data), expected...)
}

func AssertBuiltLine(t *testing.T, line builtLogLine, expected ...string) {
	for _, pair := range expected {
		if strings.HasPrefix(pair, "~") {
			pair = pair[1:]
			found := false
			for actual := range line.pairs {
				if !found {
					found = strings.HasPrefix(actual, pair)
				}
			}

			if !found {
				t.Errorf("Expected partial pair ~ '%s' in %s", pair, line.full)
			}
		} else {
			if _, ok := line.pairs[pair]; !ok {
				t.Errorf("Expected pair '%s' in %s", pair, line.full)
			}
		}
	}

	if expectedLen := len(expected); expectedLen != len(line.pairs) {
		t.Errorf("Expected %d pairs in %s", expectedLen, line.full)
	}
}

func AssertString(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Errorf("Expected %s\nGot: %s", expected, actual)
	}
}

type builtLogLine struct {
	pairs map[string]bool
	full  string
}

func buildLogLine(d Data) builtLogLine {
	m := make(map[string]bool)
	parts := BuildLogParts(d, false)
	for _, pair := range parts {
		m[pair] = true
	}
	return builtLogLine{m, strings.Join(parts, space)}
}
