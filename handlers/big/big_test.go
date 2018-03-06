// Copyright © 2016-2017 Martin Tournoij
// See the bottom of this file for the full copyright.

package big

import (
	"fmt"
	"math/big"
	"strings"
	"testing"

	"arp242.net/sconfig"
)

func TestMath(t *testing.T) {
	cases := []struct {
		fun     sconfig.TypeHandler
		in      []string
		want    interface{}
		wantErr string
	}{
		{handleInt, []string{"42"}, big.NewInt(42), ""},
		{handleInt, []string{"42.1"}, nil, fmt.Sprintf(errHandleInt, 42.1)},
		{handleInt, []string{"9223372036854775808"},
			big.NewInt(0).Add(big.NewInt(9223372036854775807), big.NewInt(1)),
			""},

		{handleFloat, []string{"42"}, big.NewFloat(42), ""},
		{handleFloat, []string{"42.1"}, big.NewFloat(42.1), ""},
		{handleFloat, []string{"4x"}, nil, fmt.Sprintf(errHandleFloat, "4x")},

		{handleIntSlice, []string{"100", "101"}, []*big.Int{big.NewInt(100), big.NewInt(101)}, ""},
		{handleIntSlice, []string{"100", "10x1"}, nil, "unable to convert 10x1 to big.Int"},
		{handleFloatSlice, []string{"100", "101"}, []*big.Float{big.NewFloat(100), big.NewFloat(101)}, ""},
		{handleFloatSlice, []string{"100", "10x1"}, nil, "unable to convert 10x1 to big.Float"},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) {
			out, err := tc.fun(tc.in)
			if !errorContains(err, tc.wantErr) {
				t.Errorf("err wrong\nwant: %v\nout:  %v\n", tc.wantErr, err)
			}

			o := fmt.Sprintf("%#v", out)
			w := fmt.Sprintf("%#v", tc.want)
			if o != w {
				t.Errorf("\nwant: %#v (%[1]T)\nout:  %#v (%[2]T)\n", tc.want, out)
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
