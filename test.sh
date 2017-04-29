#!/bin/sh

set -euC

# Cache some stuff
go test -race -covermode=atomic -i ./...

echo >| coverage.txt
for d in $(go list ./...); do
	go test -race -covermode=atomic -coverprofile=coverage.tmp $d
	if [ -f coverage.tmp ]; then
		cat coverage.tmp >> coverage.txt
		rm coverage.tmp
	fi
done

curl -s https://codecov.io/bash | bash
rm coverage.txt
