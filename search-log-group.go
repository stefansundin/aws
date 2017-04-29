package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

func formatURL(awsRegion string, logGroupName string, logStreamName string) string {
	domain := "console.aws.amazon.com"
	if awsRegion != "us-east-1" {
		domain = awsRegion + "." + domain
	}
	return fmt.Sprintf("https://%s/cloudwatch/home?region=%s#logEventViewer:group=%s;stream=%s", domain, awsRegion, logGroupName, url.QueryEscape(logStreamName))
}

func main() {
	var logGroupName string
	var start time.Duration
	flag.DurationVar(&start, "start", time.Duration(10*time.Minute), "How far back to look for logs")
	flag.Parse()
	if flag.NArg() < 2 {
		fmt.Println("Requires two arguments, log group and search term.")
		flag.Usage()
		os.Exit(1)
	}
	logGroupName = flag.Arg(0)
	searchTerm := flag.Arg(1)
	fmt.Printf("Filter pattern: %s\n", searchTerm)

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		fmt.Println("Please set region with AWS_REGION.")
	}

	_, offset := time.Now().Zone()
	hours := offset / 60 / 60

	// call aws
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := cloudwatchlogs.New(sess)

	params := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName:  aws.String(logGroupName),
		Interleaved:   aws.Bool(true),
		FilterPattern: aws.String(searchTerm),
		StartTime:     aws.Int64(time.Now().UTC().Add(-start).Unix() * 1000),
		// EndTime:    aws.Int64(time.Now().UTC().Add(-start).Unix() * 1000),
	}

	err := svc.FilterLogEventsPages(params,
		func(page *cloudwatchlogs.FilterLogEventsOutput, lastPage bool) bool {
			for _, event := range page.Events {
				// fmt.Println(event)
				s := *event.Message
				// if !strings.Contains(s, searchTerm) {
				// 	continue
				// }
				t := time.Unix(*event.IngestionTime/1000, *event.IngestionTime%1000)
				fmt.Println(formatURL(awsRegion, logGroupName, *event.LogStreamName))
				fmt.Printf("%s (UTC %+03d:00) | %s", t.Format("2006-01-02 15:04:05"), hours, s)
				if s[len(s)-1] != '\n' {
					fmt.Println()
				}
				fmt.Println()
			}
			return true
		})
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
