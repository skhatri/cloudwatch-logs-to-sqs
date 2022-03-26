
export ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
export LOG_GROUP_NAME="/aws/rds/cluster/dasdemo/postgresql"
export LOG_STREAM_NAME="dasdemo-instance-1.0"
export AWS_REGION="ap-southeast-2"
export TOPIC_NAME="cloudwatch-broadcast.fifo"
export QUEUE_NAME="cloudwatch-receiver.fifo"

aws iam create-role --role-name lambda-exec --assume-role-policy-document '{"Version": "2012-10-17","Statement": [{ "Effect": "Allow", "Principal": {"Service": "lambda.amazonaws.com"}, "Action": "sts:AssumeRole"}]}'

aws iam attach-role-policy --role-name lambda-exec --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole
aws iam attach-role-policy --role-name lambda-exec --policy-arn arn:aws:iam::aws:policy/AmazonSNSFullAccess

export SNS_TOPIC_ARN="arn:aws:sns:${AWS_REGION}:${ACCOUNT_ID}:${TOPIC_NAME}"
aws sns delete-topic --topic-arn ${SNS_TOPIC_ARN}
aws sns create-topic --name ${TOPIC_NAME} \
--attributes "{\"DisplayName\":\"cloudwatch broadcast\",\"FifoTopic\": \"true\", \"ContentBasedDeduplication\": \"false\"}"

export SQS_QUEUE_ARN="arn:aws:sqs:${AWS_REGION}:${ACCOUNT_ID}:${QUEUE_NAME}"
export SQS_QUEUE_URL="https://sqs.${AWS_REGION}.amazonaws.com/${ACCOUNT_ID}/${QUEUE_NAME}"

aws sqs delete-queue --queue-url ${SQS_QUEUE_URL}
#wait 60 seconds if queue-deleted
aws sqs create-queue --queue-name ${QUEUE_NAME} \
--attributes "{  \"ContentBasedDeduplication\":\"false\", \"FifoQueue\": \"true\", \"DelaySeconds\": \"0\", \"ReceiveMessageWaitTimeSeconds\": \"0\", \"VisibilityTimeout\": \"30\"}"

#ensure access policy for the queue is set

aws sns subscribe --topic-arn ${SNS_TOPIC_ARN} --protocol sqs \
--notification-endpoint ${SQS_QUEUE_ARN}

export curr=$(pwd)
cd go-lambda && go mod vendor && GOOS=linux CGO_ENABLED=0 go build printlogsgo.go && cd $curr
cp go-lambda/printlogsgo . && jar cfM printlogsgo.zip printlogsgo
rm printlogsgo




aws lambda delete-function --function-name printlogsgo

export role_arn=$(aws iam get-role --role-name lambda-exec|jq -r .Role.Arn)
aws lambda create-function \
--function-name printlogsgo \
--zip-file fileb://./printlogsgo.zip \
--role ${role_arn} \
--environment "Variables={SNS_TOPIC_ARN=${SNS_TOPIC_ARN}}" \
--handler printlogsgo \
--runtime go1.x

rm printlogsgo.zip

aws lambda add-permission \
--function-name "printlogsgo" \
--statement-id "printlogsgo" \
--principal "logs.${AWS_REGION}.amazonaws.com" \
--action "lambda:InvokeFunction" \
--source-arn "arn:aws:logs:${AWS_REGION}:${ACCOUNT_ID}:log-group:${LOG_GROUP_NAME}:*" \
--source-account "${ACCOUNT_ID}"

aws logs delete-subscription-filter --log-group-name ${LOG_GROUP_NAME} --filter-name cloudwatchlogslambdago

aws logs put-subscription-filter \
--log-group-name ${LOG_GROUP_NAME} \
--filter-name cloudwatchlogslambdago \
--filter-pattern "?connection ?execute ?disconnection" \
--destination-arn arn:aws:lambda:${AWS_REGION}:${ACCOUNT_ID}:function:printlogsgo

export now="$(date +%s)000" 
aws logs put-log-events --log-group-name ${LOG_GROUP_NAME} \
--log-stream-name ${LOG_STREAM_NAME} \
--log-events "[{\"timestamp\":${now} , \"message\": \"connection postgres demo ${now}\"}]"
