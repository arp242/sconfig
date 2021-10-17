// Package big contains handlers for parsing values with the math/big package.
//
// It currently implements the big.Int and big.Float types.
package big

import (
	"fmt"
	"math/big"
	"strings"

	"zgo.at/sconfig"
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
