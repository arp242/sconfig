package sconfig_test

import (
	"fmt"
	"net"
	"regexp"

	"arp242.net/sconfig"
	_ "arp242.net/sconfig/handlers/regexp"
)

type Config struct {
	Port    int64
	BaseURL string
	Match   []*regexp.Regexp
	Order   []string
	Hosts   []string
	Address string
}

func Example() {
	config := Config{}
	err := sconfig.Parse(&config, "example.config", sconfig.Handlers{
		// Custom handler
		"address": func(line []string) error {
			addr, err := net.LookupHost(line[0])
			if err != nil {
				return err
			}

			config.Address = addr[0]
			return nil
		},
	})
	if err != nil {
		panic(fmt.Errorf("error parsing config: %s", err))
	}

	fmt.Printf("%+v\n", config)

	// Output: {Port:8080 BaseURL:http://example.com Match:[^foo.+ ^b[ao]r] Order:[allow deny] Hosts:[arp242.net goatcounter.com] Address:arp242.net}
}
