// Copyright © 2016-2017 Martin Tournoij
// See the bottom of this file for the full copyright.

package sconfig

import (
	"fmt"
	"reflect"
	"testing"
)

func TestValidate(t *testing.T) {
	cases := []struct {
		fun         TypeHandler
		in          []string
		expectedErr error
	}{
		{ValidateNoValue(), []string{}, nil},
		{ValidateNoValue(), []string{"1"}, errValidateNoValue},
		{ValidateNoValue(), []string{"asd", "zxa"}, errValidateNoValue},

		{ValidateSingleValue(), []string{"qwe"}, nil},
		{ValidateSingleValue(), []string{}, errValidateSingleValue},
		{ValidateSingleValue(), []string{"asd", "zxc"}, errValidateSingleValue},

		{ValidateValueLimit(0, 1), []string{}, nil},
		{ValidateValueLimit(0, 1), []string{"Asd"}, nil},
		{ValidateValueLimit(0, 1), []string{"zxc", "asd"}, fmt.Errorf(errValidateValueLimitFewer, 1, 2)},

		{ValidateValueLimit(2, 3), []string{}, fmt.Errorf(errValidateValueLimitMore, 2, 0)},
		{ValidateValueLimit(2, 3), []string{"ads"}, fmt.Errorf(errValidateValueLimitMore, 2, 1)},
		{ValidateValueLimit(2, 3), []string{"ads", "asd"}, nil},
		{ValidateValueLimit(2, 3), []string{"ads", "zxc", "qwe"}, nil},
		{ValidateValueLimit(2, 3), []string{"ads", "zxc", "qwe", "hjkl"}, fmt.Errorf(errValidateValueLimitFewer, 3, 4)},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out, err := tc.fun(tc.in)

			switch tc.expectedErr {
			case nil:
				if err != nil {
					t.Errorf("expected err to be nil; is: %#v", err)
				}
				if !reflect.DeepEqual(out, tc.in) {
					t.Errorf("out wrong\nexpected:  %#v\nout:       %#v\n",
						tc.in, out)
				}
			default:
				if err.Error() != tc.expectedErr.Error() {
					t.Errorf("err wrong\nexpected:  %#v\nout:       %#v\n",
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
