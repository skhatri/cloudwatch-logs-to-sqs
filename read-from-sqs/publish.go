package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/google/uuid"
	"os"
	"strings"
)

func main() {
	cwSession := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	accountId := os.Getenv("ACCOUNT_ID")
	topicArn := os.Getenv("SNS_TOPIC_ARN")
	queueName := os.Getenv("SQS_QUEUE_NAME")
	awsRegion := os.Getenv("AWS_REGION")
	queueArn := fmt.Sprintf("arn:aws:sqs:%s:%s:%s", awsRegion, accountId, queueName)

	snsService := sns.New(cwSession)

	items, listEr := snsService.ListSubscriptions(&sns.ListSubscriptionsInput{
	})
	if listEr != nil {
		fmt.Println("could not list subscriptions", listEr)
	}
	hasSubscription := false
	for _, item := range items.Subscriptions {
		fmt.Println("subscription", *item.TopicArn, *item.SubscriptionArn, *item.Endpoint)
		if !hasSubscription {
			hasSubscription = topicArn == *item.TopicArn && strings.Contains(*item.Endpoint, queueName)
		}
	}
	if !hasSubscription {
		fmt.Println("needs subscription")
		sout, serr := snsService.Subscribe(&sns.SubscribeInput{
			TopicArn:              &topicArn,
			Protocol:              aws.String("sqs"),
			Endpoint:              &queueArn,
			ReturnSubscriptionArn: aws.Bool(true),
		})
		if serr != nil {
			fmt.Println("subscription error", serr)
			return
		}
		fmt.Println("subscription", *sout.SubscriptionArn)
	}
	for i := 0; i < 10; i++ {
		msg := "hello there from sns" + uuid.New().String()
		pout, perr := snsService.Publish(&sns.PublishInput{
			Message:  aws.String(msg),
			TopicArn: &topicArn,
			Subject:  aws.String("hello" + uuid.New().String()),
			MessageDeduplicationId: aws.String(uuid.New().String()),
			MessageGroupId: aws.String("postgres"),
			MessageAttributes: map[string]*sns.MessageAttributeValue{
				"text": &sns.MessageAttributeValue{
					DataType: aws.String("String"),
					StringValue: aws.String(msg),
				},
			},
		})
		if perr != nil {
			fmt.Println("publish err", perr)
		} else {
			fmt.Println("published", *pout.MessageId)
		}
	}

}
