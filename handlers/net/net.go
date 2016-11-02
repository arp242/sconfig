package time

import (
	"fmt"
	"net"

	"arp242.net/sconfig"
)

func init() {
	sconfig.RegisterType("IP", sconfig.ValidateSingleValue, HandleIP)
	//sconfig.TypeHandlers["IPaddr"] =
	//sconfig.TypeHandlers["IPMask"] =
	//sconfig.TypeHandlers["IPNet"] =
}

// HandleIP parses an IPv4 or IPv6 address
func HandleIP(v []string) (interface{}, error) {
	IP, IPNet, err := net.ParseCIDR(v[0])
	_ = IPNet // TODO: What to do with this?
	if err != nil {
		IP = net.ParseIP(v[0])
	}
	if IP == nil {
		return nil, fmt.Errorf("not a valid IP address: %v", v[0])
	}
	return IP, nil
}
