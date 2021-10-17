package regexp

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"zgo.at/sconfig"
)

func TestRegexp(t *testing.T) {
	cases := []struct {
		fun         sconfig.TypeHandler
		in          []string
		expected    interface{}
		expectedErr error
	}{
		{handleRegexp, []string{"a"}, regexp.MustCompile(`a`), nil},
		{handleRegexp, []string{"[", "A-Z", "]"}, regexp.MustCompile("[A-Z]"), nil},
		{handleRegexp, []string{"("}, nil, errors.New("error parsing regexp: missing closing ): `(`")},
		{handleRegexp, []string{"[", "a-z", "0-9", "]"}, regexp.MustCompile("[a-z0-9]"), nil},

		{
			handleRegexpSlice,
			[]string{"[a-z]", "[0-9]"},
			[]*regexp.Regexp{regexp.MustCompile("[a-z]"), regexp.MustCompile("[0-9]")},
			nil,
		},
		{
			handleRegexpSlice,
			[]string{"[a-z]", "[0-9"},
			nil,
			errors.New("error parsing regexp: missing closing ]: `[0-9`"),
		},
		{
			handleRegexpSlice,
			[]string{"[a-z", "[0-9]"},
			nil,
			errors.New("error parsing regexp: missing closing ]: `[a-z`"),
		},
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
