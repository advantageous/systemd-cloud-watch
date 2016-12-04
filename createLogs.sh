#!/usr/bin/env bash

for x in {1..100}
do
    for i in {1..7}
    do
        echo "$i JOURNAL D TEST $x"
        systemd-cat echo "$i JOURNAL D TEST $x"
    done
done