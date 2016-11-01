package sconfig

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

// testfile will write data to a temporary file and will return the full file
// path. It is the caller's responsibility to clean the file.
func testfile(data string) (filename string) {
	fp, err := ioutil.TempFile(os.TempDir(), "sconfigtest")
	if err != nil {
		panic(err)
	}
	defer func() { _ = fp.Close() }()

	_, err = fp.WriteString(data)
	if err != nil {
		panic(err)
	}
	return fp.Name()
}

func TestReadFileError(t *testing.T) {
	// File doesn't exist
	out, err := readFile("/nonexistent-file")
	if err == nil {
		t.Error("no error on reading /nonexistent-file")
	}
	if len(out) > 0 {
		t.Fail()
	}

	// Sourced file doesn't exist
	f := testfile("source /nonexistent-file")
	defer os.Remove(f)
	out, err = readFile(f)
	if err == nil {
		t.Error("no error on sourcing /nonexistent-file")
	}
	if len(out) > 0 {
		t.Error("len(out) > 0")
	}

	// First line is indented: makes no sense.
	f2 := testfile(" indented")
	defer os.Remove(f2)
	out, err = readFile(f2)
	if err == nil {
		t.Error("no error when first line is indented")
	}
	if len(out) > 0 {
		t.Error("len(out) > 0")
	}
}

func TestReadFile(t *testing.T) {
	source := testfile("sourced file")
	defer os.Remove(source)

	test := fmt.Sprintf(`
# A comment
key value # Ignore this too

key
 value1 # Also ignored
			value2

another−€¡ Hé€ Well...

collapse     many			   whitespaces  

ig\#nore comments \# like this

uni-code    　 white    　 space   　
pre_serve \ spaces \ \ like \		\	so

back s\\la\sh


source %v

`, source)

	expected := [][]string{
		{"3", "key value"},
		{"5", "key value1 value2"},
		{"9", "another−€¡ Hé€ Well..."},
		{"11", "collapse many whitespaces"},
		{"13", "ig#nore comments # like this"},
		{"15", "uni-code white space"},
		{"16", "pre_serve  spaces   like 		so"},
		{"18", `back s\lash`},
		{"1", "sourced file"},
	}

	f := testfile(test)
	defer os.Remove(f)
	out, err := readFile(f)
	if err != nil {
		t.Errorf("readFile: got err: %v", err)
	}

	if len(out) != len(expected) {
		t.Logf("len(out) != len(expected)\nout: %#v", out)
		t.FailNow()
	}

	for i := range expected {
		if out[i][0] != expected[i][0] || out[i][1] != expected[i][1] {
			t.Errorf("%v failed\nexpected:  %#v\nout:       %#v\n",
				i, expected[i], out[i])
		}
	}
}

func TestFindConfigErrors(t *testing.T) {
	f := FindConfig("hieperdepiephoera")
	if f != "" {
		t.Fail()
	}
}

func TestFindConfig(t *testing.T) {
	// TODO: Test this
}

type testPrimitives struct {
	Str     string
	Int64   int64
	UInt64  uint64
	Bool    bool
	Bool2   bool
	Bool3   bool
	Bool4   bool
	Float32 float32
	Float64 float64

	TimeType time.Time
}

func TestMustParse(t *testing.T) {
	out := testPrimitives{}
	f := testfile("str okay")
	defer os.Remove(f)
	MustParse(&out, f, nil)

	defer func() {
		err := recover()
		if err == nil {
			t.Errorf("expected panic")
		}

		expected := " line 1: error parsing not: unknown option (field Not or Nots is missing)"
		if !strings.HasSuffix(err.(error).Error(), expected) {
			t.Errorf("\nexpected:  %#v\nout:       %#v\n", expected, err.(error).Error())
		}
	}()

	f2 := testfile("not okay")
	defer os.Remove(f2)
	MustParse(&out, f2, nil)
}

func TestParseError(t *testing.T) {
	out := testPrimitives{}
	err := Parse(&out, "/nonexistent-file", nil)
	if err == nil {
		t.Error("no error when parsing /nonexistent-file")
	}
	e := testPrimitives{}
	if out != e {
		t.Error("out isn't empty")
	}
}

func TestParsePrimitives(t *testing.T) {
	test := `
str foo
int64 46
uint64 51
bool yes
bool2 true
bool3 1
bool4 no
float32 3.14
float64 3.14159
`
	expected := testPrimitives{
		Str:     "foo",
		Int64:   46,
		UInt64:  51,
		Bool:    true,
		Bool2:   true,
		Bool3:   true,
		Bool4:   false,
		Float32: 3.14,
		Float64: 3.14159,
	}

	out := testPrimitives{}
	f := testfile(test)
	defer os.Remove(f)
	err := Parse(&out, f, nil)
	if err != nil {
		t.Error(err.Error())
	}
	if out != expected {
		t.Errorf("\nexpected:  %#v\nout:       %#v\n", expected, out)
	}
}

func TestInvalidPrimitives(t *testing.T) {
	tests := map[string]string{
		"\n\nInt64 false":            `line 3: error parsing Int64: strconv.ParseInt: parsing "false": invalid syntax`,
		"Bool what?":                 `line 1: error parsing Bool: unable to parse "what?" as a boolean`,
		"woot field":                 `line 1: error parsing woot: unknown option (field Woot or Woots is missing)`,
		"\n\n\n\ntime-type 2016\n\n": `line 5: error parsing time-type: don't know how to set fields of the type time.Time`,

		"float32 42,42": `invalid syntax`,
		"float64 42,42": `invalid syntax`,

		"int64 nope":  `invalid syntax`,
		"uint64 nope": `invalid syntax`,

		//"int 32 64": `TODO`,
	}

	for test, expected := range tests {
		f := testfile(test)
		defer os.Remove(f)

		out := testPrimitives{}
		err := Parse(&out, f, nil)
		if err == nil {
			t.Error("got to have an error")
			t.FailNow()
		}
		if !strings.HasSuffix(err.Error(), expected) {
			t.Errorf("\nexpected:  %#v\nout:       %#v\n", expected, err.Error())
		}
	}

}

func TestDefaults(t *testing.T) {
	out := testPrimitives{
		Str: "default value",
	}
	f := testfile("bool on\n")
	defer os.Remove(f)
	err := Parse(&out, f, nil)
	if err != nil {
		t.Error(err.Error())
	}
	if out.Str != "default value" {
		t.Error()
	}

	f2 := testfile("str changed\n")
	defer os.Remove(f2)
	err = Parse(&out, f2, nil)
	if err != nil {
		t.Error(err.Error())
	}
	if out.Str != "changed" {
		t.Error()
	}
}

func TestParseHandlers(t *testing.T) {
	out := testPrimitives{}
	f := testfile("bool false\nInt64 42\n")
	defer os.Remove(f)

	err := Parse(&out, f, Handlers{
		"Bool": func(line []string) (err error) {
			if line[0] == "false" {
				out.Bool = true
			}
			return
		},
	})
	if err != nil {
		t.Error(err.Error())
	}
	if !out.Bool {
		t.Error()
	}

	err = Parse(&out, f, Handlers{
		"Int64": func(line []string) (err error) {
			return errors.New("Oh noes!")
		},
	})
	if err == nil {
		t.Error("error is nil")
	}
	expected := " line 2: error parsing Int64: Oh noes! (from handler)"
	if !strings.HasSuffix(err.Error(), expected) {
		t.Errorf("\nexpected:  %#v\nout:       %#v\n", expected, err.Error())
	}
}

type testArray struct {
	Str      []string
	Int64    []int64
	UInt64   []uint64
	Bool     []bool
	Float32  []float32
	Float64  []float64
	TimeType []time.Time
}

func TestParseArray(t *testing.T) {
	test := `
str foo bar
int64 46 700
uint64 51 705
bool yes no yes
float32 3.14 1.1
float64 3.14159 1.2
`

	expected := testArray{
		Str:     []string{"foo", "bar"},
		Int64:   []int64{46, 700},
		UInt64:  []uint64{51, 705},
		Bool:    []bool{true, false, true},
		Float32: []float32{3.14, 1.1},
		Float64: []float64{3.14159, 1.2},
	}

	out := testArray{}
	f := testfile(test)
	defer os.Remove(f)
	err := Parse(&out, f, nil)
	if err != nil {
		t.Error(err.Error())
	}
	if fmt.Sprintf("%#v", out) != fmt.Sprintf("%#v", expected) {
		t.Errorf("\nexpected:  %#v\nout:       %#v\n", expected, out)
	}
}

func TestInvalidArray(t *testing.T) {
	tests := map[string]string{
		"\n\nInt64 false":            `line 3: error parsing Int64: strconv.ParseInt: parsing "false": invalid syntax`,
		"Bool what?":                 `line 1: error parsing Bool: unable to parse "what?" as a boolean`,
		"woot field":                 `line 1: error parsing woot: unknown option (field Woot or Woots is missing)`,
		"\n\n\n\ntime-type 2016\n\n": `line 5: error parsing time-type: don't know how to set fields of the type []time.Time`,

		"float32 42,42": `invalid syntax`,
		"float64 42,42": `invalid syntax`,

		"int64 nope":  `invalid syntax`,
		"uint64 nope": `invalid syntax`,

		//"int 32 64": `TODO`,
	}

	for test, expected := range tests {
		f := testfile(test)
		defer os.Remove(f)

		out := testArray{}
		err := Parse(&out, f, nil)
		if err == nil {
			t.Error("got to have an error")
			t.FailNow()
		}
		if !strings.HasSuffix(err.Error(), expected) {
			t.Errorf("\nexpected:  %#v\nout:       %#v\n", expected, err.Error())
		}
	}

}

type testTypeHandlers struct {
	Str  string
	Reg  *regexp.Regexp
	Regs []*regexp.Regexp
}

func TestParseTypeHandlers(t *testing.T) {
	defer defaultTypeHandlers()

	TypeHandlers["string"] = func(v []string) (interface{}, error) {
		return "type handler", nil
	}
	TypeHandlers["*regexp.Regexp"] = func(v []string) (interface{}, error) {
		return regexp.MustCompile(v[0]), nil
	}
	TypeHandlers["[]*regexp.Regexp"] = func(v []string) (interface{}, error) {
		r := []*regexp.Regexp{}
		for _, s := range v {
			r = append(r, regexp.MustCompile(s))
		}
		return r, nil
	}

	test := `
str override this

reg foo.*

regs bar.* [hH]
	`

	out := testTypeHandlers{}
	f := testfile(test)
	defer os.Remove(f)
	err := Parse(&out, f, nil)
	if err != nil {
		t.Error(err.Error())
	}

	if out.Str != "type handler" {
		t.Error()
	}
	if out.Reg.String() != "foo.*" {
		t.Error()
	}
	if len(out.Regs) < 2 {
		t.Error()
	}
	if out.Regs[0].String() != "bar.*" {
		t.Error()
	}
	if out.Regs[1].String() != "[hH]" {
		t.Error()
	}

	// Just in case...
	delete(TypeHandlers, "string")
	err = Parse(&out, f, nil)
	if err == nil {
		t.Error("expected an error")
		t.FailNow()
	}
	if !strings.HasSuffix(err.Error(), "don't know how to set fields of the type string") {
		t.Error("wrong error")
	}

	f2 := testfile("# Hello\n")
	defer os.Remove(f2)

	TypeHandlers = nil
	err = Parse(&out, f2, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestExample(t *testing.T) {
	test := `# This is a comment

port 8080 # This is also a comment

# Look ma, no quotes!
base-url http://example.com

# We'll parse these in a []*regexp.Regexp
match ^foo.+
match ^b[ao]r

# Two values
order allow deny

host  # Multiline stuff
	arp242.net         # My website
	stackoverflow.com  # I like this too
`

	type Config struct {
		Port    int64
		BaseURL string
		Match   []*regexp.Regexp
		Order   []string
		Hosts   []string
	}

	config := Config{}
	TypeHandlers["[]*regexp.Regexp"] = func(v []string) (interface{}, error) {
		a := make([]*regexp.Regexp, len(v))
		for i := range v {
			r, err := regexp.Compile(v[i])
			if err != nil {
				return nil, err
			}
			a[i] = r
		}
		return a, nil
	}

	f := testfile(test)
	defer os.Remove(f)
	err := Parse(&config, f, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config: %v", err)
		t.Error(err.Error())
	}

	//if err == nil {
	//	fmt.Printf("%#v\n", config)
	//	t.Fail()
	//}
}
