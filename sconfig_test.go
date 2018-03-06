// Copyright © 2016-2017 Martin Tournoij
// See the bottom of this file for the full copyright.

package sconfig

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
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

func rm(t *testing.T, path string) {
	err := os.Remove(path)
	if err != nil {
		t.Errorf("cannot remove %#v: %v", path, err)
	}
}

func rmAll(t *testing.T, path string) {
	err := os.RemoveAll(path)
	if err != nil {
		t.Errorf("cannot remove %#v: %v", path, err)
	}
}

func TestRegisterType(t *testing.T) {
	defer defaultTypeHandlers()
	didint := false
	didint64 := false
	RegisterType("int", func(v []string) (interface{}, error) {
		didint = true
		return int(42), nil
	})
	RegisterType("int64", func(v []string) (interface{}, error) {
		didint64 = true
		return int64(42), nil
	})

	f := testfile("hello 42\nworld 42")
	defer rm(t, f)

	c := &struct {
		Hello int64
		World int
	}{}

	err := Parse(c, f, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !didint {
		t.Error("didint was not true")
	}
	if !didint64 {
		t.Error("didint64 was not true")
	}
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
	defer rm(t, f)
	out, err = readFile(f)
	if err == nil {
		t.Error("no error on sourcing /nonexistent-file")
	}
	if len(out) > 0 {
		t.Error("len(out) > 0")
	}

	// First line is indented: makes no sense.
	f2 := testfile(" indented")
	defer rm(t, f2)
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
	defer rm(t, source)

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
	defer rm(t, f)
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
	find := FindConfig("sure_this_wont_exist/anywhere")
	if find != "" {
		t.Fail()
	}

	dir, err := ioutil.TempDir(os.TempDir(), "sconfig_test")
	if err != nil {
		t.Error(err)
	}
	defer rmAll(t, dir)

	f, err := ioutil.TempFile(dir, "config")
	if err != nil {
		t.Fatal(err)
	}

	err = os.Setenv("XDG_CONFIG", dir)
	if err != nil {
		t.Fatal(err)
	}
	find = FindConfig(filepath.Base(f.Name()))
	if find != f.Name() {
		t.Fail()
	}

	//t.Fail()
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
	defer rm(t, f)
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
	defer rm(t, f2)
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

// Make sure we give a sane error
func TestGetValues(t *testing.T) {
	out := struct {
		Foo string
	}{}

	f := testfile(`foo bar`)
	defer rm(t, f)

	err := Parse(out, f, nil)
	if err == nil {
		t.Fatal("Err is nil")
	}
	switch err.(type) {
	case *reflect.ValueError:
		t.Fatal("still reflect.ValueError")
	}
}

func TestParsePrimitives(t *testing.T) {
	test := `
str foo bar
int64 46
uint64 51
bool yes
bool2 true
bool3
bool4 no
float32 3.14
float64 3.14159
`
	expected := testPrimitives{
		Str:     "foo bar",
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
	defer rm(t, f)
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

		`int64 1 2`: `line 1: error parsing int64: must have exactly one value`,
		`uint64`:    `line 1: error parsing uint64: must have exactly one value`,
	}

	for test, expected := range tests {
		f := testfile(test)
		defer rm(t, f)

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
	defer rm(t, f)
	err := Parse(&out, f, nil)
	if err != nil {
		t.Error(err.Error())
	}
	if out.Str != "default value" {
		t.Error()
	}

	f2 := testfile("str changed\n")
	defer rm(t, f2)
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
	defer rm(t, f)

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
			return errors.New("oh noes")
		},
	})
	if err == nil {
		t.Error("error is nil")
	}
	expected := " line 2: error parsing Int64: oh noes (from handler)"
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
str append this
int64 46 700
uint64 51 705
bool yes no yes
float32 3.14 1.1
float64 3.14159 1.2
`

	expected := testArray{
		Str:     []string{"foo", "bar", "append", "this"},
		Int64:   []int64{46, 700},
		UInt64:  []uint64{51, 705},
		Bool:    []bool{true, false, true},
		Float32: []float32{3.14, 1.1},
		Float64: []float64{3.14159, 1.2},
	}

	out := testArray{}
	f := testfile(test)
	defer rm(t, f)
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

		"int64": `line 1: error parsing int64: must have more than 1 values (has: 0)`,
	}

	for test, expected := range tests {
		f := testfile(test)
		defer rm(t, f)

		out := testArray{}
		err := Parse(&out, f, nil)
		if err == nil {
			t.Errorf("got to have an error for %v", test)
			t.FailNow()
		}
		if !strings.HasSuffix(err.Error(), expected) {
			t.Errorf("\nexpected:  %#v\nout:       %#v\n", expected, err.Error())
		}
	}

}

func TestInflect(t *testing.T) {
	c := &struct {
		Key    []string
		Planes []string
	}{}

	f := testfile("key a\nplanes b\nkeys a\nplane b")
	defer rm(t, f)

	err := Parse(c, f, nil)
	if err != nil {
		t.Fatal(err)
	}
}

// Make sure it doesn't panic.
func TestWeirdType(t *testing.T) {
	f := testfile("foo.bar a\nasd.zxc 42\n")
	defer rm(t, f)

	c := "foo"
	err := Parse(&c, f, nil)
	if err == nil {
		t.Fatal("no err?!")
	}
}

func TestMapString(t *testing.T) {
	f := testfile("foo.bar a\nasd.zxc 42\n")
	defer rm(t, f)

	c := map[string][]string{}
	err := Parse(&c, f, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(c["foo.bar"], []string{"a"}) {
		t.Errorf("wrong output: %#v", c["foo.bar"])
	}
}

func TestX(t *testing.T) {
	f := testfile("hello one two three\nhello foo bar")
	defer rm(t, f)

	c := struct {
		Hello []string
	}{}
	err := Parse(&c, f, Handlers{
		"Hello": func(line []string) error {
			fmt.Printf("%#v\n", line)
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Printf("%#v\n", c)
}

func TestFields(t *testing.T) {
	c := testPrimitives{Str: "init"}
	names := Fields(&c)

	v, ok := names["Str"]
	if !ok {
		t.Fatalf("Str not in map")
	}
	if v.Interface().(string) != "init" {
		t.Fatalf("Str wrong value")
	}

	v.SetString("XXX")
	if v.Interface().(string) != "XXX" {
		t.Fatalf("Str wrong value")
	}
}

type Marsh struct{ v string }

func (m *Marsh) UnmarshalText(text []byte) error {
	m.v = string(text)
	if m.v == "error" {
		return errors.New("error")
	}
	return nil
}

func TestTextUnmarshaler(t *testing.T) {
	c := struct{ Field *Marsh }{}

	t.Run("set value", func(t *testing.T) {
		f := testfile("field !! ??")
		defer rm(t, f)

		err := Parse(&c, f, nil)
		if err != nil {
			t.Fatal("error", err)
		}
		if c.Field.v != "!! ??" {
			t.Errorf("value wrong: %#v", c.Field.v)
		}
	})

	t.Run("error", func(t *testing.T) {
		f := testfile("field error")
		defer rm(t, f)

		err := Parse(&c, f, nil)
		if err == nil {
			t.Fatal("error is nil")
		}
		if !strings.Contains(err.Error(), "line 1: error parsing field: error") {
			t.Errorf("wrong error: %#v", err.Error())
		}
	})
}

// The MIT License (MIT)
//
// Copyright © 2016-2017 Martin Tournoij
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// The software is provided "as is", without warranty of any kind, express or
// implied, including but not limited to the warranties of merchantability,
// fitness for a particular purpose and noninfringement. In no event shall the
// authors or copyright holders be liable for any claim, damages or other
// liability, whether in an action of contract, tort or otherwise, arising
// from, out of or in connection with the software or the use or other dealings
// in the software.
