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
