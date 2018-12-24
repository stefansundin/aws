#!/bin/bash -e
DIR="$1-$(date +%F)"
export AWS_DEFAULT_REGION=${2:-us-east-1}
export AWS_DEFAULT_PROFILE=${3:-default}
mkdir -p "$DIR"
aws cloudformation describe-stacks --stack-name "$1" --query Stacks[0] > "$DIR/stack.json"
aws cloudformation get-template --stack-name "$1" --query TemplateBody > "$DIR/template.json"
if [[ "$(head -c1 "$DIR/template.json")" != "{" ]]; then
  rm "$DIR/template.json"
  aws cloudformation get-template --stack-name "$1" --query TemplateBody --output text > "$DIR/template.yml"
fi
aws cloudformation describe-stack-resources --stack-name "$1" > "$DIR/resources.json"
aws cloudformation describe-stack-events --stack-name "$1" > "$DIR/events.json"
set -x
ls -lh "$DIR"
