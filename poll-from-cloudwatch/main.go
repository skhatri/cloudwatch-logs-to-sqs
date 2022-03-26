package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	cw "github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"log"
	"time"
)

func main() {
	cwSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	logs := cw.New(cwSession)
	logGroupName := "/aws/rds/cluster/dasdemo/postgresql"
	logStreamName := "dasdemo-instance-1.0"

	windowSize := time.Duration(-20) * time.Minute
	startTime := time.Now().Add(windowSize).UnixMilli()
	for i := 1; ; i++ {
		endTime := time.Now().UnixMilli()
		go retrieveLogs(logs, logGroupName, logStreamName, startTime, endTime, i)
		startTime = endTime + 1
		time.Sleep(time.Duration(2) * time.Minute)
	}
}

func retrieveLogs(logs *cw.CloudWatchLogs, logGroupName string, logStreamName string, startTime int64, endTime int64, cnt int) {

	filterPattern := "?connection ?execute ?disconnection"
	err := logs.FilterLogEventsPages(&cw.FilterLogEventsInput{
		LogGroupName:   &logGroupName,
		EndTime:        &endTime,
		StartTime:      &startTime,
		LogStreamNames: []*string{&logStreamName},
		FilterPattern:  &filterPattern,
	}, func(filtered *cw.FilterLogEventsOutput, fin bool) bool {
		fmt.Println("printing data ----")
		for _, e := range filtered.Events {
			fmt.Println("Filter", cnt, "", *e.Timestamp, e.String())
		}
		return fin
	})

	if err != nil {
		log.Println("error", err)
	}
}
