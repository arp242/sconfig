// Package url contains handlers for parsing values with the net/url package.
//
// It currently implements the url.URL type. Note Go's url package does not do a
// lot of validation, and will happily "parse" wildly invalid URLs without
// returning an error.
package url

import (
	"net/url"
	"strings"

	"zgo.at/sconfig"
)

func init() {
	sconfig.RegisterType("*url.URL", sconfig.ValidateSingleValue(), handleURL)
	sconfig.RegisterType("[]*url.URL", sconfig.ValidateValueLimit(1, 0), handleURLSlice)
}

func handleURL(v []string) (interface{}, error) {
	u, err := url.Parse(strings.Join(v, ""))
	if err != nil {
		return nil, err
	}
	return u, nil
}

func handleURLSlice(v []string) (interface{}, error) {
	a := make([]*url.URL, len(v))
	for i := range v {
		u, err := url.Parse(v[i])
		if err != nil {
			return nil, err
		}
		a[i] = u
	}
	return a, nil
}
