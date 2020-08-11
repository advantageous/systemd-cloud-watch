# Systemd Journal CloudWatch Writer

This utility reads from the [systemd journal](https://www.freedesktop.org/software/systemd/man/systemd-journald.service.html),
 and sends the data in batches to [Cloudwatch](https://aws.amazon.com/cloudwatch/).
 
This is an alternative process to the AWS-provided logs agent.
The AWS logs agent copies data from on-disk text log files into [Cloudwatch](https://aws.amazon.com/cloudwatch/).
This utility `systemd-cloud-watch` reads the `systemd journal` and writes that data in batches to CloudWatch.

There are other ways to do this using various techniques. But depending on the size of log messages and size of the core parts
these other methods are fragile as AWS CloudWatch limits the size of the messages. 
This utility allows you cap the log field size, include only the fields that you want, or
exclude the fields you don't want. We find that this is not only useful but essential. 


## Log format

The journal event data is written to ***CloudWatch*** Logs in JSON format, making it amenable to filtering using the JSON filter syntax.
Log records are translated to ***CloudWatch*** JSON events using a structure like the following:

#### Sample log
```json
{
    "instanceId" : "i-xxxxxxxx",
    "pid" : 12354,
    "uid" : 0,
    "gid" : 0,
    "cmdName" : "cron",
    "exe" : "/usr/sbin/cron",
    "cmdLine" : "/usr/sbin/CRON -f",
    "systemdUnit" : "cron.service",
    "bootId" : "fa58079c7a6d12345678b6ebf1234567",
    "hostname" : "ip-10-1-0-15",
    "transport" : "syslog",
    "priority" : "INFO",
    "message" : "pam_unix(cron:session): session opened for user root by (uid=0)",
    "syslogFacility" : 10,
    "syslogIdent" : "CRON"
}
```

The JSON-formatted log events could also be exported into an AWS ElasticSearch instance using the ***CloudWatch***
sync mechanism. Once in ElasticSearch, you can use an ELK stack to obtain more elaborate filtering and query capabilities.


## Installation

If you have a binary distribution, you just need to drop the executable file somewhere.

This tool assumes that it is running on an EC2 instance.

This tool uses `libsystemd` to access the journal. systemd-based distributions generally ship
with this already installed, but if yours doesn't you must manually install the library somehow before
this tool will work.

There are instructions on how to install the Linux requirements for development below see - 
[Setting up a Linux env for testing/developing (CentOS7)](#setting-up-a-linux-env-for-testingdeveloping-centos7).

We also have two excellent examples of setting up a dev environment using [bin.packer](https://www.packer.io/) for both 
[AWS EC2](#building-the-ec2-image-with-packer-to-build-the-linux-instance-to-build-this-project) and 
[Docker](#building-the-docker-image-to-build-the-linux-instance-to-build-this-project). We setup CentoOS 7.
The EC2 instance bin.packer build uses the ***aws command line*** to create and connect to a running image. 
These should be instructive for how to setup this utility in your environment to run with ***systemd*** as we provide
all of the systemd scripts in the bin.packer provision scripts for EC2.  An example is good. A running example is better.

## Configuration

This tool uses a small configuration file to set some values that are required for its operation.
Most of the configuration values are optional and have default settings, but a couple are required.

The configuration file uses a syntax like this:

```js
log_group = "my-awesome-app"

```

The following configuration settings are supported:

* `aws_region`: (Optional) The AWS region whose CloudWatch Logs API will be written to. If not provided,
  this defaults to the region where the host EC2 instance is running.

* `ec2_instance_id`: (Optional) The id of the EC2 instance on which the tool is running. There is very
  little reason to set this, since it will be automatically set to the id of the host EC2 instance.

* `journal_dir`: (Optional) Override the directory where the systemd journal can be found. This is
  useful in conjunction with remote log aggregation, to work with journals synced from other systems.
  The default is to use the local system's journal.

* `log_group`: (Required) The name of the cloudwatch log group to write logs into.

* `log_priority`: (Optional) The highest priority of the log messages to read (on a 0-7 scale). This defaults
    to DEBUG (all messages). This has a behaviour similar to `journalctl -p <priority>`. At the moment, only
    a single value can be specified, not a range. Possible values are: `0,1,2,3,4,5,6,7` or one of the corresponding
    `"emerg", "alert", "crit", "err", "warning", "notice", "info", "debug"`.
    When a single log level is specified, all messages with this log level or a lower (hence more important)
    log level are read and pushed to CloudWatch. For more information about priority levels, look at
    https://www.freedesktop.org/software/systemd/man/journalctl.html

* `log_stream`: (Optional) The name of the cloudwatch log stream to write logs into. This defaults to
  the EC2 instance name joined with the hostname and region. Each running instance of this application
  (along with any other applications writing logs into the same log group) must have a unique `log_stream`
  value. If the given log stream doesn't exist then it will be created before writing the first set of
  journal events.

* `buffer_size`: (Optional) The size of the event buffer to send to CloudWatch Logs API. The default is 50.
 This means that cloud watch will send 50 logs at a time. 

* `fields`: (Optional) Specifies which fields should be included in the JSON map that is sent to CloudWatch.

* `omit_fields`: (Optional) Specifies which fields should NOT be included in the JSON map that is sent to CloudWatch.

* `field_length`: (Optional) Specifies how long string fileds can be in the JSON  map that is sent to CloudWatch.
   The default is 255 characters.
   
*  `queue_batch_size` : (Optional) Internal. Default to 10,000 entries, how large the queue buffer is. This is chunks of log entries
that can be sent to the cloud watch repeater.
   
*  `queue_channel_size`: (Optional) Internal.  Default to 3 entries, how large the queue buffer is. This is how many `queue_batch_size`
can be around to send before the journald reader waits for the cloudwatch repeater. 

*  `queue_poll_duration_ms` : (Optional) Internal. Default to 10 ms, how long the queue manager will wait if there are no log entries to send
to check again to see if there are log entries to send. 

*  `queue_flush_log_ms` : (Optional) If `queue_batch_size` has not been met because there are no more journald entries to 
read, how long to flush the buffer to cloud watch receiver. Defaults to 100 ms.

* `debug`: (Optional) Turns on debug logging.

* `local`: (Optional) Used for unit testing. Will not try to create an AWS meta-data client to read region and AWS credentials.

* `tail`: (Optional) Start from the tail of log. Only send new log entries. This is good for reboot so you don't send all of the
logs in the system, which is the default behavior. 

* `rewind`: (Optional) Used to rewind X number of entries from the tail of the log. Must be used in conjunction with the 
`tail` setting.

* `mock-cloud-watch` : (Optional) Used to send logs to a Journal Repeater that just spits out message and priority to the console.
This is used for development only. 


If your average log message was 500 bytes, and your used the default setting then assuming the server was generating 
journald messages rapidly you could use a heap of up to `queue_channel_size` (3) * `queue_batch_size`(10,000) * 500 bytes
(15,000,000). If you had a very resource constrained env, reduce the `queue_batch_size` and/or the `queue_channel_size`.



### AWS API access

This program requires access to call some of the Cloudwatch API functions. The recommended way to
achieve this is to create an
[IAM Instance Profile](http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_switch-role-ec2_instance-profiles.html)
that grants your EC2 instance a role that has Cloudwatch API access. The program will automatically
discover and make use of instance profile credentials.

The following IAM policy grants the required access across all log groups in all regions:

#### IAM file
```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:DescribeInstances"
            ],
            "Resource": [
                "*"
            ]
        },
        {
            "Effect": "Allow",
            "Action": [
                "logs:CreateLogGroup",
                "logs:CreateLogStream",
                "logs:PutLogEvents",
                "logs:DescribeLogStreams"
            ],
            "Resource": [
                "arn:aws:logs:*:*:log-group:*",
                "arn:aws:logs:*:*:log-group:*:log-stream:*"
            ]
        }
    ]
}
```

In more complex environments you may want to restrict further which regions, groups and streams
the instance can write to. You can do this by adjusting the two ARN strings in the `"Resource"` section:

* The first `*` in each string can be replaced with an AWS region name like `us-east-1`
  to grant access only within the given region.
* The `*` after `log-group` in each string can be replaced with a Cloudwatch Logs log group name
  to grant access only to the named group.
* The `*` after `log-stream` in the second string can be replaced with a Cloudwatch Logs log stream
  name to grant access only to the named stream.

Other combinations are possible too. For more information, see
[the reference on ARNs and namespaces](http://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html#arn-syntax-cloudwatch-logs).



### Coexisting with the official Cloudwatch Logs agent

This application can run on the same host as the official Cloudwatch Logs agent but care must be taken
to ensure that they each use a different log stream name. Only one process may write into each log
stream.

## Running on System Boot

This program is best used as a persistent service that starts on boot and keeps running until the
system is shut down. If you're using `journald` then you're presumably using systemd; you can create
a systemd unit for this service. For example:

```
[Unit]
Description=journald-cloudwatch-logs
Wants=basic.target
After=basic.target network.target

[Service]
User=nobody
Group=nobody
ExecStart=/usr/local/bin/journald-cloudwatch-logs /usr/local/etc/journald-cloudwatch-logs.conf
KillMode=process
Restart=on-failure
RestartSec=42s
```

This program is designed under the assumption that it will run constantly from some point during
system boot until the system shuts down.

If the service is stopped while the system is running and then later started again, it will
"lose" any journal entries that were written while it wasn't running. However, on the initial
run after each boot it will clear the backlog of logs created during the boot process, so it
is not necessary to run the program particularly early in the boot process unless you wish
to *promptly* capture startup messages.

## Building

#### Test cloud-watch package
```sh
go test -v  github.com/advantageous/systemd-cloud-watch/cloud-watch
```


#### Build and Test on Linux (Centos7)
```sh
 ./run_build_linux.sh
```

The above starts up a docker container, runs `go get`, `go build`, `go test` and then copies the binary to
`systemd-cloud-watch_linux`.

#### Debug process running Linux
```sh
 ./run_test_container.sh
```


The above starts up a docker container that you can develop with that has all the prerequisites needed to
compile and test this project.

#### Sample debug session
```sh
$ ./run_test_container.sh
latest: Pulling from advantageous/golang-cloud-watch
Digest: sha256:eaf5c0a387aee8cc2d690e1c5e18763e12beb7940ca0960ce1b9742229413e71
Status: Image is up to date for advantageous/golang-cloud-watch:latest
[root@6e0d1f984c03 /]# cd gopath/src/github.com/advantageous/systemd-cloud-watch/
.git/                      README.md                  cloud-watch/               bin.packer/                    sample.conf                
.gitignore                 build_linux.sh             main.go                    run_build_linux.sh         systemd-cloud-watch.iml    
.idea/                     cgroup/                    output.json                run_test_container.sh      systemd-cloud-watch_linux  

[root@6e0d1f984c03 /]# cd gopath/src/github.com/advantageous/systemd-cloud-watch/

[root@6e0d1f984c03 systemd-cloud-watch]# ls
README.md  build_linux.sh  cgroup  cloud-watch  main.go  output.json  bin.packer  run_build_linux.sh  
run_test_container.sh  sample.conf  systemd-cloud-watch.iml  systemd-cloud-watch_linux

[root@6e0d1f984c03 systemd-cloud-watch]# source ~/.bash_profile

[root@6e0d1f984c03 systemd-cloud-watch]# export GOPATH=/gopath

[root@6e0d1f984c03 systemd-cloud-watch]# /usr/lib/systemd/systemd-journald &
[1] 24

[root@6e0d1f984c03 systemd-cloud-watch]# systemd-cat echo "RUNNING JAVA BATCH JOB - ADF BATCH from `pwd`"

[root@6e0d1f984c03 systemd-cloud-watch]# echo "Running go clean"
Running go clean

[root@6e0d1f984c03 systemd-cloud-watch]# go clean

[root@6e0d1f984c03 systemd-cloud-watch]# echo "Running go get"
Running go get

[root@6e0d1f984c03 systemd-cloud-watch]# go get

[root@6e0d1f984c03 systemd-cloud-watch]# echo "Running go build"
Running go build
[root@6e0d1f984c03 systemd-cloud-watch]# go build

[root@6e0d1f984c03 systemd-cloud-watch]# echo "Running go test"
Running go test

[root@6e0d1f984c03 systemd-cloud-watch]# go test -v github.com/advantageous/systemd-cloud-watch/cloud-watch
=== RUN   TestRepeater
config DEBUG: 2016/11/30 08:53:34 config.go:66: Loading log...
aws INFO: 2016/11/30 08:53:34 aws.go:42: Config set to local
aws INFO: 2016/11/30 08:53:34 aws.go:72: Client missing credentials not looked up
aws INFO: 2016/11/30 08:53:34 aws.go:50: Client missing using config to set region
aws INFO: 2016/11/30 08:53:34 aws.go:52: AWSRegion missing using default region us-west-2
repeater ERROR: 2016/11/30 08:53:44 cloudwatch_journal_repeater.go:141: Error from putEvents NoCredentialProviders: no valid providers in chain. Deprecated.
	For verbose messaging see aws.Config.CredentialsChainVerboseErrors
--- SKIP: TestRepeater (10.01s)
	cloudwatch_journal_repeater_test.go:43: Skipping WriteBatch, you need to setup AWS credentials for this to work
=== RUN   TestConfig
test DEBUG: 2016/11/30 08:53:44 config.go:66: Loading log...
test INFO: 2016/11/30 08:53:44 config_test.go:33: [Foo Bar]
--- PASS: TestConfig (0.00s)
=== RUN   TestLogOmitField
test DEBUG: 2016/11/30 08:53:44 config.go:66: Loading log...
--- PASS: TestLogOmitField (0.00s)
=== RUN   TestNewJournal
--- PASS: TestNewJournal (0.00s)
=== RUN   TestSdJournal_Operations
--- PASS: TestSdJournal_Operations (0.00s)
	journal_linux_test.go:41: Read value=Runtime journal is using 8.0M (max allowed 4.0G, trying to leave 4.0G free of 55.1G available → current limit 4.0G).
=== RUN   TestNewRecord
test DEBUG: 2016/11/30 08:53:44 config.go:66: Loading log...
--- PASS: TestNewRecord (0.00s)
=== RUN   TestLimitFields
test DEBUG: 2016/11/30 08:53:44 config.go:66: Loading log...
--- PASS: TestLimitFields (0.00s)
=== RUN   TestOmitFields
test DEBUG: 2016/11/30 08:53:44 config.go:66: Loading log...
--- PASS: TestOmitFields (0.00s)
PASS
ok  	github.com/advantageous/systemd-cloud-watch/cloud-watch	10.017s
```




#### Building the docker image to build the linux instance to build this project

```sh
# from project root
cd bin.packer
bin.packer build packer_docker.json
```


#### To run docker dev image
```sh
# from project root
cd bin.packer
./run.sh

```

#### Building the ec2 image with bin.packer to build the linux instance to build this project

```sh
# from project root
cd bin.packer
bin.packer build packer_ec2.json
```

We use the [docker](https://www.packer.io/docs/builders/docker.html) support for [bin.packer](https://www.packer.io/).
("Packer is a tool for creating machine and container images for multiple platforms from a single source configuration.")

Use `ec2_env.sh_example` to create a `ec2_env.sh` with the instance id that was just created. 

#### ec2_env.sh_example
```
#!/usr/bin/env bash
export ami=ami-YOURAMI
export subnet=subnet-YOURSUBNET
export security_group=sg-YOURSG
export iam_profile=YOUR_IAM_ROLE
export key_name=MY_PEM_FILE_KEY_NAME

```

##### Using EC2 image (assumes you have ~/.ssh config setup)
```sh
# from project root
cd bin.packer

# Run and log into dev env running in EC2
./runEc2Dev.sh

# Log into running server
./loginIntoEc2Dev.sh

```





## Setting up a Linux env for testing/developing (CentOS7).
```sh
yum -y install wget
yum install -y git
yum install -y gcc
yum install -y systemd-devel


echo "installing go"
cd /tmp
wget https://storage.googleapis.com/golang/go1.7.3.linux-amd64.tar.gz
tar -C /usr/local/ -xzf go1.7.3.linux-amd64.tar.gz
rm go1.7.3.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bash_profile
```

## Setting up Java to write to systemd journal

#### gradle build
```
compile 'org.gnieh:logback-journal:0.2.0'

```

#### logback.xml
```xml
<?xml version="1.0" encoding="UTF-8"?>
<configuration>

    <appender name="journal" class="org.gnieh.logback.SystemdJournalAppender" />

    <root level="INFO">
        <appender-ref ref="journal" />
        <customFields>{"serviceName":"adfCalcBatch","serviceHost":"${HOST}"}</customFields>
    </root>


    <logger name="com.mycompany" level="INFO"/>

</configuration>
```

## Commands for controlling systemd service EC2 dev env

```sh
# Get status
sudo systemctl status journald-cloudwatch
# Stop Service
sudo systemctl stop journald-cloudwatch
# Find the service
ps -ef | grep cloud
# Run service manually
/usr/bin/systemd-cloud-watch_linux /etc/journald-cloudwatch.conf

```



## Derived
This is based on [advantageous journald-cloudwatch-logs](https://github.com/advantageous/journald-cloudwatch-logs)
which was forked from [saymedia journald-cloudwatch-logs](https://github.com/saymedia/journald-cloudwatch-logs).


## Status
Done and released. 



### Using as a lib.

You can use this project as a lib and you can pass in your own *JournalRepeater* and your own *Journal*.


#### Interface for JournalRepeater
```go
package cloud_watch


type Record struct {...} //see source code

type JournalRepeater interface {
	// Close closes a journal opened with NewJournal.
	Close() error;
	WriteBatch(records []Record) error;
}
```

#### Interface for Journal
```go
type Journal interface {
	// Close closes a journal opened with NewJournal.
	Close() error;

	// Next advances the read pointer into the journal by one entry.
	Next() (uint64, error);

	// NextSkip advances the read pointer by multiple entries at once,
	// as specified by the skip parameter.
	NextSkip(skip uint64) (uint64, error);

	// Previous sets the read pointer into the journal back by one entry.
	Previous() (uint64, error);

	// PreviousSkip sets back the read pointer by multiple entries at once,
	// as specified by the skip parameter.
	PreviousSkip(skip uint64) (uint64, error);

	// GetDataValue gets the data object associated with a specific field from the
	// current journal entry, returning only the value of the object.
	GetDataValue(field string) (string, error);


	// GetRealtimeUsec gets the realtime (wallclock) timestamp of the current
	// journal entry.
	GetRealtimeUsec() (uint64, error);

 	AddLogFilters(config *Config)

	// GetMonotonicUsec gets the monotonic timestamp of the current journal entry.
	GetMonotonicUsec() (uint64, error);

	// GetCursor gets the cursor of the current journal entry.
	GetCursor() (string, error);


	// SeekHead seeks to the beginning of the journal, i.e. the oldest available
	// entry.
	SeekHead() error;

	// SeekTail may be used to seek to the end of the journal, i.e. the most recent
	// available entry.
	SeekTail() error;

	// SeekCursor seeks to a concrete journal cursor.
	SeekCursor(cursor string) error;

	// Wait will synchronously wait until the journal gets changed. The maximum time
	// this call sleeps may be controlled with the timeout parameter.  If
	// sdjournal.IndefiniteWait is passed as the timeout parameter, Wait will
	// wait indefinitely for a journal change.
	Wait(timeout time.Duration) int;
}

```

#### Using as a lib
```go

package main

import (
	jcw  "github.com/advantageous/systemd-cloud-watch/cloud-watch"
	"flag"
	"os"
)

var help = flag.Bool("help", false, "set to true to show this help")

func main() {

	logger := jcw.NewSimpleLogger("main", nil)

	flag.Parse()

	if *help {
		usage(logger)
		os.Exit(0)
	}

	configFilename := flag.Arg(0)
	if configFilename == "" {
		usage(logger)
		panic("config file name must be set!")
	}

	config := jcw.CreateConfig(configFilename, logger)
	logger = jcw.NewSimpleLogger("main", config)  
	journal := jcw.CreateJournal(config, logger) //Instead of this, load your own journal
	repeater := jcw.CreateRepeater(config, logger) //Instead of this, load your own repeater

	jcw.RunWorkers(journal, repeater, logger, config )
}

func usage(logger *jcw.Logger) {
	logger.Error.Println("Usage: systemd-cloud-watch  <config-file>")
	flag.PrintDefaults()
}

```

You could for example create a *JournalRepeater* that writes to *InfluxDB* instead of *CloudWatch*.




Improvements:

* Added unit tests (there were none).
* Heavily reduced locking by using [qbit](https://github.com/advantageous/go-qbit) instead of original implementation.
* Added cross compile so I can develop/test on my laptop (MacOS).
* Made logging stateless. No more need for a state file.
* No more getting out of sync with CloudWatch.
* Detects being out of sync and recovers.
* Fixed error with log messages being too big.
* Added ability to include or omit logging fields.
* Created docker image and scripts to test on Linux (CentOS7).
* Created EC2 image and scripts to test on Linux running in AWS EC2 (CentOS7).
* Code organization (we use a package).
* Added comprehensive logging which includes debug logging by config.
* Uses actual timestamp from journal log record instead of just current time
* Auto-creates CloudWatch log group if it does not exist
* Allow this to be used as a library by providing interface for Journal and JournalWriter.


## License

The original work was from Say Media Inc. We had issues with it and did about a 90% rewrite.

All additional work is covered under Apache 2.0 license.
Copyright (c) 2016 Geoff Chandler, Rick Hightower


Copyright (c) 2015 Say Media Inc

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

