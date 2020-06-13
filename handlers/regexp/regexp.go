// Package regexp contains handlers for parsing values with the regexp package.
//
// It currently implements the regexp.Regexp types.
package regexp

import (
	"regexp"
	"strings"

	"arp242.net/sconfig"
)

func init() {
	sconfig.RegisterType("*regexp.Regexp", sconfig.ValidateSingleValue(), handleRegexp)
	sconfig.RegisterType("[]*regexp.Regexp", sconfig.ValidateValueLimit(1, 0), handleRegexpSlice)
}

func handleRegexp(v []string) (interface{}, error) {
	r, err := regexp.Compile(strings.Join(v, ""))
	if err != nil {
		return nil, err
	}

	return r, nil
}

func handleRegexpSlice(v []string) (interface{}, error) {
	a := make([]*regexp.Regexp, len(v))
	for i := range v {
		r, err := regexp.Compile(v[i])
		if err != nil {
			return nil, err
		}
		a[i] = r
	}
	return a, nil
}
