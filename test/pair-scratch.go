package test

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type VPC struct{}
type ECS struct{}

type Resource interface {
	DescribeAWSResourceInput()
}

func generateVpcInput(vpcName string) *ec2.DescribeVpcsInput {
	return &ec2.DescribeVpcsInput{
		VpcIds: []*string{
			aws.String(vpcName),
		},
	}
}

func atestGeneric(name string) {
	svc := ec2.New(session.New())
	result, err := svc.DescribeVpcs(generateVpcInput("test"))
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	fmt.Println(result)
}
