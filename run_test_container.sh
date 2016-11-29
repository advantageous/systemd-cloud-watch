#!/usr/bin/env bash
docker pull advantageous/golang-cloud-watch:latest
docker run -it --name runner -v `pwd`:/source/src/github.com/RichardHightower/systemd-cloud-watch \
advantageous/golang-cloud-watch
docker rm runner
