#!/usr/bin/env bash


rm systemd-cloud-watch_linux

set -e

cd /gopath/src/github.com/advantageous/systemd-cloud-watch/
source ~/.bash_profile
export GOPATH=/gopath


/usr/lib/systemd/systemd-journald &

priorities=("emerg" "alert" "crit" "err" "warning" "notice" "info" "debug")

for x in {1..100}
do
    for priority in "${priorities[@]}"
    do
        echo "[$priority] TEST WITH LATEST LEVEL $x" | systemd-cat -p "$priority"
    done
done


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

pkill -9 systemd
