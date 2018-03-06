// Copyright © 2016-2017 Martin Tournoij
// See the bottom of this file for the full copyright.

// Package net contains handlers for parsing values with the net package.
//
// It currently implements the net.IP type.
package net // import "arp242.net/sconfig/handlers/net"
import (
	"fmt"
	"net"
	"strings"

	"arp242.net/sconfig"
)

func init() {
	sconfig.RegisterType("net.IP", sconfig.ValidateSingleValue(), handleIP)
	sconfig.RegisterType("[]net.IP", sconfig.ValidateValueLimit(1, 0), handleIPSlice)
}

// handleIP parses an IPv4 or IPv6 address
func handleIP(v []string) (interface{}, error) {
	IP, _, err := net.ParseCIDR(strings.Join(v, ""))
	if err != nil {
		IP = net.ParseIP(v[0])
	}
	if IP == nil {
		return nil, fmt.Errorf("not a valid IP address: %v", v[0])
	}
	return IP, nil
}

func handleIPSlice(v []string) (interface{}, error) {
	a := make([]net.IP, len(v))
	for i := range v {
		ip, err := handleIP([]string{v[i]})
		if err != nil {
			return nil, err
		}
		a[i] = ip.(net.IP)
	}
	return a, nil
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
