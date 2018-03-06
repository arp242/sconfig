// Copyright © 2016-2017 Martin Tournoij
// See the bottom of this file for the full copyright.

// Package big contains handlers for parsing values with the math/big package.
//
// It currently implements the big.Int and big.Float types.
package big // import "arp242.net/sconfig/handlers/big"

import (
	"fmt"
	"math/big"
	"strings"

	"arp242.net/sconfig"
)

var (
	errHandleInt   = "unable to convert %v to big.Int"
	errHandleFloat = "unable to convert %v to big.Float"
)

func init() {
	sconfig.RegisterType("*big.Int", sconfig.ValidateSingleValue(), handleInt)
	sconfig.RegisterType("*big.Float", sconfig.ValidateSingleValue(), handleFloat)
	sconfig.RegisterType("[]*big.Int", sconfig.ValidateValueLimit(1, 0), handleIntSlice)
	sconfig.RegisterType("[]*big.Float", sconfig.ValidateValueLimit(1, 0), handleFloatSlice)
}

func handleInt(v []string) (interface{}, error) {
	n := big.Int{}
	z, success := n.SetString(strings.Join(v, ""), 10)
	if !success {
		return nil, fmt.Errorf(errHandleInt, strings.Join(v, ""))
	}
	return z, nil
}

func handleFloat(v []string) (interface{}, error) {
	n := big.Float{}
	z, success := n.SetString(strings.Join(v, ""))
	if !success {
		return nil, fmt.Errorf(errHandleFloat, strings.Join(v, ""))
	}
	return z, nil
}

func handleIntSlice(v []string) (interface{}, error) {
	a := make([]*big.Int, len(v))
	for i := range v {
		a[i] = &big.Int{}
		z, success := a[i].SetString(v[i], 10)
		if !success {
			return nil, fmt.Errorf(errHandleInt, v[i])
		}
		a[i] = z
	}
	return a, nil
}

func handleFloatSlice(v []string) (interface{}, error) {
	a := make([]*big.Float, len(v))
	for i := range v {
		a[i] = &big.Float{}
		z, success := a[i].SetString(v[i])
		if !success {
			return nil, fmt.Errorf(errHandleFloat, v[i])
		}
		a[i] = z
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
