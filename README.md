## Forwarding Postgres Logs

Logs of interest are forwarded to Postgres using a custom RDS parameters settings.


```mermaid
flowchart LR
    style postgres fill:#fa0987,stroke:#fa0987,stroke-width:4px
    style subscription fill:#3faa33,stroke:#3faa33,stroke-width:4px

    style go-lambda:lambda fill:#3faa33,stroke:#3faa33,stroke-width:4px
    style sns fill:#3faa33,stroke:#3faa33,stroke-width:4px
    style sqs fill:#3faa33,stroke:#3faa33,stroke-width:4px
    style read-from-sqs:audit-handler fill:#3faa33,stroke:#3faa33,stroke-width:4px


    style poll-from-cloudwatch fill:#0f7af3,stroke:#0f7af3,stroke-width:4px
    style api fill:#0f7af3,stroke:#0f7af3,stroke-width:4px

    
    style cloudwatch fill:#fa0987,stroke:#fa0987,stroke-width:4px        
    
    subgraph Aurora
      postgres
    end
    
    subgraph Cloudwatch
      postgres-->cloudwatch
    end
    
    subgraph Cloudwatch Integration
      cloudwatch-..->alarm
      cloudwatch-..->events
      cloudwatch-..->metrics
      cloudwatch-->api
      cloudwatch-->subscription
    end
    
    subgraph Subscription With Filter
      subscription-->kinesis
      subscription-->kinesis-FH
      subscription-->go-lambda:lambda
    end
    
    subgraph SNS
      go-lambda:lambda-->sns
      sns-->sqs
      sns-..->http
      sns-..->email
      sns-..->notification:lambda
      sns-..->sms
      sns-..->notification:kinesis-FH
    end
    
    
    subgraph Client
      api-->poll-from-cloudwatch
      sqs-->read-from-sqs:audit-handler
    end
    
```


### project: poll-from-cloudwatch

Read from CloudWatch using Filter endpoint. This is a polling exercise. This can be quick way of reading cloudwatch from
outside EC2 environment

Filter performs a filter for connection, execute and disconnection events

### project: go-lambda

Lambda that forwards messages to a SNS topic. Lambda performs a filter for connection, execute and disconnection events

#### setup lamdda

Refer to lambda.md to setup lambda

#### project: read-from-sqs

To read SQS messages and delete upon receipt. read.md has instructions to publish locally as well as read locally.
