package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"os"
	"time"
)

func main() {
	cwSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	queueName := os.Getenv("SQS_QUEUE_NAME")

	sqsService := sqs.New(cwSession)
	queueUrlOut, err := sqsService.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})
	if err != nil {
		fmt.Println("error retrieving queue url", err)
		return
	}
	queueUrl := queueUrlOut.QueueUrl
	waitTime := int64(10)
	for i := 0; ; i++ {
		fmt.Println("queue", *queueUrl)
		read := readMessages(sqsService, queueUrl, waitTime)
		fmt.Printf("Attempt %d read %d messages\n", i, read)
		time.Sleep(time.Duration(5) * time.Second)
	}

}

func readMessages(sqsService *sqs.SQS, queueUrl *string, waitTime int64) int {
	out, receiveErr := sqsService.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            queueUrl,
		MaxNumberOfMessages: aws.Int64(10),
		AttributeNames: aws.StringSlice([]string{
			"SentTimestamp",
		}),
		MessageAttributeNames: aws.StringSlice([]string{
			"All",
		}),
		WaitTimeSeconds: &waitTime,
	})
	if receiveErr != nil {
		fmt.Println("queue read error", receiveErr)
		return 0
	}
	read := len(out.Messages)
	for _, msg := range out.Messages {
		fmt.Println("READ", msg.String())
		delOut, delEr := sqsService.DeleteMessage(&sqs.DeleteMessageInput{
			QueueUrl:      queueUrl,
			ReceiptHandle: msg.ReceiptHandle,
		})
		if delEr != nil {
			fmt.Println("delete message err", delEr)
		} else {
			fmt.Println("deleted message", delOut.String())
		}
	}

	return read
}
