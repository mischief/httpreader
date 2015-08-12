#!/bin/sh

ORG="github.com/mischief"
PKG="${ORG}/httpreader"

if [ ! -h gopath/src/${PKG} ]; then
	mkdir -p gopath/src/${ORG}
	ln -s ../../../.. gopath/src/${PKG} || exit 255
fi

GOPATH=$(pwd)/gopath go test

export GOBIN=${PWD}/bin
export GOPATH=${PWD}/gopath

eval $(go env)

go test -v -cover -race
