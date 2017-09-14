Handy awscli aliases:
- federate: Sign-in to the AWS console with a role.
- assume-role: Assume a role by setting the environment variables.
- s3-url: Translate http urls to S3 into s3:// urls.
- s3-cat: Output the contents of a file on S3.
- s3-sign: Easily sign a GET request.
- ec2-migrate-instance: Stop and start an instance to migrate it to new hardware. If the instance is in an autoscaling group, you should first suspend the HealthCheck process (do not forget to remove it again!).
- cf-validate: Validate a CloudFormation template.
- cf-diff: Diff a stack against a template file.
- cf-dump: Download info about a stack (useful to "backup" a stack along with its parameters before you delete it).
- cf-watch: Watch a stack update in real-time.
- logs-ls: List all CloudWatch log groups.
- kms-decrypt: Easily decrypt some base64-encoded ciphertext.

# Usage

Example commands:

```bash
aws federate admin
aws assume-role admin
aws s3-url https://myrandombucket.s3.amazonaws.com/assets/img/logo.png # => s3://myrandombucket/assets/img/logo.png
aws s3-url http://s3.amazonaws.com/myrandombucket/logs/build.log?X-Amz-Date=... # => s3://myrandombucket/logs/build.log
aws s3-cat http://s3.amazonaws.com/myrandombucket/logs/build.log
aws s3-sign myrandombucket/logs/build.log
aws ec2-migrate-instance i-01234567890abcdef
aws ec2-complex-migrate-instance i-01234567890abcdef
aws rds-wait-for-instance production-db
aws rds-wait-for-snapshot production-db-2017-09-13
aws cf-validate webservers.yml
aws cf-diff prod-webservers webservers.yml
AWS_REGION=us-west-2 aws cf-diff stage-webservers webservers.yml
AWS_PROFILE=test aws cf-diff stage-webservers webservers.yml
aws cf-dump prod-webservers
aws cf-watch prod-webservers
aws logs-ls
aws kms-decrypt YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo=
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
pip3 install -U virtualenv
python3 -m virtualenv venv -p python3
(source venv/bin/activate && pip3 install -U -r requirements.txt)
```

Add awscli aliases. `cat ~/.aws/cli/alias`

```
[toplevel]

whoami = sts get-caller-identity
version = --version
upgrade = !aws version && pip2 install -U awscli

federate =
  !f() {
    DIR=~/src/aws/cli
    source "$DIR/venv/bin/activate"
    "$DIR/federate.py" "$@"
  }; f

assume-role =
  !f() {
    DIR=~/src/aws/cli
    source "$DIR/venv/bin/activate"
    "$DIR/assume-role.py" "$@"
  }; f

decode-message =
  !f() {
    aws sts decode-authorization-message --encoded-message "$1" | jq -Mr .DecodedMessage | jq
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

s3-sign =
  !f() {
    DAY=$(date -v+2d +%s)
    EXPIRES=${2:-$DAY}
    SIG=$(printf "GET\n\n\n$EXPIRES\n/$1" | \
          openssl dgst -sha1 -binary -hmac "$(aws configure get aws_secret_access_key)" | \
          openssl base64 | perl -MURI::Escape -ne 'chomp;print uri_escape($_),"\n"')
    echo "https://s3.amazonaws.com/$1?AWSAccessKeyId=$(aws configure get aws_access_key_id)&Expires=$EXPIRES&Signature=$SIG"
  }; f

ec2-migrate-instance =
  !f() {
    if [ $# -gt 1 ]; then
      echo "Region: $2"
      export AWS_DEFAULT_REGION=$2
    fi
    echo "Stopping $1..."
    aws ec2 stop-instances --output text --instance-ids "$1" || return
    while [ $(aws ec2 describe-instance-status --include-all-instances --query InstanceStatuses[0].InstanceState.Name --output text --instance-ids "$1") != "stopped" ]; do
      echo "Waiting for $1 to reach state 'stopped'..."
      sleep 1
    done
    echo "Starting $1..."
    aws ec2 start-instances --output text --instance-ids "$1"
  }; f

ec2-complex-migrate-instance =
  !f() {
    if [ $# -gt 1 ]; then
      echo "Region: $2"
      export AWS_DEFAULT_REGION=$2
    fi
    export AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION:-$(aws configure get region)}
    if aws ec2 describe-instance-status --include-all-instances --query InstanceStatuses[0].InstanceState.Name --output text --instance-ids "$1" 2>&1 | grep InvalidInstanceID.NotFound > /dev/null; then
      echo "Can't find $1 in $AWS_DEFAULT_REGION. Is it in another region?"
      return
    fi
    if aws ec2 describe-instance-status --instance-ids "$1" --query InstanceStatuses[0].Events | jq -Mre '. | map(select(.Description | startswith("[Completed]") | not)) | length == 0' > /dev/null; then
      echo "It doesn't look like $1 has a scheduled event."
      echo "Press [Enter] to continue anyway."
      read
    fi
    NAME=$(aws ec2 describe-tags --filters "Name=resource-type,Values=instance" "Name=resource-id,Values=$1" "Name=key,Values=Name" --query Tags[0].Value --output text)
    echo "NAME: $NAME"
    echo "EC2: https://console.aws.amazon.com/ec2/v2/home?region=$AWS_DEFAULT_REGION#Instances:search=$1;sort=desc:launchTime"
    ASG=$(aws ec2 describe-tags --filters "Name=resource-type,Values=instance" "Name=resource-id,Values=$1" "Name=key,Values=aws:autoscaling:groupName" --query Tags[0].Value --output text)
    if [ "$ASG" = "None" ]; then
      echo "Instance $1 does not belong to an ASG."
    else
      echo "ASG: $ASG"
      echo "ASG: https://console.aws.amazon.com/ec2/autoscaling/home?region=$AWS_DEFAULT_REGION#AutoScalingGroups:id=$ASG;filter=$ASG;view=details"
    fi
    echo "Press [Enter] to continue."
    read
    if [ "$ASG" != "None" ]; then
      if aws autoscaling describe-auto-scaling-groups --auto-scaling-group-names "$ASG" --query AutoScalingGroups[0].SuspendedProcesses --output text | grep HealthCheck; then
        echo "HealthCheck is already suspended."
      else
        echo "HealthCheck is not suspended, suspending..."
        aws autoscaling suspend-processes --auto-scaling-group-name "$ASG" --scaling-processes HealthCheck
      fi
      ASGINFO=$(aws autoscaling describe-auto-scaling-groups --auto-scaling-group-names "$ASG" --query AutoScalingGroups[0])
      if [ "$(echo $ASGINFO | jq -M .LoadBalancerNames)" != "[]" ]; then
        ELB=$(echo $ASGINFO | jq -Mr .LoadBalancerNames[0])
        echo "ELB: $ELB"
        echo "ELB: https://console.aws.amazon.com/ec2/v2/home?region=$AWS_DEFAULT_REGION#LoadBalancers:search=$ELB"
      fi
      echo "Sleeping 10 seconds before stopping $1..."
      sleep 10
    fi
    echo "Stopping $1..."
    aws ec2 stop-instances --output text --instance-ids "$1" || return
    while [ $(aws ec2 describe-instance-status --include-all-instances --query InstanceStatuses[0].InstanceState.Name --output text --instance-ids "$1") != "stopped" ]; do
      echo "Waiting for $1 to reach state 'stopped'..."
      sleep 1
    done
    echo "Starting $1..."
    aws ec2 start-instances --output text --instance-ids "$1"
    if [ "$ASG" != "None" ]; then
      echo "Sleeping 5 seconds before resuming HealthCheck on ASG..."
      sleep 5
      aws autoscaling resume-processes --auto-scaling-group-name "$ASG" --scaling-processes HealthCheck
    fi
  }; f

# TODO: test this with an image backed by multiple snapshots!!!
delete-ami =
  !f() {
    export AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION:-$(aws configure get region)}
    RUNNING_INSTANCES=$(aws ec2 describe-instances --filters "Name=image-id,Values=$1" --query Reservations[*].Instances[*].[InstanceId] --output text)
    if [[ -n "$RUNNING_INSTANCES" ]]; then
      echo "This ami was used to launch the following instances:"
      echo "$RUNNING_INSTANCES"
      return 1
    fi

    aws ec2 describe-images --owners self --image-ids "$1" --query Images[0] || return
    SNAPSHOTS=$(aws ec2 describe-images --owners self --image-ids "$1" --query Images[*].BlockDeviceMappings[?Ebs].Ebs.SnapshotId --output text)
    aws ec2 describe-snapshots --snapshot-ids "$SNAPSHOTS" --query Snapshots[*] || return
    aws ec2 describe-image-attribute --attribute launchPermission --image-id "$1"
    echo
    echo "Will delete $1 and $SNAPSHOTS"
    echo "Press [Enter] to continue."
    read

    aws ec2 deregister-image --image-id "$1" || return
    aws ec2 delete-snapshot --snapshot-id "$SNAPSHOTS" || return
    echo "Done!"
  }; f

rds-wait-for-instance =
  !f() {
    if [ $# -gt 1 ]; then
      export AWS_DEFAULT_REGION=$2
    fi
    while [ true ]; do
      data=$(aws rds describe-db-instances --db-instance-identifier "$1" --query DBInstances[0])
      state=$(echo $data | jq -Mr .DBInstanceStatus)
      echo "$(date "+%F %T"): $1 state: $state"
      if [ "$state" == "available" ]; then
        afplay /System/Library/Sounds/Ping.aiff
        #say "Database instance ready."
      fi
      sleep 1
    done
  }; f

rds-wait-for-snapshot =
  !f() {
    if [ $# -gt 1 ]; then
      export AWS_DEFAULT_REGION=$2
    fi
    while [ true ]; do
      data=$(aws rds describe-db-snapshots --db-snapshot-identifier "$1" --query DBSnapshots[0])
      state=$(echo $data | jq -Mr .Status)
      progress=$(echo $data | jq -Mr .PercentProgress)
      echo "$(date "+%F %T"): $1 state: $state $progress%"
      if [ "$state" == "available" ]; then
        afplay /System/Library/Sounds/Ping.aiff
        #say "Database snapshot done."
      fi
      sleep 1
    done
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
    watch -d -n1 "~/stuff/aws/cli/cf-watch.sh \"$@\""
  }; f

logs-ls =
  !f() {
    aws logs describe-log-groups --query 'logGroups[*].logGroupName' | jq -r '.[]'
  }; f

kms-decrypt =
  !f() {
    export AWS_PROFILE="${AWS_PROFILE:-admin}"
    bash -c 'aws kms decrypt --ciphertext-blob fileb://<(echo "$@" | base64 -D) --query Plaintext --output text | base64 -D' dummy "$@"
    echo
  }; f

```
