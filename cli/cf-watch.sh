#!/bin/bash -e
aws cloudformation describe-stack-events --stack-name $1 --region ${2:-us-east-1} --profile ${3:-default} --max-items 20 --query 'StackEvents[*].[Timestamp, ResourceStatus, LogicalResourceId, ResourceStatusReason]' --output text | column -t -s $'\t' | cut -c -$COLUMNS
