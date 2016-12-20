package cloud_watch

import (
	"github.com/aws/aws-sdk-go/aws"
	awsCredentials "github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"os"
	"strings"
	lg "github.com/advantageous/go-logback/logging"
)

var awsLogger = lg.NewSimpleLogger("aws")

func NewAWSSession(cfg *Config) *awsSession.Session {

	metaDataClient, session := getClient(cfg)
	credentials := getCredentials(metaDataClient)

	if credentials != nil {
		awsConfig := &aws.Config{
			Credentials: getCredentials(metaDataClient),
			Region:      aws.String(getRegion(metaDataClient, cfg, session)),
			MaxRetries:  aws.Int(3),
		}
		return awsSession.New(awsConfig)
	} else {
		return awsSession.New(&aws.Config{
			Region:     aws.String(getRegion(metaDataClient, cfg, session)),
			MaxRetries: aws.Int(3),
		})
	}

}

func getClient(config *Config) (*ec2metadata.EC2Metadata, *awsSession.Session) {
	if !config.Local {
		awsLogger.Debug("Config NOT set to local using meta-data client to find local")
		var session = awsSession.New(&aws.Config{})
		return ec2metadata.New(session), session
	} else {
		awsLogger.Info("Config set to local")
		return nil, nil
	}
}

func getRegion(client *ec2metadata.EC2Metadata, config *Config, session *awsSession.Session) string {

	if client == nil {
		awsLogger.Info("Client missing using config to set region")
		if config.AWSRegion == "" {
			awsLogger.Info("AWSRegion missing using default region us-west-2")
			return "us-west-2"
		} else {
			return config.AWSRegion
		}
	} else {
		region, err := client.Region()
		if err != nil {
			awsLogger.Errorf("Unable to get region from aws meta client : %s %v", err.Error(), err)
			os.Exit(3)
		}

		config.AWSRegion = region
		config.EC2InstanceId, err = client.GetMetadata("instance-id")
		if err != nil {
			awsLogger.Errorf("Unable to get instance id from aws meta client : %s %v", err.Error(), err)
			os.Exit(4)
		}

		if config.LogStreamName == "" {
			var az, name, ip string
			az = findAZ(client)
			ip = findLocalIp(client)

			name = findInstanceName(config.EC2InstanceId, config.AWSRegion, session)
			config.LogStreamName = name + "-" + strings.Replace(ip, ".", "-", -1) + "-" + az
			awsLogger.Infof("LogStreamName was not set so using %s \n", config.LogStreamName)
		}

		return region
	}

}
func findLocalIp(metaClient *ec2metadata.EC2Metadata) string {
	ip, err := metaClient.GetMetadata("local-ipv4")

	if err != nil {
		awsLogger.Errorf("Unable to get private ip address from aws meta client : %s %v", err.Error(), err)
		os.Exit(6)
	}

	return ip

}

func getCredentials(client *ec2metadata.EC2Metadata) *awsCredentials.Credentials {

	if client == nil {
		awsLogger.Infof("Client missing credentials not looked up")
		return nil
	} else {
		return awsCredentials.NewChainCredentials([]awsCredentials.Provider{
			&awsCredentials.EnvProvider{},
			&ec2rolecreds.EC2RoleProvider{
				Client: client,
			},
		})
	}

}

func findAZ(metaClient *ec2metadata.EC2Metadata) string {

	az, err := metaClient.GetMetadata("placement/availability-zone")

	if err != nil {
		awsLogger.Errorf("Unable to get az from aws meta client : %s %v", err.Error(), err)
		os.Exit(5)
	}

	return az
}

func findInstanceName(instanceId string, region string, session *awsSession.Session) string {

	var name = "NO_NAME"
	var err error

	ec2Service := ec2.New(session, aws.NewConfig().WithRegion(region))

	params := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{
			aws.String(instanceId), // Required
			// More values...
		},
	}

	resp, err := ec2Service.DescribeInstances(params)

	if err != nil {
		awsLogger.Errorf("Unable to get instance name tag DescribeInstances failed : %s %v", err.Error(), err)
		return name
	}

	if len(resp.Reservations) > 0 && len(resp.Reservations[0].Instances) > 0 {
		var instance = resp.Reservations[0].Instances[0]
		if len(instance.Tags) > 0 {

			for _, tag := range instance.Tags {
				if *tag.Key == "Name" {
					return *tag.Value
				}
			}
		}
		awsLogger.Errorf("Unable to get find name tag ")
		return name

	} else {
		awsLogger.Errorf("Unable to get find name tag ")
		return name
	}
}
