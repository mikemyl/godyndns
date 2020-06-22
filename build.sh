#!/bin/sh

set -eux

PROJECT_ROOT="/go/src/github.com/${GITHUB_REPOSITORY}"

mkdir -p $PROJECT_ROOT
rmdir $PROJECT_ROOT
ln -s $GITHUB_WORKSPACE $PROJECT_ROOT
cd $PROJECT_ROOT
go get -v ./...

EXT=''

if [ "$GOOS" == 'windows' ]; then
  EXT='.exe'
fi

go build -o "godaddy-dyndns${EXT}" cmd/godaddy-dyndns/main.go
go build -o "public-ip-getter${EXT}" cmd/public-ip-getter/main.go


echo "godaddy-dyndns${EXT}" "public-ip-getter${EXT}"
