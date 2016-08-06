package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

func main() {
	var logGroupName string
	var start time.Duration
	flag.DurationVar(&start, "start", time.Duration(10*time.Minute), "How far back to look for logs")
	flag.Parse()
	if flag.NArg() > 0 {
		logGroupName = flag.Arg(0)
	} else {
		logGroupName = "/aws/lambda/default-log-group"
	}

	svc := cloudwatchlogs.New(session.New())

	params := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName: aws.String(logGroupName),
		Interleaved:  aws.Bool(true),
		StartTime:    aws.Int64(time.Now().UTC().Add(-start).Unix() * 1000),
	}

	waiting := false
	for {
		err := svc.FilterLogEventsPages(params,
			func(page *cloudwatchlogs.FilterLogEventsOutput, lastPage bool) bool {
				if len(page.Events) == 0 {
					fmt.Print(".")
					waiting = true
				} else {
					lastEvent := page.Events[len(page.Events)-1]
					params.StartTime = aws.Int64(*lastEvent.Timestamp + 1)
					if waiting {
						fmt.Print("\r")
						waiting = false
					}
				}
				for _, event := range page.Events {
					fmt.Printf(*event.Message)
				}
				return true
			})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		time.Sleep(time.Second)
	}
}
