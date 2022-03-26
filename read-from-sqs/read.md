
### publish test
export TOPIC_NAME="cloudwatch-broadcast.fifo"
export QUEUE_NAME="cloudwatch-receiver.fifo"

export ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
export AWS_REGION="ap-southeast-2"
export SNS_TOPIC_ARN="arn:aws:sns:${AWS_REGION}:${ACCOUNT_ID}:${TOPIC_NAME}"
export SQS_QUEUE_NAME=${QUEUE_NAME}
export SQS_QUEUE_ARN="arn:aws:sqs:${AWS_REGION}:${ACCOUNT_ID}:${QUEUE_NAME}"

echo publish sample messages
go run publish.go


### read
export QUEUE_NAME="cloudwatch-receiver.fifo"
export SQS_QUEUE_NAME=${QUEUE_NAME}
go run read.go
