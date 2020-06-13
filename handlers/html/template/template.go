// Package template contains handlers for parsing values with the html/template package.
package template

import (
	"html/template"
	"strings"

	"arp242.net/sconfig"
)

func init() {
	sconfig.RegisterType("template.HTML", handleHTML)
}

func handleHTML(v []string) (interface{}, error) {
	return template.HTML(strings.Join(v, " ")), nil
}
