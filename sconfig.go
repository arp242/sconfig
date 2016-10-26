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
type TypeHandler func(*reflect.Value, []string) interface{}

// Handler functions can be used to run special code for a field. The function
// takes the unprocessed line split by whitespace and with the option name
// removed.
type Handler func(line []string)

// Handlers can be used to run special code for a field. The map key is the name
// of the field in the struct.
type Handlers map[string]Handler

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
		// Record source file and line number for errors
		src := file + ":" + line[0]

		// Split by spaces
		v := strings.Split(line[1], " ")

		// Infer the field name from the key
		fieldName, err := fieldNameFromKey(v[0], src, values)
		if err != nil {
			return err
		}
		field := values.FieldByName(fieldName)

		// Use the handler that was given
		if handlers != nil {
			handler, has := handlers[fieldName]
			if has {
				handler(v[1:])
				continue
			}
		}

		err = setValue(&field, v[1:])
		if err != nil {
			return err
		}
	}

	return nil
}

func fieldNameFromKey(key, src string, values reflect.Value) (string, error) {
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
			return "", fmt.Errorf("%s: unknown option %s (field %s or %s is missing)",
				src, key, fieldName, fieldNamePlural)
		}
		fieldName = fieldNamePlural
	}

	return fieldName, nil
}

// setValue sets the struct field to the given value. The value will be
// type-coerced in the field's type.
func setValue(field *reflect.Value, value []string) error {
	// Try to get it from TypeHandlers first so we can override primitives
	// if we want to.
	if fun, has := TypeHandlers[field.Type().String()]; has {
		v := fun(field, value)
		field.Set(reflect.ValueOf(v))
		return nil
	}

	iface := field.Interface()
	switch iface.(type) {
	// Primitives
	case int, int8, int16, int32, int64:
		bs := 32
		switch iface.(type) {
		case int8:
			bs = 8
		case int16:
			bs = 16
		case int64:
			bs = 64
		}
		i, err := strconv.ParseInt(strings.Join(value, " "), 10, bs)
		if err != nil {
			return err
		}
		field.SetInt(i)
	case uint, uint8, uint16, uint32, uint64:
		bs := 32
		switch iface.(type) {
		case uint8:
			bs = 8
		case uint16:
			bs = 16
		case uint64:
			bs = 64
		}
		i, err := strconv.ParseUint(strings.Join(value, " "), 10, bs)
		if err != nil {
			return err
		}
		field.SetUint(i)
	case bool:
		b, err := parseBool(strings.Join(value, " "))
		if err != nil {
			return err
		}
		field.SetBool(b)
	case float32, float64:
		bs := 32
		switch iface.(type) {
		case float64:
			bs = 64
		}
		i, err := strconv.ParseFloat(strings.Join(value, " "), bs)
		if err != nil {
			return err
		}
		field.SetFloat(i)
	case string:
		field.SetString(strings.Join(value, " "))

	// Arrays of primitives
	// TODO: this code is a bit more verbose than I'd like...
	case []int, []int8, []int16, []int32, []int64:
		for _, v := range value {
			switch iface.(type) {
			case []int:
				i, err := strconv.ParseInt(v, 10, 32)
				if err != nil {
					return err
				}
				field.Set(reflect.Append(*field, reflect.ValueOf(int(i))))
			case []int8:
				i, err := strconv.ParseInt(v, 10, 8)
				if err != nil {
					return err
				}
				field.Set(reflect.Append(*field, reflect.ValueOf(int8(i))))
			case []int16:
				i, err := strconv.ParseInt(v, 10, 16)
				if err != nil {
					return err
				}
				field.Set(reflect.Append(*field, reflect.ValueOf(int16(i))))
			case []int32:
				i, err := strconv.ParseInt(v, 10, 32)
				if err != nil {
					return err
				}
				field.Set(reflect.Append(*field, reflect.ValueOf(int32(i))))
			case []int64:
				i, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					return err
				}
				field.Set(reflect.Append(*field, reflect.ValueOf(i)))
			}
		}
	case []uint, []uint8, []uint16, []uint32, []uint64:
		for _, v := range value {
			switch iface.(type) {
			case []uint:
				i, err := strconv.ParseUint(v, 10, 32)
				if err != nil {
					return err
				}
				field.Set(reflect.Append(*field, reflect.ValueOf(uint(i))))
			case []uint8:
				i, err := strconv.ParseUint(v, 10, 8)
				if err != nil {
					return err
				}
				field.Set(reflect.Append(*field, reflect.ValueOf(uint8(i))))
			case []uint16:
				i, err := strconv.ParseUint(v, 10, 16)
				if err != nil {
					return err
				}
				field.Set(reflect.Append(*field, reflect.ValueOf(uint16(i))))
			case []uint32:
				i, err := strconv.ParseUint(v, 10, 32)
				if err != nil {
					return err
				}
				field.Set(reflect.Append(*field, reflect.ValueOf(uint32(i))))
			case []uint64:
				i, err := strconv.ParseUint(v, 10, 64)
				if err != nil {
					return err
				}
				field.Set(reflect.Append(*field, reflect.ValueOf(i)))
			}
		}
	case []bool:
		for _, v := range value {
			b, err := parseBool(v)
			if err != nil {
				return err
			}
			field.Set(reflect.Append(*field, reflect.ValueOf(b)))
		}
	case []float32, []float64:
		for _, v := range value {
			switch iface.(type) {
			case []float32:
				i, err := strconv.ParseFloat(v, 32)
				if err != nil {
					return err
				}
				field.Set(reflect.Append(*field, reflect.ValueOf(float32(i))))
			case []float64:
				i, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return err
				}
				field.Set(reflect.Append(*field, reflect.ValueOf(i)))
			}
		}
	case []string:
		field.Set(reflect.ValueOf(value))

	// Give up :-(
	default:
		return fmt.Errorf("don't know how to set fields of the type %s",
			field.Type().String())
	}

	return nil
}

func parseBool(v string) (bool, error) {
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on", "enable", "enabled":
		return true, nil
	case "0", "false", "no", "off", "disable", "disabled":
		return false, nil
	default:
		return false, fmt.Errorf("unable to parse %s as a boolean", v)
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
