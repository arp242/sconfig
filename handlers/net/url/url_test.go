// Copyright © 2016-2017 Martin Tournoij
// See the bottom of this file for the full copyright.

package url

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"arp242.net/sconfig"
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
