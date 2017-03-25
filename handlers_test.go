// Copyright © 2016-2017 Martin Tournoij
// See the bottom of this file for the full copyright.

package sconfig

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func TestHandlers(t *testing.T) {
	cases := []struct {
		fun         TypeHandler
		in          []string
		expected    interface{}
		expectedErr error
	}{
		{handleString, []string{}, "", nil},
		{handleString, []string{"H€llo"}, "H€llo", nil},
		{handleString, []string{"Hello", "world!"}, "Hello world!", nil},
		{handleString, []string{"3.14"}, "3.14", nil},

		{handleBool, []string{"false"}, false, nil},
		{handleBool, []string{"TRUE"}, true, nil},
		{handleBool, []string{"enabl", "ed"}, true, nil},
		{handleBool, []string{}, false, errors.New(`unable to parse "" as a boolean`)},
		{handleBool, []string{"it is true"}, false, errors.New(`unable to parse "it is true" as a boolean`)},

		{handleFloat32, []string{}, float32(0.0), errors.New(`strconv.ParseFloat: parsing "": invalid syntax`)},
		{handleFloat32, []string{"0.0"}, float32(0.0), nil},
		{handleFloat32, []string{".000001"}, float32(0.000001), nil},
		{handleFloat32, []string{"1"}, float32(1), nil},
		{handleFloat32, []string{"1.1", "12"}, float32(1.112), nil},

		{handleFloat64, []string{}, float64(0.0), errors.New(`strconv.ParseFloat: parsing "": invalid syntax`)},
		{handleFloat64, []string{"0.0"}, float64(0.0), nil},
		{handleFloat64, []string{".000001"}, float64(0.000001), nil},
		{handleFloat64, []string{"1"}, float64(1), nil},
		{handleFloat64, []string{"1.1", "12"}, float64(1.112), nil},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out, err := tc.fun(tc.in)

			switch tc.expectedErr {
			case nil:
				if err != nil {
					t.Errorf("expected err to be nil; is: %#v", err)
				}
				if !reflect.DeepEqual(out, tc.expected) {
					t.Errorf("out wrong\nexpected:  %#v\nout:       %#v\n",
						tc.expected, out)
				}
			default:
				if err.Error() != tc.expectedErr.Error() {
					t.Errorf("err wrong\nexpected:  %v\nout:       %v\n",
						tc.expectedErr, err)
				}

				if out != nil {
					t.Errorf("out should be nil if there's an error")
				}
			}

		})
	}
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
