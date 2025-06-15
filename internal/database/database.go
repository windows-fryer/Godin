package database

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/golang/glog"
)

var dynamoClient *dynamodb.Client

// getConfig loads the default AWS SDK configuration.
func getConfig() (*aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Start() {
	cfg, err := getConfig()

	if err != nil {
		glog.Fatalf("Unable to load SDK config, %v", err)
	}

	dynamoClient = dynamodb.NewFromConfig(*cfg)

	if _, err := GetService(&ServiceItemSearch{
		ServiceID:   "936ed3ef-326f-44e9-9ba3-453d68997c90",
		ServiceCode: "iris",
	}); err != nil {
		glog.Fatalf("Failed to get service: %v", err)
	}

}

func Stop() {}
