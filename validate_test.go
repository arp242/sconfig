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
