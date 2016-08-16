package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
	var listBucketsResp *s3.ListBucketsOutput
	var err error
	if len(os.Args) > 1 {
		listBucketsResp = &s3.ListBucketsOutput{
			Buckets: []*s3.Bucket{},
		}
		for _, arg := range os.Args[1:] {
			listBucketsResp.Buckets = append(listBucketsResp.Buckets, &s3.Bucket{
				Name: aws.String(arg),
			})
		}
	}

	sess := session.New()
	client := s3.New(sess)

	// list buckets if no args are provided
	if listBucketsResp == nil {
		listBucketsResp, err = client.ListBuckets(&s3.ListBucketsInput{})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
	}

	for _, bucket := range listBucketsResp.Buckets {
		fmt.Println(*bucket.Name)
		var err2 error

		getBucketLocationResp, err2 := client.GetBucketLocation(&s3.GetBucketLocationInput{
			Bucket: bucket.Name,
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
		cfg := aws.NewConfig().WithRegion(region)
		svc := s3.New(sess, cfg)

		getBucketVersioningResp, err2 := svc.GetBucketVersioning(&s3.GetBucketVersioningInput{
			Bucket: bucket.Name,
		})
		if err2 != nil {
			fmt.Println(err2.Error())
			return
		}
		if getBucketVersioningResp.Status == nil {
			getBucketVersioningResp.Status = aws.String("Not enabled")
		}
		if *getBucketVersioningResp.Status != "Enabled" {
			fmt.Printf("- \033[31;1mVersioning: %s\033[0m\n", *getBucketVersioningResp.Status)
			// fmt.Printf("aws s3api put-bucket-versioning --bucket %s --versioning-configuration Status=Enabled\n", *bucket.Name)
		}

		getBucketLifecycleConfigurationResp, err2 := svc.GetBucketLifecycleConfiguration(&s3.GetBucketLifecycleConfigurationInput{
			Bucket: bucket.Name,
		})
		if err2 != nil {
			fmt.Println("- \033[31;1mLifecycle rules not configured!\033[0m")
			// fmt.Printf("aws s3api put-bucket-lifecycle-configuration --bucket %s --lifecycle-configuration '{ \"Rules\": [ { \"ID\": \"Rule for the Entire Bucket\", \"Prefix\": \"\", \"Status\": \"Enabled\", \"NoncurrentVersionExpiration\": {\"NoncurrentDays\":7}, \"AbortIncompleteMultipartUpload\": {\"DaysAfterInitiation\":7} } ] }'\n", *bucket.Name)
		}
		for _, rule := range getBucketLifecycleConfigurationResp.Rules {
			if *rule.Status != "Enabled" {
				fmt.Println("- \033[31;1mLfecycle rule disabled!\033[0m")
				continue
			}
			if rule.AbortIncompleteMultipartUpload == nil ||
				*rule.AbortIncompleteMultipartUpload.DaysAfterInitiation != int64(7) ||
				rule.NoncurrentVersionExpiration == nil ||
				*rule.NoncurrentVersionExpiration.NoncurrentDays != int64(7) {
				fmt.Println("- \033[33;1mNon-standard lifecycle rule!\033[0m")
			}
			if rule.Expiration != nil {
				fmt.Printf("- \033[41mDanger: Objects expire after %d days!\033[0m\n", *rule.Expiration.Days)
			}
		}
	}
}
