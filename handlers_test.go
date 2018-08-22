// Copyright © 2016-2017 Martin Tournoij
// See the bottom of this file for the full copyright.

package sconfig

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestHandlers(t *testing.T) {
	cases := []struct {
		fun     TypeHandler
		in      []string
		want    interface{}
		wantErr string
	}{
		{handleString, []string{}, "", ""},
		{handleString, []string{"H€llo"}, "H€llo", ""},
		{handleString, []string{"Hello", "world!"}, "Hello world!", ""},
		{handleString, []string{"3.14"}, "3.14", ""},

		{handleBool, []string{"false"}, false, ""},
		{handleBool, []string{"TRUE"}, true, ""},
		{handleBool, []string{"enabl", "ed"}, true, ""},
		{handleBool, []string{}, true, ""},
		{handleBool, []string{"it is true"}, nil, `unable to parse "it is true" as a boolean`},

		{handleFloat32, []string{}, nil, `strconv.ParseFloat: parsing "": invalid syntax`},
		{handleFloat32, []string{"0.0"}, float32(0.0), ""},
		{handleFloat32, []string{".000001"}, float32(0.000001), ""},
		{handleFloat32, []string{"1"}, float32(1), ""},
		{handleFloat32, []string{"1.1", "12"}, float32(1.112), ""},

		{handleFloat64, []string{}, nil, `strconv.ParseFloat: parsing "": invalid syntax`},
		{handleFloat64, []string{"0.0"}, float64(0.0), ""},
		{handleFloat64, []string{".000001"}, float64(0.000001), ""},
		{handleFloat64, []string{"1"}, float64(1), ""},
		{handleFloat64, []string{"1.1", "12"}, float64(1.112), ""},

		{handleStringMap, []string{"a", "b"}, map[string]string{"a": "b"}, ""},
		{handleStringMap, []string{"a", "b", "x", "y"}, map[string]string{"a": "b", "x": "y"}, ""},
		{handleStringMap, []string{"a", "b", "x"}, nil, "uneven number of arguments: 3"},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out, err := tc.fun(tc.in)
			if !errorContains(err, tc.wantErr) {
				t.Errorf("err wrong\nwant: %v\nout:  %v\n", tc.wantErr, err)
			}
			if !reflect.DeepEqual(out, tc.want) {
				t.Errorf("\nwant: %#v\nout:  %#v\n", tc.want, out)
			}
		})
	}
}

func errorContains(out error, want string) bool {
	if out == nil {
		return want == ""
	}
	if want == "" {
		return false
	}
	return strings.Contains(out.Error(), want)
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
