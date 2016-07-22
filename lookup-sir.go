package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please supply a sir id.")
		os.Exit(1)
	}
	sir := os.Args[1]
	if !regexp.MustCompile("sir-[0-9a-z]{8}").MatchString(sir) {
		fmt.Println("Supplied argument does not look like a valid sir id. Sorry.")
		os.Exit(1)
	}

	asClient := autoscaling.New(session.New())
	var groups []string
	err := asClient.DescribeAutoScalingGroupsPages(&autoscaling.DescribeAutoScalingGroupsInput{},
		func(page *autoscaling.DescribeAutoScalingGroupsOutput, lastPage bool) bool {
			for _, asg := range page.AutoScalingGroups {
				if int64(len(asg.Instances)) >= *asg.DesiredCapacity {
					continue
				}
				groups = append(groups, *asg.AutoScalingGroupName)
			}
			return true
		})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("Suspected autoscaling groups:")
	fmt.Println(groups)
	fmt.Println()

	for _, group := range groups {
		resp, err := asClient.DescribeScalingActivities(&autoscaling.DescribeScalingActivitiesInput{
			AutoScalingGroupName: aws.String(group),
			MaxRecords:           aws.Int64(4),
		})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		for _, activity := range resp.Activities {
			if strings.Contains(*activity.Description, sir) {
				fmt.Printf("Found sir in group: %s\n", group)
				fmt.Printf("Direct URL: https://console.aws.amazon.com/ec2/autoscaling/home?region=us-east-1#AutoScalingGroups:id=%s;view=history\n", group)
				os.Exit(0)
			}
		}
	}
	fmt.Printf("Could not find %s in any autoscaling groups. :/\n", sir)
}
