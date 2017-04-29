package main

// https://aws.amazon.com/ebs/pricing/
// https://aws.amazon.com/ebs/previous-generation/

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func getName(tags []*ec2.Tag) *string {
	for _, tag := range tags {
		if *tag.Key == "Name" {
			return tag.Value
		}
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please supply a region.")
		os.Exit(1)
	}
	region := os.Args[1]

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config:            aws.Config{Region: aws.String(region)},
	}))
	client := ec2.New(sess)

	pricing := map[string]float64{
		"standard": 0.05,
		"gp2":      0.10,
		"io1":      0.125,
		"st1":      0.045,
		"sc1":      0.025,
	}

	usage := map[string]int64{
		"standard": 0,
		"gp2":      0,
		"io1":      0,
		"st1":      0,
		"sc1":      0,
	}

	stoppedUsage := map[string]int64{
		"standard": 0,
		"gp2":      0,
		"io1":      0,
		"st1":      0,
		"sc1":      0,
	}

	err := client.DescribeVolumesPages(&ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("status"),
				Values: []*string{
					aws.String("available"),
				},
			},
		},
	},
		func(page *ec2.DescribeVolumesOutput, lastPage bool) bool {
			for _, vol := range page.Volumes {
				usage[*vol.VolumeType] += *vol.Size
			}
			return true
		})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println("EBS volumes not attached to an instance:")
	for volumeType, volumeUsage := range usage {
		if volumeUsage == 0 {
			continue
		}
		fmt.Printf("%8s: using %4d GB costing $%.2f per month\n", volumeType, volumeUsage, float64(volumeUsage)*pricing[volumeType])
	}
	fmt.Println()

	fmt.Println("EBS volumes associated with stopped instances:")
	err = client.DescribeVolumesPages(
		&ec2.DescribeVolumesInput{
			MaxResults: aws.Int64(100), // DescribeInstanceStatus() below takes max 100 items
			Filters: []*ec2.Filter{
				{
					Name: aws.String("status"),
					Values: []*string{
						aws.String("in-use"),
					},
				},
			},
		},
		func(page *ec2.DescribeVolumesOutput, lastPage bool) bool {
			var instancesIds []*string
			for _, vol := range page.Volumes {
				if len(vol.Attachments) > 1 {
					panic("more than one attachment, how is this possible?")
				}
				for _, instance := range vol.Attachments {
					instancesIds = append(instancesIds, instance.InstanceId)
				}
			}

			// make a call to check what instances are stopped
			resp, err2 := client.DescribeInstances(&ec2.DescribeInstancesInput{
				InstanceIds: instancesIds,
				Filters: []*ec2.Filter{
					{
						Name: aws.String("instance-state-code"),
						Values: []*string{
							aws.String("80"), // 80 = stopped
						},
					},
				},
			})
			if err2 != nil {
				fmt.Println(err2.Error())
				os.Exit(1)
			}

			for _, reservations := range resp.Reservations {
				for _, instance := range reservations.Instances {
					for _, vol := range page.Volumes {
						if *instance.InstanceId == *vol.Attachments[0].InstanceId {
							fmt.Printf("- %-25s (%s) using %4d GB of %s costing $%.2f per month\n", *getName(instance.Tags), *instance.InstanceId, *vol.Size, *vol.VolumeType, float64(*vol.Size)*pricing[*vol.VolumeType])
							stoppedUsage[*vol.VolumeType] += *vol.Size
							break
						}
					}
				}
			}

			return true
		})
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println()
	fmt.Println("Summary:")
	for volumeType, volumeUsage := range stoppedUsage {
		if volumeUsage == 0 {
			continue
		}
		fmt.Printf("%8s: using %4d GB costing $%.2f per month\n", volumeType, volumeUsage, float64(volumeUsage)*pricing[volumeType])
	}
}
