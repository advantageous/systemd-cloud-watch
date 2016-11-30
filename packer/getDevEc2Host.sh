#!/usr/bin/env bash

aws ec2 describe-instances --filters  "Name=tag:Name,Values=i.int.dev.systemd.cloudwatch" | jq --raw-output .Reservations[].Instances[].PublicDnsName
