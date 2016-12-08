#!/usr/bin/env bash

priorities=("emerg" "alert" "crit" "err" "warning" "notice" "info" "debug")

for x in {1..100}
do
    for priority in "${priorities[@]}"
    do
        echo "[$priority] TEST WITH LATEST LEVEL $x" | systemd-cat -p "$priority"
    done
done
