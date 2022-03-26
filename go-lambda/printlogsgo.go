package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/google/uuid"
	"io/ioutil"
	"os"
	"strings"
)

/*
Sample Data from Lambda Test console:

{
  "awslogs": {
    "data": "H4sIAAAAAAAAAHWPwQqCQBCGX0Xm7EFtK+smZBEUgXoLCdMhFtKV3akI8d0bLYmibvPPN3wz00CJxmQnTO41whwWQRIctmEcB6sQbFC3CjW3XW8kxpOpP+OC22d1Wml1qZkQGtoMsScxaczKN3plG8zlaHIta5KqWsozoTYw3/djzwhpLwivWFGHGpAFe7DL68JlBUk+l7KSN7tCOEJ4M3/qOI49vMHj+zCKdlFqLaU2ZHV2a4Ct/an0/ivdX8oYc1UVX860fQDQiMdxRQEAAA=="
  }
}
*/
type Input struct {
	AwsLogs AwsLogs `json:"awslogs"`
}

type AwsLogs struct {
	Data string `json:"data"`
}

/*
{"messageType":"DATA_MESSAGE","owner":"123456789123",
"logGroup":"testLogGroup","logStream":"testLogStream",
"subscriptionFilters":["testFilter"],
"logEvents":[{"id":"eventId1","timestamp":1440442987000,"message":"[ERROR] First test message"},{"id":"eventId2","timestamp":1440442987001,"message":"[ERROR] Second test message"}]}
*/
type CloudwatchLogs struct {
	MessageType         string      `json:"messageType"`
	Owner               string      `json:"owner"`
	LogGroup            string      `json:"logGroup"`
	SubscriptionFilters []string    `json:"subscriptionFilters"`
	LogStream           string      `json:"logStream"`
	LogEvents           []LogEvents `json:"logEvents"`
}

type LogEvents struct {
	Id        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
	Message   string `json:"message"`
}

func HandleRequest(ctx context.Context, input Input) (string, error) {
	gzbytes, err := base64.StdEncoding.DecodeString(input.AwsLogs.Data)
	if err != nil {
		return "", err
	}
	gzreader, gerr := gzip.NewReader(bytes.NewBuffer(gzbytes))
	if gerr != nil {
		return "", gerr
	}
	defer gzreader.Close()
	data, rerr := ioutil.ReadAll(gzreader)
	if rerr != nil {
		return "", rerr
	}
	cloudwatchLogs := CloudwatchLogs{}
	jerr := json.NewDecoder(bytes.NewBuffer(data)).Decode(&cloudwatchLogs)
	if jerr != nil {
		return "", jerr
	}
	size := len(cloudwatchLogs.LogEvents)
	message := fmt.Sprintf("to process %d", size)
	publishInputs := make([]*sns.PublishInput, 0)
	snsTopicArn := os.Getenv("SNS_TOPIC_ARN")
	for _, logEvt := range cloudwatchLogs.LogEvents {
		evt := &logEvt
		subject := "execute"
		if strings.Contains(evt.Message, " connection ") {
			subject = "login"
		} else if strings.Contains(evt.Message, " disconnection ") {
			subject = "logout"
		}

		publishInputs = append(publishInputs, &sns.PublishInput{
			TopicArn:               &snsTopicArn,
			Message:                &evt.Message,
			Subject:                aws.String(subject + uuid.New().String()),
			MessageDeduplicationId: aws.String(uuid.New().String()),
			MessageGroupId:         aws.String("postgres" + uuid.New().String()),
		})
	}

	if snsTopicArn != "" {
		sess := session.Must(session.NewSessionWithOptions(session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}))
		snsService := sns.New(sess)
		fail := 0
		successful := 0
		for _, pI := range publishInputs {
			_, pe := snsService.Publish(pI)
			if pe != nil {
				fail++
				fmt.Println("error in publish", pe)
			} else {
				successful++
			}
		}
		message = fmt.Sprintf("failed: %d, successful: %d out of %d", fail, successful, size)
	} else {
		fmt.Println("topic not set")
	}
	return message, nil
}

func main() {

	if l := os.Getenv("AWS_EXECUTION_ENV"); l != "" {
		lambda.Start(HandleRequest)
	} else {
		localTest()
	}
}

func localTest() {
	ms, e := HandleRequest(context.Background(), Input{
		AwsLogs: AwsLogs{
			Data: "H4sIAAAAAAAAAHWPwQqCQBCGX0Xm7EFtK+smZBEUgXoLCdMhFtKV3akI8d0bLYmibvPPN3wz00CJxmQnTO41whwWQRIctmEcB6sQbFC3CjW3XW8kxpOpP+OC22d1Wml1qZkQGtoMsScxaczKN3plG8zlaHIta5KqWsozoTYw3/djzwhpLwivWFGHGpAFe7DL68JlBUk+l7KSN7tCOEJ4M3/qOI49vMHj+zCKdlFqLaU2ZHV2a4Ct/an0/ivdX8oYc1UVX860fQDQiMdxRQEAAA==",
		},
	})
	if e != nil {
		fmt.Println("error", e)
	} else {
		fmt.Println(ms)
	}
}
