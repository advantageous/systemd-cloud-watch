package cloud_watch

import (
	awsCredentials "github.com/aws/aws-sdk-go/aws/credentials"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"os"
)

var logger = InitSimpleLog("aws", nil)


func NewAWSSession(cfg *Config) *awsSession.Session {

	metaDataClient := getClient(cfg)
	credentials := getCredentials(metaDataClient)

	if credentials!=nil {
		awsConfig := &aws.Config{
			Credentials: getCredentials(metaDataClient),
			Region:      aws.String(getRegion(metaDataClient, cfg)),
			MaxRetries:  aws.Int(3),
		}
		return awsSession.New(awsConfig)
	} else {
		return awsSession.New(&aws.Config{
			Region:      aws.String(getRegion(metaDataClient, cfg)),
			MaxRetries:  aws.Int(3),
		})
	}

}

func getClient(config *Config) *ec2metadata.EC2Metadata {
	if !config.Local {
		logger.Debug.Println("Config NOT set to local using meta-data client to find local")
		var session = awsSession.New(&aws.Config{})
		return  ec2metadata.New(session)
	} else {
		logger.Info.Println("Config set to local")
		return nil
	}
}

func getRegion(client *ec2metadata.EC2Metadata, config *Config) string {

	if client == nil {
		logger.Info.Println("Client missing using config to set region")
		if config.AWSRegion == "" {
			logger.Info.Println("AWSRegion missing using default region us-west-2")
			return "us-west-2"
		} else {
			return config.AWSRegion
		}
	} else {
		region, err := client.Region()
		if err != nil {
			logger.Error.Printf("Unable to get region from aws meta client %s", err)
			os.Exit(3)
		}
		return region
	}

}

func getCredentials(client *ec2metadata.EC2Metadata) *awsCredentials.Credentials {


	if client == nil {
		logger.Info.Printf("Client missing credentials not looked up")
		return nil
	} else {
		return awsCredentials.NewChainCredentials([]awsCredentials.Provider{
			&awsCredentials.EnvProvider{},
			&ec2rolecreds.EC2RoleProvider{
				Client:client,
			},
		})
	}

}

