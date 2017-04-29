#!/bin/sh

set -euC

echo >| coverage.txt

go test -race -covermode=atomic -i ./...

for d in $(go list ./...); do
	go test -race -covermode=atomic -coverprofile=coverage.tmp $d
	if [ -f coverage.tmp ]; then
		cat coverage.tmp >> coverage.txt
		rm coverage.tmp
	fi
done

curl -s https://codecov.io/bash | bash
rm coverage.txt
