// Copyright © 2016-2017 Martin Tournoij
// See the bottom of this file for the full copyright.

// Package net TODO
package net // import "arp242.net/sconfig/handlers/net"
import (
	"fmt"
	"net"
	"strings"

	"arp242.net/sconfig"
)

func init() {
	sconfig.RegisterType("IP", sconfig.ValidateSingleValue(), handleIP)
	sconfig.RegisterType("IPAddr", sconfig.ValidateSingleValue(), handleIPAddr)
	sconfig.RegisterType("IPMask", sconfig.ValidateSingleValue(), handleIPMask)
	sconfig.RegisterType("IPNet", sconfig.ValidateSingleValue(), handleIPNet)
}

// handleIP parses an IPv4 or IPv6 address
func handleIP(v []string) (interface{}, error) {
	IP, IPNet, err := net.ParseCIDR(strings.Join(v, ""))
	_ = IPNet // TODO: What to do with this?
	if err != nil {
		IP = net.ParseIP(v[0])
	}
	if IP == nil {
		return nil, fmt.Errorf("not a valid IP address: %v", v[0])
	}
	return IP, nil
}

func handleIPAddr(v []string) (interface{}, error) {
	return nil, nil
}
func handleIPMask(v []string) (interface{}, error) {
	return nil, nil
}
func handleIPNet(v []string) (interface{}, error) {
	return nil, nil
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
