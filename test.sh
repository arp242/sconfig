#!/bin/sh

set -euC

pkgname=$(go list .)

# Cache some stuff
go test -race -covermode=atomic -i .

# Find all packages that depend on this package. We can pass this to -coverpkg
# so that lines in these packages are counted as well.
find_deps() {
	(
		echo "$1"
		go list -f $'{{range $f := .Deps}}{{$f}}\n{{end}}' "$1"
		go list -f $'{{range $f := .TestImports}}{{$f}}\n{{end}}' "$1" | 
			while read imp; do
				go list -f $'{{range $f := .Deps}}{{$f}}\n{{end}}' "$imp";
			done
	) | sort -u | grep ^$pkgname | grep -v /vendor/ |
		tr '\n' ' ' | sed 's/ $//' | tr ' ' ','
}

echo 'mode: atomic' >| coverage.txt
for pkg in $(go list ./... | grep -v /vendor/); do
	go test -race \
		-covermode=atomic \
		-coverprofile=coverage.tmp \
		-coverpkg=$(find_deps "$pkg") \
		"$pkg" 2>&1 | grep -v 'warning: no packages being tested depend on '

	if [ -f coverage.tmp ]; then
		tail -n+2 coverage.tmp >> coverage.txt
		rm coverage.tmp
	fi
done

[ -n "${TRAVIS:-}" ] && curl -s https://codecov.io/bash | bash
