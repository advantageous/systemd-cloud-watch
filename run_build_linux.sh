#!/usr/bin/env bash
docker pull advantageous/golang-cloud-watch:latest
docker run -it --name build -v `pwd`:/gopath/src/github.com/RichardHightower/systemd-cloud-watch \
advantageous/golang-cloud-watch \
/bin/sh -c "/source/src/github.com/RichardHightower/systemd-cloud-watch/build_linux.sh"
docker rm build
