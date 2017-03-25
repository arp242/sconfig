// Copyright © 2016-2017 Martin Tournoij
// See the bottom of this file for the full copyright.

package big

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"arp242.net/sconfig"
)

func TestMath(t *testing.T) {
	cases := []struct {
		fun         sconfig.TypeHandler
		in          []string
		expected    interface{}
		expectedErr error
	}{
		// TODO: test big values to make sure it won't overflow
		{handleInt, []string{"42"}, big.NewInt(42), nil},
		{handleInt, []string{"42.1"}, nil, fmt.Errorf(errHandleInt, 42.1)},

		//{handleFloat, []string{"42"}, big.NewFloat(42), nil},
		//{handleFloat, []string{"42.1"}, big.NewFloat(42.1), nil},
		//{handleFloat, []string{"4x"}, nil, fmt.Errorf(errHandleFloat, "4x")},

		//{handleInt, []string{"42"}, big.NewRat(42), nil},
		//{handleInt, []string{"42.1"}, nil, fmt.Errorf(errHandleRat, 42.1)},
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
