#!/usr/bin/env bash


rm systemd-cloud-watch_linux

set -e

cd /gopath/src/github.com/advantageous/systemd-cloud-watch/

systemd-cat echo "RUNNING JAVA BATCH JOB - ADF BATCH from `pwd`"


echo "Running go clean"
go clean
echo "Running go get"
go get
echo "Running go build"
go build
echo "Running go test"
go test -v github.com/advantageous/systemd-cloud-watch/cloud-watch
echo "Renaming output to _linux"
mv systemd-cloud-watch systemd-cloud-watch_linux


