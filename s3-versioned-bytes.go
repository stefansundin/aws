package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/dustin/go-humanize"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please supply a bucket name.")
		os.Exit(1)
	}
	bucket := os.Args[1]

	sess := session.New()
	client := s3.New(sess)

	fmt.Printf("Getting bucket region... ")
	getBucketLocationResp, err2 := client.GetBucketLocation(&s3.GetBucketLocationInput{
		Bucket: &bucket,
	})
	if err2 != nil {
		fmt.Println(err2.Error())
		return
	}
	var region string
	if getBucketLocationResp.LocationConstraint == nil {
		region = "us-east-1"
	} else {
		region = *getBucketLocationResp.LocationConstraint
	}
	fmt.Printf("%s\n", region)
	cfg := aws.NewConfig().WithRegion(region)

	fmt.Println("Getting CloudWatch metric to estimate number of objects...")
	now := time.Now()
	oneDayAgo := time.Unix(now.Unix()-(60*60*24), 0)
	cwClient := cloudwatch.New(sess, cfg)
	resp, err := cwClient.GetMetricStatistics(&cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/S3"),
		MetricName: aws.String("NumberOfObjects"),
		Unit:       aws.String("Count"),
		StartTime:  aws.Time(oneDayAgo),
		EndTime:    aws.Time(now),
		Period:     aws.Int64(60 * 60),
		Statistics: []*string{
			aws.String("Sum"),
		},
		Dimensions: []*cloudwatch.Dimension{
			{
				Name:  aws.String("BucketName"),
				Value: aws.String(bucket),
			},
			{
				Name:  aws.String("StorageType"),
				Value: aws.String("AllStorageTypes"),
			},
		},
	})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	datapoint := resp.Datapoints[len(resp.Datapoints)-1]
	fmt.Printf("Number of objects: %d (measured %s on %s)\n", int64(*datapoint.Sum), humanize.Time(*datapoint.Timestamp), *datapoint.Timestamp)

	// list the bucket
	pageNum := 0
	numVersions := 0
	var oldBytes int64
	s3Client := s3.New(sess, cfg)
	// fmt.Println("Deleted objects:")
	err = s3Client.ListObjectVersionsPages(&s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket),
	},
		func(page *s3.ListObjectVersionsOutput, lastPage bool) bool {
			pageNum++
			numVersions += len(page.Versions)
			fmt.Printf("\rListing bucket page %d (%d objects)... %s in previous versions so far.", pageNum, numVersions, humanize.Bytes(uint64(oldBytes)))

			// TODO: Split out bytes based on storage type as they have different costs!
			for _, obj := range page.Versions {
				if *obj.IsLatest {
					continue
				}
				// fmt.Printf("- %s: %s (%d bytes)\n", *obj.Key, humanize.Bytes(uint64(*obj.Size)), *obj.Size)
				oldBytes += *obj.Size
			}
			return true
		})
	fmt.Printf("\n")

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Printf("Number of versions: %d\n", numVersions)

	// summary
	fmt.Println()
	fmt.Printf("Deleted %s (%d bytes)\n", humanize.Bytes(uint64(oldBytes)), oldBytes)
	fmt.Println("Costs:")
	fmt.Printf("- $%f / month\n", float64(oldBytes)/1000000000.0*0.0300)
	fmt.Printf("- $%f / hour\n", float64(oldBytes)/1000000000.0*0.0300/(24*30))
}
