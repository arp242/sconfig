package regexp

import (
	"errors"
	"regexp"

	"arp242.net/sconfig"
)

func init() {
	sconfig.RegisterType("[]*regexp.Regexp", handleRegexpSlice)
	sconfig.RegisterType("*regexp.Regexp", handleRegexp)
}

func handleRegexp(v []string) (interface{}, error) {
	if len(v) != 1 {
		return nil, errors.New("must have exactly one value")
	}

	r, err := regexp.Compile(v[0])
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
