package example

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"testing"

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

func TestExample(t *testing.T) {
	config := Config{}
	err := sconfig.Parse(&config, "config", sconfig.Handlers{
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
		fmt.Fprintf(os.Stderr, "Error parsing config: %v", err)
		t.Fail()
	}

	fmt.Printf("%+v\n", config)
}
