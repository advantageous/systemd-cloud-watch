#!/usr/bin/env bash
set -e

cd /source/src/github.com/RichardHightower/systemd-cloud-watch/
source ~/.bash_profile
export GOPATH=/source
go clean
go get
go build
mv systemd-cloud-watch systemd-cloud-watch_linux
