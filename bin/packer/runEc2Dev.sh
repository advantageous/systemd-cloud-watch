#!/usr/bin/env bash

source ec2_env.sh


instance_id=$(aws ec2 run-instances --image-id "$ami" --subnet-id  "$subnet" \
 --instance-type m4.large --iam-instance-profile "Name=$iam_profile" \
 --associate-public-ip-address --security-group-ids "$security_group" \
 --key-name "$key_name" | jq --raw-output .Instances[].InstanceId)

echo "${instance_id} is being created"

aws ec2 wait instance-exists --instance-ids "$instance_id"

aws ec2 create-tags --resources "${instance_id}" --tags Key=Name,Value="i.int.dev.systemd.cloudwatch"

echo "${instance_id} was tagged waiting to login"

aws ec2 wait instance-status-ok --instance-ids "$instance_id"

./loginIntoEc2Dev.sh



