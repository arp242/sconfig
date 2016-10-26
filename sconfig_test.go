package sconfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

// testfile will write data to a temporary file and will return the full file
// path. It is the caller's responsibility to clean the file.
func testfile(t *testing.T, data string) (filename string) {
	fp, err := ioutil.TempFile(os.TempDir(), "sconfigtest")
	if err != nil {
		t.Fail()
	}
	defer func() { _ = fp.Close() }()

	_, err = fp.WriteString(data)
	if err != nil {
		t.Fail()
	}
	return fp.Name()
}

func TestReadFileError(t *testing.T) {
	// File doesn't exist
	out, err := readFile("/nonexistent")
	if err == nil {
		t.Fail()
	}
	if len(out) > 0 {
		t.Fail()
	}

	// Sourced file doesn't exist
	out, err = readFile(testfile(t, "source /nonexistent"))
	if err == nil {
		t.Fail()
	}
	if len(out) > 0 {
		t.Fail()
	}

	// First line is indented: makes no sense.
	out, err = readFile(testfile(t, " indented"))
	if err == nil {
		t.Fail()
	}
	if len(out) > 0 {
		t.Fail()
	}
}

func TestReadFile(t *testing.T) {
	source := testfile(t, "sourced file")

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
		[]string{"3", "key value"},
		[]string{"5", "key value1 value2"},
		[]string{"9", "another−€¡ Hé€ Well..."},
		[]string{"11", "collapse many whitespaces"},
		[]string{"13", "ig#nore comments # like this"},
		[]string{"15", "uni-code white space"},
		[]string{"16", "pre_serve  spaces   like 		so"},
		[]string{"18", `back s\lash`},
		[]string{"1", "sourced file"},
	}

	out, err := readFile(testfile(t, test))
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
	Int     int
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	UInt    int
	UInt8   int8
	UInt16  int16
	UInt32  int32
	UInt64  int64
	Bool    bool
	Bool2   bool
	Bool3   bool
	Bool4   bool
	Float32 float32
	Float64 float32
}

func TestMustParse(t *testing.T) {
	out := testPrimitives{}
	MustParse(&out, testfile(t, "str okay"), nil)

	defer func() {
		err := recover()
		if err == nil {
			t.Errorf("expected panic")
		}

		if !strings.Contains(err.(error).Error(), "unknown option not") {
			t.Errorf("panic has unexpected message")
		}
	}()
	MustParse(&out, testfile(t, "not okay"), nil)
}

func TestParseError(t *testing.T) {
	out := testPrimitives{}
	err := Parse(&out, "/nonexistent", nil)
	if err == nil {
		t.Fail()
	}
	e := testPrimitives{}
	if out != e {
		t.Fail()
	}
}

func TestParsePrimitives(t *testing.T) {
	test := `
str foo

int 42
int8 43
int16 44
int32 45
int64 46

uint 47
uint8 48
uint16 49
uint32 50
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
		Int:     42,
		Int8:    43,
		Int16:   44,
		Int32:   45,
		Int64:   46,
		UInt:    47,
		UInt8:   48,
		UInt16:  49,
		UInt32:  50,
		UInt64:  51,
		Bool:    true,
		Bool2:   true,
		Bool3:   true,
		Bool4:   false,
		Float32: 3.14,
		Float64: 3.14159,
	}

	out := testPrimitives{}
	err := Parse(&out, testfile(t, test), nil)
	if err != nil {
		t.Error(err.Error())
	}
	if out != expected {
		t.Errorf("\nexpected:  %#v\nout:       %#v\n", expected, out)
	}
}

func TestDefaults(t *testing.T) {
	out := testPrimitives{
		Str: "default value",
	}
	err := Parse(&out, testfile(t, "bool on\n"), nil)
	if err != nil {
		t.Error(err.Error())
	}
	if out.Str != "default value" {
		t.Error()
	}

	err = Parse(&out, testfile(t, "str changed\n"), nil)
	if err != nil {
		t.Error(err.Error())
	}
	if out.Str != "changed" {
		t.Error()
	}
}

func TestParseHandlers(t *testing.T) {
	out := testPrimitives{}
	err := Parse(&out, testfile(t, "bool yup\n"), Handlers{
		"Bool": func(line []string) {
			if line[0] == "yup" {
				out.Bool = true
			}
		},
	})
	if err != nil {
		t.Error(err.Error())
	}

	if !out.Bool {
		t.Error()
	}
}

type testArray struct {
	Str     []string
	Int     []int
	Int8    []int8
	Int16   []int16
	Int32   []int32
	Int64   []int64
	UInt    []uint
	UInt8   []uint8
	UInt16  []uint16
	UInt32  []uint32
	UInt64  []uint64
	Bool    []bool
	Float32 []float32
	Float64 []float64
}

func TestParseArray(t *testing.T) {
	test := `
str foo bar

int 42 666
int8 43 100
int16 44 668
int32 45 669
int64 46 700

uint 47 701
uint8 48 101
uint16 49 703
uint32 50 704
uint64 51 705

bool yes no yes

float32 3.14 1.1
float64 3.14159 1.2
`

	expected := testArray{
		Str:     []string{"foo", "bar"},
		Int:     []int{42, 666},
		Int8:    []int8{43, 100},
		Int16:   []int16{44, 668},
		Int32:   []int32{45, 669},
		Int64:   []int64{46, 700},
		UInt:    []uint{47, 701},
		UInt8:   []uint8{48, 101},
		UInt16:  []uint16{49, 703},
		UInt32:  []uint32{50, 704},
		UInt64:  []uint64{51, 705},
		Bool:    []bool{true, false, true},
		Float32: []float32{3.14, 1.1},
		Float64: []float64{3.14159, 1.2},
	}

	out := testArray{}
	err := Parse(&out, testfile(t, test), nil)
	if err != nil {
		t.Error(err.Error())
	}
	if fmt.Sprintf("%#v", out) != fmt.Sprintf("%#v", expected) {
		t.Errorf("\nexpected:  %#v\nout:       %#v\n", expected, out)
	}
}

type testTypeHandlers struct {
	Str  string
	Reg  *regexp.Regexp
	Regs []*regexp.Regexp
}

func TestParseTypeHandlers(t *testing.T) {
	TypeHandlers["string"] = func(field *reflect.Value, v []string) interface{} {
		return "type handler"
	}
	TypeHandlers["*regexp.Regexp"] = func(field *reflect.Value, v []string) interface{} {
		return regexp.MustCompile(v[0])
	}
	TypeHandlers["[]*regexp.Regexp"] = func(field *reflect.Value, v []string) interface{} {
		r := []*regexp.Regexp{}
		for _, s := range v {
			r = append(r, regexp.MustCompile(s))
		}
		return r
	}

	test := `
str override this

reg foo.*

regs bar.* [hH]
	`

	out := testTypeHandlers{}
	err := Parse(&out, testfile(t, test), nil)
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
		Port    int
		BaseURL string
		Match   []*regexp.Regexp
		Order   []string
		Hosts   []string
	}

	config := Config{}
	TypeHandlers["[]*regexp.Regexp"] = func(field *reflect.Value, v []string) interface{} {
		r := []*regexp.Regexp{}
		for _, s := range v {
			r = append(r, regexp.MustCompile(s))
		}
		return r
	}

	err := Parse(&config, testfile(t, test), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing config: %v", err)
		t.Error(err.Error())
	}

	//if err == nil {
	//	fmt.Printf("%#v\n", config)
	//	t.Fail()
	//}
}
