package database

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/golang/glog"
)

type ServiceItemSearch struct {
	ServiceID   string `dynamodbav:"ServiceID"`
	ServiceCode string `dynamodbav:"ServiceCode"`
}

type ServiceItem struct {
	ServiceID   string `dynamodbav:"ServiceID"`
	ServiceCode string `dynamodbav:"ServiceCode"`
}

func CreateService(service *ServiceItem) error {
	marshaledItem, err := attributevalue.MarshalMap(service)

	if err != nil {
		return err
	}

	_, err = dynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String("GodinVFS-Services"),

		Item: marshaledItem,
	})

	if err != nil {
		return err
	}

	glog.V(2).Infof("Created service item: %v", service)

	return nil
}

func GetService(service *ServiceItemSearch) (*ServiceItem, error) {
	marshaledItem, err := attributevalue.MarshalMap(service)

	if err != nil {
		return nil, err
	}

	result, err := dynamoClient.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String("GodinVFS-Services"),

		Key: marshaledItem,
	})

	if err != nil {
		return nil, err
	}

	if len(result.Item) == 0 {
		return nil, fmt.Errorf("service not found: %v", service)
	}

	var serviceItem ServiceItem

	err = attributevalue.UnmarshalMap(result.Item, &serviceItem)

	if err != nil {
		return nil, err
	}

	glog.V(2).Infof("Retrieved service item: %v", serviceItem)

	return &serviceItem, nil
}
