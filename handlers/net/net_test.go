package net

import (
	"fmt"
	"net"
	"reflect"
	"strings"
	"testing"

	"arp242.net/sconfig"
)

func TestNet(t *testing.T) {
	cases := []struct {
		fun     sconfig.TypeHandler
		in      []string
		want    interface{}
		wantErr string
	}{
		{
			handleIP, []string{"127.0.0.1"},
			net.IP{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xff, 0x7f, 0x0, 0x0, 0x1},
			"",
		},
		{
			handleIP, []string{"127.0.0.1X"},
			nil, "not a valid IP address: 127.0.0.1X",
		},
		{
			handleIPSlice, []string{"127.0.0.1", "192.168.0.1"},
			[]net.IP{
				{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xff, 0x7f, 0x0, 0x0, 0x1},
				{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xff, 0xc0, 0xa8, 0x0, 0x1},
			},
			"",
		},
		{
			handleIPSlice, []string{"127.0.0.1", "127.0.0.1X"},
			nil, "not a valid IP address: 127.0.0.1X",
		},
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
