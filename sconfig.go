// Package sconfig is a simple yet functional configuration file parser.
//
// See the README.markdown for an introduction.
package sconfig

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"bitbucket.org/pkg/inflect"
)

// TypeHandlers can be used to handle types other than the basic builtin ones;
// it's also possible to override the default types. The key is the name of the
// type.
var TypeHandlers = make(map[string]TypeHandler)

// TypeHandler takes the field to set and the value to set it to. It is expected
// to return the value to set it to.
type TypeHandler func([]string) (interface{}, error)

// Handler functions can be used to run special code for a field. The function
// takes the unprocessed line split by whitespace and with the option name
// removed.
type Handler func([]string) error

// Handlers can be used to run special code for a field. The map key is the name
// of the field in the struct.
type Handlers map[string]Handler

func init() {
	defaultTypeHandlers()
}

func defaultTypeHandlers() {
	TypeHandlers = map[string]TypeHandler{
		// TODO: Parameters after the first are now all just ignored; should be
		// an error
		"string":  handleString,
		"bool":    handleBool,
		"float32": handleFloat32,
		"float64": handleFloat64,
		"int":     handleInt,
		"int8":    handleInt8,
		"int16":   handleInt16,
		"int32":   handleInt32,
		"int64":   handleInt64,
		"uint":    handleUint,
		"uint8":   handleUint8,
		"uint16":  handleUint16,
		"uint32":  handleUint32,
		"uint64":  handleUint64,

		"[]string":  handleStringSlice,
		"[]bool":    handleBoolSlice,
		"[]float32": handleFloat32Slice,
		"[]float64": handleFloat64Slice,

		"[]int":   handleIntSlice,
		"[]int8":  handleInt8Slice,
		"[]int16": handleInt16Slice,
		"[]int32": handleInt32Slice,
		"[]int64": handleInt64Slice,

		"[]uint":   handleUintSlice,
		"[]uint8":  handleUint8Slice,
		"[]uint16": handleUint16Slice,
		"[]uint32": handleUint32Slice,
		"[]uint64": handleUint64Slice,
	}
}

// readFile will read a file, strip comments, and collapse indents. This also
// deals with the special "source" command.
//
// The return value is an array of arrays, where the first item is the original
// line number and the second is the actual line; for example:
//
//     [][]string{
//         []string{3, "key value"},
//         []string{9, "key2 value1 value2"},
//     }
//
// It expects the input file to be utf-8 encoded; other encodings are not
// supported.
func readFile(file string) (lines [][]string, err error) {
	fp, err := os.Open(file)
	if err != nil {
		return lines, err
	}
	defer func() { _ = fp.Close() }()

	i := 0
	no := 0
	for scanner := bufio.NewScanner(fp); scanner.Scan(); {
		no++
		line := scanner.Text()

		isIndented := len(line) > 0 && unicode.IsSpace(rune(line[0]))
		line = strings.TrimSpace(line)
		if line == "" || line[0] == '#' {
			continue
		}

		line = collapseWhitespace(removeComments(line))

		if isIndented {
			if i == 0 {
				return lines, fmt.Errorf("first line can't be indented")
			}
			lines[i-1][1] += " " + line
		} else {
			// Source
			if strings.HasPrefix(line, "source ") {
				sourced, err := readFile(line[7:])
				if err != nil {
					return [][]string{}, err
				}
				lines = append(lines, sourced...)
			} else {
				lines = append(lines, []string{fmt.Sprintf("%d", no), line})
			}
			i++
		}
	}

	return lines, nil
}

func removeComments(line string) string {
	prevcmt := 0
	for {
		cmt := strings.Index(line[prevcmt:], "#")
		if cmt < 0 {
			break
		}

		cmt += prevcmt
		prevcmt = cmt

		// Allow escaping # with \#
		if line[cmt-1] == '\\' {
			line = line[:cmt-1] + line[cmt:]
		} else {
			// Found comment
			line = line[:cmt]
			break
		}
	}

	return line
}

func collapseWhitespace(line string) string {
	nl := ""
	prevSpace := false
	for i, char := range line {
		switch {
		case char == '\\':
			// \ is escaped with \: "\\"
			if line[i-1] == '\\' {
				nl += `\`
			}
		case unicode.IsSpace(char):
			if prevSpace {
				// Escaped with \: "\ "
				if line[i-1] == '\\' {
					nl += string(char)
				}
			} else {
				prevSpace = true
				if i != len(line)-1 {
					nl += " "
				}
			}
		default:
			nl += string(char)
			prevSpace = false
		}
	}

	return nl
}

// MustParse behaves like Parse, but panics if there is an error.
func MustParse(c interface{}, file string, handlers Handlers) {
	err := Parse(c, file, handlers)
	if err != nil {
		panic(err)
	}
}

// Parse will reads file from disk and populates the given config struct c.
//
// The Handlers map can be given to customize the the behaviour; the key is the
// name of the config struct field, and the function is passed a slice with all
// the values on the line.
// There is no return value, the function is epected to set any settings on the
// struct; for example:
//
//     MustParse(&config, "config", Handlers{
//     	"Bool": func(line []string) {
//     		if line[0] == "yup" {
//     			config.Bool = true
//     		}
//     	},
//     })
//
// TODO: Document more
func Parse(c interface{}, file string, handlers Handlers) error {
	lines, err := readFile(file)
	if err != nil {
		return err
	}

	values := reflect.ValueOf(c).Elem()

	// Get list of rule names from tags
	for _, line := range lines {
		// Split by spaces
		v := strings.Split(line[1], " ")

		// Infer the field name from the key
		fieldName, err := fieldNameFromKey(v[0], values)
		if err != nil {
			return fmterr(file, line[0], v[0], err)
		}
		field := values.FieldByName(fieldName)

		// Use the handler if it exists
		if has, err := setFromHandler(fieldName, v[1:], handlers); has {
			if err != nil {
				return fmterr(file, line[0], v[0], err)
			}
			continue
		}

		// Set from typehandler
		if has, err := setFromTypeHandler(&field, v[1:]); has {
			if err != nil {
				return fmterr(file, line[0], v[0], err)
			}
			continue
		}

		// Give up :-(
		return fmterr(file, line[0], v[0], fmt.Errorf(
			"don't know how to set fields of the type %s",
			field.Type().String()))
	}

	return nil
}

func fmterr(file, line, key string, err error) error {
	return fmt.Errorf("%v line %v: error parsing %s: %v",
		file, line, key, err)
}

func fieldNameFromKey(key string, values reflect.Value) (string, error) {
	fieldName := inflect.Camelize(key)

	// TODO: Maybe find better inflect package that deals with this already?
	// This list is from golint
	acr := []string{"Api", "Ascii", "Cpu", "Css", "Dns", "Eof", "Guid", "Html",
		"Https", "Http", "Id", "Ip", "Json", "Lhs", "Qps", "Ram", "Rhs",
		"Rpc", "Sla", "Smtp", "Sql", "Ssh", "Tcp", "Tls", "Ttl", "Udp",
		"Ui", "Uid", "Uuid", "Uri", "Url", "Utf8", "Vm", "Xml", "Xsrf",
		"Xss"}
	for _, a := range acr {
		fieldName = strings.Replace(fieldName, a, strings.ToUpper(a), -1)
	}

	field := values.FieldByName(fieldName)
	if !field.CanAddr() {
		// Check plural version too; we're not too fussy
		fieldNamePlural := inflect.Pluralize(fieldName)
		field = values.FieldByName(fieldNamePlural)
		if !field.CanAddr() {
			return "", fmt.Errorf("unknown option (field %s or %s is missing)",
				fieldName, fieldNamePlural)
		}
		fieldName = fieldNamePlural
	}

	return fieldName, nil
}

func setFromHandler(fieldName string, values []string, handlers Handlers) (bool, error) {
	if handlers == nil {
		return false, nil
	}

	handler, has := handlers[fieldName]
	if !has {
		return false, nil
	}

	err := handler(values)
	if err != nil {
		return true, fmt.Errorf("%v (from handler)", err)
	}

	return true, nil
}

func setFromTypeHandler(field *reflect.Value, value []string) (bool, error) {
	handler, has := TypeHandlers[field.Type().String()]
	if !has {
		return false, nil
	}

	v, err := handler(value)
	if err != nil {
		return true, err
	}
	field.Set(reflect.ValueOf(v))
	return true, nil
}

func parseBool(v string) (bool, error) {
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on", "enable", "enabled":
		return true, nil
	case "0", "false", "no", "off", "disable", "disabled":
		return false, nil
	default:
		return false, fmt.Errorf(`unable to parse "%s" as a boolean`, v)
	}
}

// FindConfig tries to find a config file at the usual locations (in this
// order):
//
//   ~/.config/$file
//   ~/.$file
//   /etc/$file
//   /usr/local/etc/$file
//   /usr/pkg/etc/$file
//   ./$file
func FindConfig(file string) string {
	file = strings.TrimLeft(file, "/")

	locations := []string{}
	if xdg := os.Getenv("XDG_CONFIG"); xdg != "" {
		locations = append(locations, strings.TrimRight(xdg, "/")+"/"+file)
	}
	if home := os.Getenv("HOME"); home != "" {
		locations = append(locations, home+"/."+file)
	}

	locations = append(locations, []string{
		"/etc/" + file,
		"/usr/local/etc/" + file,
		"/usr/pkg/etc/" + file,
		"./" + file,
	}...)

	for _, l := range locations {
		if _, err := os.Stat(l); err == nil {
			return l
		}
	}

	return ""
}

func handleString(v []string) (interface{}, error) {
	return strings.Join(v, " "), nil
}

func handleBool(v []string) (interface{}, error) {
	r, err := parseBool(v[0])
	if err != nil {
		return nil, err
	}
	return r, nil
}

func handleFloat32(v []string) (interface{}, error) {
	r, err := strconv.ParseFloat(v[0], 32)
	if err != nil {
		return nil, err
	}
	return float32(r), nil
}
func handleFloat64(v []string) (interface{}, error) {
	r, err := strconv.ParseFloat(v[0], 64)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func handleInt(v []string) (interface{}, error) {
	// Can be 32 or 64 bits
	if int(2147483647)<<1 > 0 {
		h, _ := TypeHandlers["int64"]
		r, err := h(v)
		if err != nil {
			return nil, err
		}
		return int(r.(int64)), err
	}
	h, _ := TypeHandlers["int32"]
	r, err := h(v)
	if err != nil {
		return nil, err
	}
	return int(r.(int32)), err
}

func handleInt8(v []string) (interface{}, error) {
	r, err := strconv.ParseInt(v[0], 10, 8)
	if err != nil {
		return nil, err
	}
	return int8(r), nil
}

func handleInt16(v []string) (interface{}, error) {
	r, err := strconv.ParseInt(v[0], 10, 16)
	if err != nil {
		return nil, err
	}
	return int16(r), nil
}

func handleInt32(v []string) (interface{}, error) {
	r, err := strconv.ParseInt(v[0], 10, 32)
	if err != nil {
		return nil, err
	}
	return int32(r), nil
}
func handleInt64(v []string) (interface{}, error) {
	r, err := strconv.ParseInt(v[0], 10, 64)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func handleUint(v []string) (interface{}, error) {
	// TODO: This is just too damn ugly
	//defer func() (interface{}, error) {
	//	rec := recover()

	//	if !strings.HasSuffix(rec.(string), "overflows uint32") {
	//		panic(rec)
	//	}

	//	h, _ := TypeHandlers["int32"]
	//	r, err := h(v)
	//	if err != nil {
	//		return nil, err
	//	}
	//	return uint(r.(uint32)), err
	//}()
	//
	// _ = uint(4294967296)

	h, _ := TypeHandlers["uint64"]
	r, err := h(v)
	if err != nil {
		return nil, err
	}
	return uint(r.(uint64)), err

}

func handleUint8(v []string) (interface{}, error) {
	r, err := strconv.ParseUint(v[0], 10, 8)
	if err != nil {
		return nil, err
	}
	return uint8(r), nil
}

func handleUint16(v []string) (interface{}, error) {
	r, err := strconv.ParseUint(v[0], 10, 16)
	if err != nil {
		return nil, err
	}
	return uint16(r), nil
}

func handleUint32(v []string) (interface{}, error) {
	r, err := strconv.ParseUint(v[0], 10, 32)
	if err != nil {
		return nil, err
	}
	return uint32(r), nil
}
func handleUint64(v []string) (interface{}, error) {
	r, err := strconv.ParseUint(v[0], 10, 64)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func handleStringSlice(v []string) (interface{}, error) {
	return v, nil
}

func handleBoolSlice(v []string) (interface{}, error) {
	a := make([]bool, len(v))
	for i := range v {
		r, err := parseBool(v[i])
		if err != nil {
			return nil, err
		}
		a[i] = r
	}
	return a, nil
}

func handleFloat32Slice(v []string) (interface{}, error) {
	a := make([]float32, len(v))
	for i := range v {
		r, err := strconv.ParseFloat(v[i], 32)
		if err != nil {
			return nil, err
		}
		a[i] = float32(r)
	}
	return a, nil
}

func handleFloat64Slice(v []string) (interface{}, error) {
	a := make([]float64, len(v))
	for i := range v {
		r, err := strconv.ParseFloat(v[i], 64)
		if err != nil {
			return nil, err
		}
		a[i] = r
	}
	return a, nil
}

func handleIntSlice(v []string) (interface{}, error) {
	h, _ := TypeHandlers["int"]
	a := make([]int, len(v))
	for i := range v {
		r, err := h([]string{v[i]})
		if err != nil {
			return nil, err
		}
		a[i] = r.(int)
	}
	return a, nil
}

func handleInt8Slice(v []string) (interface{}, error) {
	a := make([]int8, len(v))
	for i := range v {
		r, err := strconv.ParseInt(v[i], 10, 8)
		if err != nil {
			return nil, err
		}
		a[i] = int8(r)
	}
	return a, nil
}

func handleInt16Slice(v []string) (interface{}, error) {
	a := make([]int16, len(v))
	for i := range v {
		r, err := strconv.ParseInt(v[i], 10, 16)
		if err != nil {
			return nil, err
		}
		a[i] = int16(r)
	}
	return a, nil
}

func handleInt32Slice(v []string) (interface{}, error) {
	a := make([]int32, len(v))
	for i := range v {
		r, err := strconv.ParseInt(v[i], 10, 32)
		if err != nil {
			return nil, err
		}
		a[i] = int32(r)
	}
	return a, nil
}

func handleInt64Slice(v []string) (interface{}, error) {
	a := make([]int64, len(v))
	for i := range v {
		r, err := strconv.ParseInt(v[i], 10, 64)
		if err != nil {
			return nil, err
		}
		a[i] = r
	}
	return a, nil
}

func handleUintSlice(v []string) (interface{}, error) {
	h, _ := TypeHandlers["uint"]
	a := make([]uint, len(v))
	for i := range v {
		r, err := h([]string{v[i]})
		if err != nil {
			return nil, err
		}
		a[i] = r.(uint)
	}
	return a, nil
}

func handleUint8Slice(v []string) (interface{}, error) {
	a := make([]uint8, len(v))
	for i := range v {
		r, err := strconv.ParseUint(v[i], 10, 8)
		if err != nil {
			return nil, err
		}
		a[i] = uint8(r)
	}
	return a, nil
}

func handleUint16Slice(v []string) (interface{}, error) {
	a := make([]uint16, len(v))
	for i := range v {
		r, err := strconv.ParseUint(v[i], 10, 16)
		if err != nil {
			return nil, err
		}
		a[i] = uint16(r)
	}
	return a, nil
}

func handleUint32Slice(v []string) (interface{}, error) {
	a := make([]uint32, len(v))
	for i := range v {
		r, err := strconv.ParseUint(v[i], 10, 32)
		if err != nil {
			return nil, err
		}
		a[i] = uint32(r)
	}
	return a, nil
}

func handleUint64Slice(v []string) (interface{}, error) {
	a := make([]uint64, len(v))
	for i := range v {
		r, err := strconv.ParseUint(v[i], 10, 64)
		if err != nil {
			return nil, err
		}
		a[i] = r
	}
	return a, nil
}

// The MIT License (MIT)
//
// Copyright © 2016 Martin Tournoij
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
