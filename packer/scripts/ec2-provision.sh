#!/bin/bash

sudo chmod +x /home/centos/000-provision.sh
sudo /home/centos/000-provision.sh

echo 'export PATH=$PATH:/usr/local/go/bin' >> /home/centos/.bash_profile
echo 'export GOPATH=/gopath' >> /home/centos/.bash_profile
chown centos /home/centos/.bash_profile

sudo mkdir -p /gopath/src/github.com/RichardHightower/
sudo chown centos /gopath/src/github.com/RichardHightower/
git clone https://github.com/RichardHightower/systemd-cloud-watch.git /gopath/src/github.com/RichardHightower/systemd-cloud-watch



sudo chown -R centos /gopath

