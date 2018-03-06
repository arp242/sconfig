// Copyright © 2016-2017 Martin Tournoij
// See the bottom of this file for the full copyright.

// Package url contains handlers for parsing values with the net/url package.
//
// It currently implements the url.URL type. Note Go's url package does not do a
// lot of validation, and will happily "parse" wildly invalid URLs without
// returning an error.
package url // import "arp242.net/sconfig/handlers/net/url"

import (
	"net/url"
	"strings"

	"arp242.net/sconfig"
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
