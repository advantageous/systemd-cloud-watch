#!/bin/bash
set -e

echo Install log agent -------------------------------
mkdir /tmp/logagent
cd /tmp/logagent
curl -OL https://github.com/RichardHightower/systemd-cloud-watch/releases/download/v0.0.1-prerelease/systemd-cloud-watch_linux
sudo mv systemd-cloud-watch_linux /usr/bin
sudo chmod +x  /usr/bin/systemd-cloud-watch_linux
sudo mkdir -p  /var/lib/journald-cloudwatch-logs/
sudo mv /home/centos/etc/journald-cloudwatch.conf /etc/
sudo mv /home/centos/etc/systemd/system/journald-cloudwatch.service /etc/systemd/system/journald-cloudwatch.service
sudo chmod 664 /etc/systemd/system/journald-cloudwatch.service
sudo chown -R centos /var/lib/journald-cloudwatch-logs/
sudo systemctl enable journald-cloudwatch.service
sudo rm -rf /tmp/llogagent
echo DONE installing log agent -------------------------------
