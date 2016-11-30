#!/usr/bin/env bash
docker pull advantageous/golang-cloud-watch:latest
docker run  -it --name runner2  \
-p 80:80 \
-v `pwd`:/gopath/src/github.com/RichardHightower/systemd-cloud-watch \
advantageous/golang-cloud-watch:latest
docker rm runner2
