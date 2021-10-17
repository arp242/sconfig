package url

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"zgo.at/sconfig"
)

func TestURL(t *testing.T) {
	cases := []struct {
		fun     sconfig.TypeHandler
		in      []string
		want    interface{}
		wantErr string
	}{
		{handleURL, []string{"%"}, nil, "invalid URL escape"},
		{handleURL, []string{"http://example.com/path"}, &url.URL{
			Scheme: "http",
			Host:   "example.com",
			Path:   "/path",
		}, ""},

		{handleURLSlice, []string{"http://example.com/path", "https://example.net"}, []*url.URL{
			{Scheme: "http", Host: "example.com", Path: "/path"},
			{Scheme: "https", Host: "example.net"},
		}, ""},
		{handleURLSlice, []string{"example.com", "%"}, nil, "invalid URL escape"},
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
