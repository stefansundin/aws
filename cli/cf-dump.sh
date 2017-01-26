#!/bin/bash -e
DIR="$1-$(date +%F)"
AWS_REGION=${2:-us-east-1}
mkdir -p "$DIR"
aws cloudformation describe-stacks --stack-name "$1" --region "$AWS_REGION" --query Stacks[0] > "$DIR/stack.json"
aws cloudformation get-template --stack-name "$1" --region "$AWS_REGION" --query TemplateBody > "$DIR/template.json"
if [[ "$(head -c1 "$DIR/template.json")" != "{" ]]; then
  rm "$DIR/template.json"
  aws cloudformation get-template --stack-name "$1" --region "$AWS_REGION" --query TemplateBody --output text > "$DIR/template.yml"
fi
aws cloudformation describe-stack-resources --stack-name "$1" --region "$AWS_REGION" > "$DIR/resources.json"
aws cloudformation describe-stack-events --stack-name "$1" --region "$AWS_REGION" > "$DIR/events.json"
set -x
ls -lh "$DIR"
