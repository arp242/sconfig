package template

import (
	"fmt"
	"html/template"
	"reflect"
	"testing"

	"zgo.at/sconfig"
)

func TestTemplate(t *testing.T) {
	cases := []struct {
		fun         sconfig.TypeHandler
		in          []string
		expected    interface{}
		expectedErr error
	}{
		{handleHTML, []string{"a"}, template.HTML("a"), nil},
		{handleHTML, []string{"a", "b"}, template.HTML("a b"), nil},
		{handleHTML, []string{"<a>"}, template.HTML("<a>"), nil},
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
