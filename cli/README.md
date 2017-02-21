Handy awscli aliases:
- federate: Assume a role and create a sign-in link to the AWS console.
- s3-url: Translate http urls to S3 into s3:// urls.
- s3-cat: Output the contents of a file on S3.
- cf-validate: Validate a CloudFormation template.
- cf-diff: Diff a stack against a template file.
- cf-dump: Download info about a stack (useful to "backup" a stack along with its parameters before you delete it).
- cf-watch: Watch a stack update in real-time.
- logs-ls: List all CloudWatch log groups.

# Usage

Example commands:

```bash
aws federate admin
aws s3-url https://myrandombucket.s3.amazonaws.com/assets/img/logo.png # => s3://myrandombucket/assets/img/logo.png
aws s3-url http://s3.amazonaws.com/myrandombucket/logs/build.log?X-Amz-Date=... # => s3://myrandombucket/logs/build.log
aws s3-cat http://s3.amazonaws.com/myrandombucket/logs/build.log
aws cf-validate webservers.yml
aws cf-diff prod-webservers webservers.yml
AWS_REGION=us-west-2 aws cf-diff stage-webservers webservers.yml
AWS_PROFILE=test aws cf-diff stage-webservers webservers.yml
aws cf-dump prod-webservers
aws cf-watch prod-webservers
aws logs-ls
```

Example federate bash aliases:

```bash
alias aws-admin="aws federate admin"
alias aws-admin="aws federate arn:aws:iam::123456789012:role/AdministratorRole arn:aws:iam::123456789012:mfa/username"
```

# Setup

Install Python pre-requisites.

```bash
brew install python3
python3 -m virtualenv venv -p python3
(source venv/bin/activate && pip3 install --upgrade -r requirements.txt)
```

Add awscli alias. `cat ~/.aws/cli/alias`

```
[toplevel]

whoami = sts get-caller-identity

federate =
  !f() {
    DIR=~/src/aws/cli
    source "$DIR/venv/bin/activate"
    "$DIR/federate.py" "$@"
  }; f

s3-url =
  !f() {
    if [[ "$1" =~ https?://s3.amazonaws.com/([^/]+)/([^?]+) ]] || [[ "$1" =~ https?://(.+).s3.amazonaws.com/([^?]+) ]] || [[ "$1" =~ s3://([^/]+)/([^?]+) ]]; then
      echo "s3://${BASH_REMATCH[1]}/${BASH_REMATCH[2]}"
    else
      >&2 echo "Invalid url."
    fi
  }; f

s3-cat =
  !f() {
    URL=$(aws s3-url "$1")
    FILE=$(mktemp)
    aws s3 cp "$URL" "$FILE"
    cat "$FILE"
    rm "$FILE"
  }; f

cf-validate =
  !f() {
    DIR=~/src/aws/cli
    source "$DIR/venv/bin/activate"
    "$DIR/cf-validate.py" ${@:-*.yml}
  }; f

cf-diff =
  !f() {
    DIR=~/src/aws/cli
    source "$DIR/venv/bin/activate"
    "$DIR/cf-diff.py" "$@"
  }; f

cf-dump =
  !f() {
    DIR=~/src/aws/cli
    "$DIR/cf-dump.sh" "$@"
  }; f

cf-watch =
  !f() {
    watch -n1 "aws cloudformation describe-stack-events --stack-name $1 --region ${2:-us-east-1} --profile ${3:-default} --max-items 12 --query 'StackEvents[*].[Timestamp, ResourceStatus, LogicalResourceId, ResourceStatusReason]' --output text | column -t -s $'\t'"
  }; f

logs-ls =
  !f() {
    aws logs describe-log-groups --query 'logGroups[*].logGroupName' | jq -r '.[]'
  }; f

```
