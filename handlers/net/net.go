// Package net contains handlers for parsing values with the net package.
//
// It currently implements the net.IP type.
package net

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
