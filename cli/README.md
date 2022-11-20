Handy awscli aliases:
- federate: Sign-in to the AWS console with a role.
- assume-role: Assume a role by setting the environment variables.
- s3-url: Translate http urls to S3 into s3:// urls.
- s3-cat: Output the contents of a file on S3.
- s3-sign: Easily sign a GET request.
- ec2-migrate-instance: Stop and start an instance to migrate it to new hardware. If the instance is in an autoscaling group, the HealthCheck process will first be suspended, then re-enabled when done.
- delete-ami: Delete an ami and the snapshot backing it. Checks if there's an instance running from that ami.
- rds-pending-reboot: List RDS databases that are pending reboot due to parameter group updates.
- rds-replicalag: List RDS replicas and their ReplicaLag.
- rds-wait-for-instance: Waits until an RDS database is "available", and then notifies you.
- rds-wait-for-snapshot: Waits until an RDS snapshot is "available", and then notifies you.
- rds-watch-instance-status: Helps you track the exact time an RDS database changes its status, and for how long.
- rds-download-logs: Download all the logs from an RDS database (with no arguments, it downloads logs from all RDS databases).
- cloudfront-wait-deployed: Waits until a CloudFront distribution is "Deployed", and then notifies you.
- cf-validate: Validate a CloudFormation template.
- cf-diff: Diff a stack against a template file.
- cf-dump: Download info about a stack (useful to "backup" a stack along with its parameters before you delete it).
- cf-watch: Watch a stack update in real-time.
- logs-ls: List all CloudWatch log groups.
- kms-decrypt: Easily decrypt some base64-encoded ciphertext.
- route53-find-dupes: Lists your Route53 zones in a way that makes it easy to spot duplicate zones.
- route53-dump: Download your zone information.

# Usage

Example commands:

```bash
aws username
aws federate admin
aws assume-role admin
aws s3-url https://myrandombucket.s3.amazonaws.com/assets/img/logo.png # => s3://myrandombucket/assets/img/logo.png
aws s3-url http://s3.amazonaws.com/myrandombucket/logs/build.log?X-Amz-Date=... # => s3://myrandombucket/logs/build.log
aws s3-cat http://s3.amazonaws.com/myrandombucket/logs/build.log
aws s3-sign myrandombucket/logs/build.log
aws ec2-migrate-instance i-01234567890abcdef
aws delete-ami ami-12352a5d
aws rds-pending-reboot
aws rds-replicalag
aws aws rds-wait-for-instance production-db
aws aws rds-wait-for-snapshot production-db-2017-09-13
aws rds-watch-instance-status production-db
aws rds-download-logs production-db
aws cloudfront-wait-deployed E7Z2NG1MI10E7Q
aws cf-validate webservers.yml
aws cf-diff prod-webservers webservers.yml
AWS_REGION=us-west-2 aws cf-diff stage-webservers webservers.yml
AWS_PROFILE=test aws cf-diff stage-webservers webservers.yml
aws cf-dump prod-webservers
aws cf-watch prod-webservers
aws logs-ls
aws kms-decrypt YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo=
aws route53-find-dupes
aws route53-dump Z6R4FU12H3ATTX
```

Example federate bash aliases:

```bash
alias aws-admin="aws federate admin"
alias aws-admin="aws federate arn:aws:iam::123456789012:role/AdministratorRole arn:aws:iam::123456789012:mfa/username"
```

# Setup

Install Python pre-requisites.

```bash
brew install python
pip3 install -U virtualenv
python3 -m virtualenv venv -p python3
(source venv/bin/activate && pip3 install -U -r requirements.txt)
```

Add awscli aliases. `cat ~/.aws/cli/alias`

```
[toplevel]

whoami = sts get-caller-identity
version = --version
upgrade = !aws --version; eb --version; pip3 install -U --user awscli awsebcli

ecr-login =
  !f() {
    ACCOUNT_ID=$(aws sts get-caller-identity --query Arn --output text | cut -d: -f5)
    yes | eval $(aws ecr get-login --no-include-email --registry-ids $ACCOUNT_ID --region us-west-2)
  }; f

username =
  !f() {
    aws sts get-caller-identity --query Arn --output text | cut -d/ -f2
  }; f

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

sts-clear-cache =
  !f() {
    rm -rvf ~/.aws/cli/cache
  }; f

decode-message =
  !f() {
    aws sts decode-authorization-message --encoded-message "$1" --query DecodedMessage --output text | jq
  }; f

s3-url =
  !f() {
    if [[ "$1" =~ https?://s3.*.amazonaws.com/([^/]+)/([^?]+) ]] || [[ "$1" =~ https?://(.+).s3.amazonaws.com/([^?]+) ]] || [[ "$1" =~ https?://console.aws.amazon.com/s3/buckets/([^/]+)/([^?]+) ]] || [[ "$1" =~ s3://([^/]+)/([^?]+) ]]; then
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
      export AWS_DEFAULT_REGION=$2
    fi
    REGION=${AWS_DEFAULT_REGION:-$(aws configure get region)}
    if aws ec2 describe-instance-status --include-all-instances --query InstanceStatuses[0].InstanceState.Name --output text --instance-ids "$1" 2>&1 | grep InvalidInstanceID.NotFound > /dev/null; then
      echo "Can't find $1 in $REGION. Is it in another region?"
      return
    fi
    if aws ec2 describe-instance-status --instance-ids "$1" --query InstanceStatuses[0].Events | jq -Mre 'if . == null then true else . | map(select(.Description | startswith("[Completed]") | not)) | length == 0 end' > /dev/null; then
      echo "It doesn't look like $1 has a scheduled event."
      echo "Press [Enter] to continue anyway."
      read
    fi
    NAME=$(aws ec2 describe-tags --filters "Name=resource-type,Values=instance" "Name=resource-id,Values=$1" "Name=key,Values=Name" --query Tags[0].Value --output text)
    echo "EC2 NAME: $NAME"
    echo "- https://console.aws.amazon.com/ec2/home?region=$REGION#Instances:search=$1;sort=desc:launchTime"
    ASG=$(aws ec2 describe-tags --filters "Name=resource-type,Values=instance" "Name=resource-id,Values=$1" "Name=key,Values=aws:autoscaling:groupName" --query Tags[0].Value --output text)
    if [ "$ASG" = "None" ]; then
      echo "Instance $1 does not belong to an ASG."
    else
      echo "ASG: $ASG"
      echo "- https://console.aws.amazon.com/ec2/autoscaling/home?region=$REGION#AutoScalingGroups:id=$ASG;filter=$ASG;view=details"
      echo "- DesiredCapacity: $(aws autoscaling describe-auto-scaling-groups --auto-scaling-group-names "$ASG" --query AutoScalingGroups[0].DesiredCapacity --output text)"
    fi
    ASGINFO=$(aws autoscaling describe-auto-scaling-groups --auto-scaling-group-names "$ASG" --query AutoScalingGroups[0])
    if [ "$(echo $ASGINFO | jq -M .LoadBalancerNames)" != "[]" ]; then
      echo "ELB: $(echo $ASGINFO | jq -Mr '.LoadBalancerNames | join(", ")')"
        echo $ASGINFO | jq -Mr ".LoadBalancerNames | map(\"- https://console.aws.amazon.com/ec2/home?region=$REGION#LoadBalancers:search=\"+.) | join(\"\n\")"
    else
      echo "ASG not assigned any ELB."
    fi
    if [ "$(echo $ASGINFO | jq -M .TargetGroupARNs)" != "[]" ]; then
      echo "Target groups: $(echo $ASGINFO | jq -Mr '.TargetGroupARNs | map(split("/")[1]) | join(", ")')"
      echo $ASGINFO | jq -Mr ".TargetGroupARNs | map(\"- https://console.aws.amazon.com/ec2/home?region=$REGION#TargetGroups:search=\"+.) | join(\"\n\")"
    else
      echo "ASG not assigned any target groups."
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
      echo "Sleeping 10 seconds before stopping $1..."
      sleep 10
    fi
    echo "Stopping $1..."
    aws ec2 stop-instances --output text --instance-ids "$1" || return
    while [ "$(aws ec2 describe-instance-status --include-all-instances --query InstanceStatuses[0].InstanceState.Name --output text --instance-ids "$1")" != "stopped" ]; do
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
    if [ $# -gt 1 ]; then
      export AWS_DEFAULT_REGION=$2
    fi

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
    #aws ec2 describe-tags --filters "Name=resource-id,Values=$1"
    echo
    echo "Will delete $1 and $SNAPSHOTS"
    echo "Press [Enter] to continue."
    read

    aws ec2 deregister-image --image-id "$1" || return
    aws ec2 delete-snapshot --snapshot-id "$SNAPSHOTS" || return
    echo "Done!"
  }; f

rds-pending-reboot =
  !f() {
    for region in ${@:-us-east-1 us-east-2 us-west-1 us-west-2}; do
      >&2 echo $region
      export AWS_DEFAULT_REGION=$region
      aws rds describe-db-instances --query 'DBInstances[].[DBInstanceIdentifier,DBInstanceStatus,DBParameterGroups[0].ParameterApplyStatus]' --output table
      aws rds describe-db-clusters --query 'DBClusters[].[DBClusterIdentifier,Status,join(`, `, DBClusterMembers[].join(`=`,[DBInstanceIdentifier,DBClusterParameterGroupStatus]))]' --output table
      echo
    done
  }; f

rds-pending-maintenance =
  !f() {
    for region in ${@:-us-east-1 us-east-2 us-west-1 us-west-2}; do
      >&2 echo $region
      export AWS_DEFAULT_REGION=$region
      aws rds describe-pending-maintenance-actions --query 'PendingMaintenanceActions[].[ResourceIdentifier,PendingMaintenanceActionDetails[0].Action,PendingMaintenanceActionDetails[0].Description]' --output table
      echo
    done
  }; f

rds-summary =
  !f() {
    for region in ${@:-us-east-1 us-east-2 us-west-1 us-west-2}; do
      >&2 echo $region
      export AWS_DEFAULT_REGION=$region
      aws rds describe-db-instances --query 'DBInstances[].[DBInstanceIdentifier,DBInstanceStatus,DBParameterGroups[0].ParameterApplyStatus,StorageEncrypted,StorageType,AllocatedStorage]' --output table
      echo
    done
  }; f

rds-replicalag =
  !f() {
    for region in us-west-2 us-east-1 us-east-2 us-west-1; do
      export AWS_DEFAULT_REGION=$region
      echo "$region"
      replicas=$(aws rds describe-db-instances --query 'DBInstances[?ReadReplicaSourceDBInstanceIdentifier!=null].DBInstanceIdentifier')
      aws cloudwatch get-metric-data --metric-data-queries "$(echo "$replicas" | jq 'map({"Id":gsub("-";"_"),"MetricStat":{"Period":60,"Stat":"Average","Metric":{"Namespace":"AWS/RDS","MetricName":"ReplicaLag","Dimensions":[{"Name":"DBInstanceIdentifier","Value":.}]}}})')" --start-time $(date -v-3M -u +%FT%TZ) --end-time $(date -v+3M -u +%FT%TZ) --query MetricDataResults | jq -Mr 'map([.Label,.Values[0] | tostring])[] | join("\t")'
      echo "CloudWatch link: https://${AWS_DEFAULT_REGION}.console.aws.amazon.com/cloudwatch/home?region=${AWS_DEFAULT_REGION}#metricsV2:graph=~(view~'timeSeries~stacked~false~metrics~($(echo "$replicas" | jq -Mr "map(\"~(~'...~'\"+.+\")\") | join(\"\") | sub(\"\\\\...\"; \"AWS*2fRDS~'ReplicaLag~'DBInstanceIdentifier\")"))~region~'${AWS_DEFAULT_REGION});search=ReplicaLag;namespace=AWS/RDS;dimensions=DBInstanceIdentifier"
      echo
    done
  }; f

rds-replicalag2 =
  !f() {
    for region in us-west-2 us-east-1 us-east-2 us-west-1; do
      export AWS_DEFAULT_REGION=$region
      while read db source; do
        [[ "$source" == "None" ]] && continue
        lag=$(aws cloudwatch get-metric-statistics --namespace AWS/RDS --dimensions Name=DBInstanceIdentifier,Value=$db --metric-name ReplicaLag --start-time $(date -v-3M -u +%FT%TZ) --end-time $(date -v+3M -u +%FT%TZ) --period 60 --statistics Average | jq -M '.Datapoints | sort_by(.Timestamp) | reverse[0].Average')
        if [[ $lag != "null" && $lag -gt 3600 ]]; then
          printf "%s | %-30s | ReplicaLag: %s seconds (%s hours)\n" $region $db $lag $(( $lag / 3600 ))
        else
          printf "%s | %-30s | ReplicaLag: %s seconds\n" $region $db $lag
        fi
      done <<< "$(aws rds describe-db-instances --query 'DBInstances[].[DBInstanceIdentifier,ReadReplicaSourceDBInstanceIdentifier]' --output text)"
    done
  }; f

rds-wait-for-instance =
  !f() {
    db="$1"
    if [[ "$db" =~ ^([a-z0-9\-]+)\.[a-z0-9]+\.([a-z0-9\-]+)\.rds\.amazonaws\.com ]]; then
      db=${BASH_REMATCH[1]}
      export AWS_DEFAULT_REGION=${BASH_REMATCH[2]}
    elif [[ "$db" =~ ^arn:aws:rds:([a-z0-9\-]+):[0-9]+:db:([a-z0-9\-]+)$ ]]; then
      db=${BASH_REMATCH[2]}
      export AWS_DEFAULT_REGION=${BASH_REMATCH[1]}
    elif [ $# -gt 1 ]; then
      export AWS_DEFAULT_REGION=$2
    fi
    while [ true ]; do
      data=$(aws rds describe-db-instances --db-instance-identifier "$db" --query DBInstances[0])
      if [ $? != 0 ]; then
        afplay /System/Library/Sounds/Ping.aiff
        printf '\007'
        sleep 1
        continue
      fi
      state=$(echo $data | jq -Mr .DBInstanceStatus)
      sourcedb=$(echo $data | jq -Mr .ReadReplicaSourceDBInstanceIdentifier)
      if [ "$sourcedb" != "null" ]; then
        lag=$(aws cloudwatch get-metric-statistics --namespace AWS/RDS --dimensions Name=DBInstanceIdentifier,Value=$db --metric-name ReplicaLag --start-time $(date -v-3M -u +%FT%TZ) --end-time $(date -v+3M -u +%FT%TZ) --period 60 --statistics Average | jq -M '.Datapoints | sort_by(.Timestamp) | reverse[0].Average')
        state="$state (ReplicaLag: $lag seconds)"
        if [[ "$state" != "available"* ]]; then
          snaps=$(aws rds describe-db-snapshots --db-instance-identifier "$sourcedb" | jq -Mr '.DBSnapshots | map(select(.Status != "available")) | map(.Status+" "+(.PercentProgress|tostring)+"%") | join(", ")')
          if [ "$snaps" != "" ]; then
            state="$state ($sourcedb snapshot progress: $snaps)"
          fi
        fi
      fi
      # slaves can also have backups enabled:
      if [[ "$state" != "available"* ]]; then
        snaps=$(aws rds describe-db-snapshots --db-instance-identifier "$db" | jq -Mr '.DBSnapshots | map(select(.Status != "available")) | map(.Status+" "+(.PercentProgress|tostring)+"%") | join(", ")')
        if [ "$snaps" != "" ]; then
          state="$state (snapshot progress: $snaps)"
        fi
      fi
      echo "$(date "+%F %T"): $db state: $state"
      if [[ "$state" == "available" || "$state" == "available (ReplicaLag: 0 seconds)" ]]; then
        afplay /System/Library/Sounds/Ping.aiff
        printf '\007'
        #say "Database instance ready."
      fi
      sleep 3
    done
  }; f

rds-wait-for-cluster =
  !f() {
    db="$1"
    if [[ "$db" =~ ^([a-z0-9\-]+)\.cluster-[a-z0-9\-]+\.([a-z0-9\-]+)\.rds\.amazonaws\.com ]]; then
      db=${BASH_REMATCH[1]}
      export AWS_DEFAULT_REGION=${BASH_REMATCH[2]}
    elif [[ "$db" =~ ^arn:aws:rds:([a-z0-9\-]+):[0-9]+:cluster:([a-z0-9\-]+)$ ]]; then
      db=${BASH_REMATCH[2]}
      export AWS_DEFAULT_REGION=${BASH_REMATCH[1]}
    elif [ $# -gt 1 ]; then
      export AWS_DEFAULT_REGION=$2
    fi
    while [ true ]; do
      data=$(aws rds describe-db-clusters --db-cluster-identifier "$db" --query DBClusters[0])
      if [ $? != 0 ]; then
        afplay /System/Library/Sounds/Ping.aiff
        printf '\007'
        sleep 1
        continue
      fi
      state=$(echo $data | jq -Mr .Status)

      instance_state=$(echo $data | jq -Mr '.DBClusterMembers | map(select(.DBInstanceIdentifier | startswith("application-autoscaling-") | not)) | sort_by(.DBInstanceIdentifier) | map(.DBInstanceIdentifier + if .IsClusterWriter then " (writer)" else "" end) | join(", ")')
      num_replicas=$(echo $data | jq -Mr '.DBClusterMembers | map(select(.DBInstanceIdentifier | startswith("application-autoscaling-"))) | length')
      echo "$(date "+%F %T"): [$db] Cluster state: $state. Instances: $instance_state. $num_replicas autoscaled replicas."

      if [[ "$state" == "available" ]]; then
        instance_data=$(aws rds describe-db-instances --filters Name=db-cluster-id,Values=$db --query 'DBInstances[].{id:DBInstanceIdentifier,status:DBInstanceStatus}')
        instance_info=$(echo $instance_data | jq -Mr 'map(select(.id | startswith("application-autoscaling-") | not)) | map(.id + " ("+.status+")") | join(", ")')
        replica_info=$(echo $instance_data | jq -Mr 'map(select(.id | startswith("application-autoscaling-"))) | group_by(.status) | map((length | tostring) + " " + .[0].status) | join(", ")')
        [[ "$replica_info" != "" ]] && replica_info=" Autoscaling: $replica_info."
        echo "Instance details: $instance_info.$replica_info"
        if [[ "$(echo $instance_data | jq -Mr 'all(.status == "available")')" == "true" ]]; then
          afplay /System/Library/Sounds/Ping.aiff
          printf '\007'
        fi
      fi
      sleep 3
    done
  }; f

rds-wait-for-snapshot =
  !f() {
    if [ $# -gt 1 ]; then
      export AWS_DEFAULT_REGION=$2
    fi
    while [ true ]; do
      data=$(aws rds describe-db-snapshots --db-snapshot-identifier "$1" --query DBSnapshots[0])
      [ $? != 0 ] && sleep 1 && continue
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

rds-watch-instance-status =
  !f() {
    db="$1"
    if [[ "$db" =~ ^([a-z0-9\-]+)\.[a-z0-9]+\.([a-z0-9\-]+)\.rds\.amazonaws\.com ]]; then
      db=${BASH_REMATCH[1]}
      export AWS_DEFAULT_REGION=${BASH_REMATCH[2]}
    elif [[ "$db" =~ ^arn:aws:rds:([a-z0-9\-]+):[0-9]+:db:([a-z0-9\-]+)$ ]]; then
      db=${BASH_REMATCH[2]}
      export AWS_DEFAULT_REGION=${BASH_REMATCH[1]}
    elif [ $# -gt 1 ]; then
      export AWS_DEFAULT_REGION=$2
    fi
    last_state=""
    while [ true ]; do
      data=$(aws rds describe-db-instances --db-instance-identifier "$db" --query DBInstances[0] 2>/dev/null)
      if [ $? == 0 ]; then
        state=$(echo "$data" | jq -Mr .DBInstanceStatus)
        sourcedb=$(echo "$data" | jq -Mr .ReadReplicaSourceDBInstanceIdentifier)
        if [ "$sourcedb" != "null" ]; then
          snaps=$(aws rds describe-db-snapshots --db-instance-identifier "$sourcedb" | jq -Mr '.DBSnapshots | map(select(.Status != "available")) | map(.Status+" "+(.PercentProgress|tostring)+"%") | join(", ")')
          if [ "$snaps" != "" ]; then
            state="$state ($sourcedb snapshot progress: $snaps)"
          fi
        fi
      else
        state="not found"
      fi
      if [ "$state" != "$last_state" ]; then
        last_state=$state
        echo "$(date "+%F %T"): $db state: $state"
        #afplay /System/Library/Sounds/Ping.aiff
      fi
      sleep 1
    done
  }; f

rds-download-logs =
  !f() {
    set -e
    dbs=("$1")
    if [ $# -eq 0 ]; then
      for region in us-east-1 us-east-2 us-west-1 us-west-2; do
        export AWS_DEFAULT_REGION=$region
        dbs+=($(aws rds describe-db-instances --query DBInstances[*].[DBInstanceArn] --output text))
      done
    elif [ $# -gt 1 ]; then
      export AWS_DEFAULT_REGION=$2
    fi
    for db in ${dbs[@]}; do
      if [[ "$db" =~ ^([a-z0-9\-]+)\.[a-z0-9]+\.([a-z0-9\-]+)\.rds\.amazonaws\.com ]]; then
        db=${BASH_REMATCH[1]}
        export AWS_DEFAULT_REGION=${BASH_REMATCH[2]}
      elif [[ "$db" =~ ^arn:aws:rds:([a-z0-9\-]+):[0-9]+:db:([a-z0-9\-]+)$ ]]; then
        db=${BASH_REMATCH[2]}
        export AWS_DEFAULT_REGION=${BASH_REMATCH[1]}
      fi
      dir="$db-$(date -u +%FT%H-%M-%SZ)"
      mkdir -p "$dir"
      aws rds describe-events --source-type db-instance --source-identifier "$db" --start-time $(date -v-13d -u +%FT%TZ) --end-time $(date -v+1d -u +%FT%TZ) --output json > "$dir/events.json"
      echo "Downloaded recent events to $dir/events.json"
      while read -r log; do
        f=$(echo "$log" | jq -Mr .LogFileName)
        d=$(echo "$log" | jq -Mr '.LastWritten / 1000')
        kb=$(echo "$log" | jq -Mr '.Size / 1024 | floor')
        echo "Downloading $dir/$f ($kb kB), last written $(date -r $d +'%Y-%m-%d %H:%M:%S')"
        mkdir -p "$dir/$(dirname "$f")"
        aws rds download-db-log-file-portion --db-instance-identifier "$db" --log-file-name "$f" --starting-token 0 --output text > "$dir/$f"
        touch -m -t "$(date -r $d +%Y%m%d%H%M.%S)" "$dir/$f"
      done <<< "$(aws rds describe-db-log-files --db-instance-identifier "$db" | jq -Mc '.DescribeDBLogFiles | sort_by(.LastWritten) | reverse | .[]')"
    done
  }; f

elasticache-wait-for-instance =
  !f() {
    db="$1"
    if [[ "$db" =~ ^([a-z0-9\-]+)\.[a-z0-9]+\.(ng\.)?[0-9]+\.([a-z0-9\-]+)\.cache\.amazonaws\.com ]]; then
      db=${BASH_REMATCH[1]}
      r=${BASH_REMATCH[3]}
      [ "$r" == "use1" ] && r=us-east-1
      [ "$r" == "use2" ] && r=us-east-2
      [ "$r" == "usw1" ] && r=us-west-1
      [ "$r" == "usw2" ] && r=us-west-2
      export AWS_DEFAULT_REGION=$r
    elif [ $# -gt 1 ]; then
      export AWS_DEFAULT_REGION=$2
    fi
    while true; do
      data=$(aws elasticache describe-cache-clusters --cache-cluster-id "$db" --query CacheClusters[0])
      if [ $? != 0 ]; then
        afplay /System/Library/Sounds/Ping.aiff
        printf '\007'
        sleep 1
        continue
      fi
      state=$(echo $data | jq -Mr .CacheClusterStatus)
      echo "$(date "+%F %T"): $db state: $state"
      if [[ "$state" == "available" ]]; then
        afplay /System/Library/Sounds/Ping.aiff
        printf '\007'
      fi
      sleep 3
    done
  }; f

elasticache-wait-for-primary =
  !f() {
    db="$1"
    if [[ "$db" =~ ^([a-z0-9\-]+)\.[a-z0-9]+\.(ng\.)?[0-9]+\.([a-z0-9\-]+)\.cache\.amazonaws\.com ]]; then
      db=${BASH_REMATCH[1]}
      r=${BASH_REMATCH[3]}
      [ "$r" == "use1" ] && r=us-east-1
      [ "$r" == "use2" ] && r=us-east-2
      [ "$r" == "usw1" ] && r=us-west-1
      [ "$r" == "usw2" ] && r=us-west-2
      export AWS_DEFAULT_REGION=$r
    elif [ $# -gt 1 ]; then
      export AWS_DEFAULT_REGION=$2
    fi
    group=$(aws elasticache describe-cache-clusters --cache-cluster-id "$db" --query CacheClusters[0].ReplicationGroupId --output text)
    echo "replication group: $group"
    while [ true ]; do
      role=$(aws elasticache describe-replication-groups --replication-group-id "$group" --query "ReplicationGroups[0].NodeGroups[0].NodeGroupMembers[?CacheClusterId=='$db'].CurrentRole | [0]" --output text)
      if [ $? != 0 ]; then
        afplay /System/Library/Sounds/Ping.aiff
        printf '\007'
        sleep 1
        continue
      fi
      echo "$(date "+%F %T"): $db role: $role"
      if [[ "$role" == "primary" ]]; then
        afplay /System/Library/Sounds/Ping.aiff
        printf '\007'
      fi
      sleep 3
    done
  }; f

cloudfront-wait-deployed =
  !f() {
    if [ $# -gt 1 ]; then
      export AWS_DEFAULT_REGION=$2
    fi
    if [ $# -gt 2 ]; then
      export AWS_PROFILE=$3
    fi
    echo "https://console.aws.amazon.com/cloudfront/home?#distribution-settings:$1"
    while [ true ]; do
      status=$(aws cloudfront get-distribution --id "$1" --query Distribution.Status --output text)
      [ $? != 0 ] && sleep 1 && continue
      echo "$(date "+%F %T"): $1 status: $status"
      if [ "$status" == "Deployed" ]; then
        afplay /System/Library/Sounds/Ping.aiff
        #say "CloudFront distribution deployed."
      fi
      sleep 1
    done
  }; f

cloudfront-invalidate =
  !f() {
    if [ $# -gt 1 ]; then
      export AWS_DEFAULT_REGION=$2
    fi
    if [ $# -gt 2 ]; then
      export AWS_PROFILE=$3
    fi
    echo "https://console.aws.amazon.com/cloudfront/home?#distribution-settings:$1"
    id=$(aws cloudfront create-invalidation --distribution-id "$1" --paths '/*' --query Invalidation.Id --output text)
    while [ true ]; do
      status=$(aws cloudfront get-invalidation --distribution-id "$1" --id "$id" --query Invalidation.Status --output text)
      [ $? != 0 ] && sleep 1 && continue
      echo "$(date "+%F %T"): $id invalidation: $status"
      if [ "$status" == "Completed" ]; then
        break
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

# AWS_DEFAULT_REGION=us-east-2 aws logs-put --log-group-name my-log-group --log-stream-name my-log-stream --log-events file://test.json
logs-put =
  !f() {
    ERROR=$(aws logs put-log-events "$@" 3>&1 1>&2 2>&3 3>&-)
    if [[ "$?" == "255" ]]; then
      token="${ERROR##* }"
      echo "Retrying with sequence token: $token"
      aws logs put-log-events "$@" --sequence-token "$token"
    fi
  }; f

# aws logs-download my-log-group my-log-stream
# currently limited to the first 10k lines. see https://stackoverflow.com/questions/28060845/aws-cloudwatch-log-is-that-possible-to-export-existing-log-data-from-it#28677049
logs-download =
  !f() {
    aws logs get-log-events --log-group-name "$1" --log-stream-name "$2" --output text > "$2.log"
  }; f

kms-decrypt =
  !f() {
    export AWS_PROFILE="${AWS_PROFILE:-admin}"
    bash -c 'aws kms decrypt --ciphertext-blob fileb://<(echo "$@" | base64 -D) --query Plaintext --output text | base64 -D' dummy "$@"
    echo
  }; f

route53domains-summary =
  !f() {
    echo "Expiration date     \tRenew\tLocked\tDomain name"
    aws route53domains list-domains --region us-east-1 --max-items 200 --query Domains | jq -Mr '.[] | .Expiry |= todate | [.Expiry,.AutoRenew,.TransferLock,.DomainName] | @tsv'
  }; f

route53-find-dupes =
  !f() {
    aws route53 list-hosted-zones --query HostedZones[].[Name] --output text | sort | uniq -c
  }; f

route53-dump =
  !f() {
    if [ $# -gt 0 ]; then
      zones="$@"
    else
      zones="$(aws route53 list-hosted-zones --query HostedZones[].[Id] --output text | sed -e 's/\/hostedzone\///')"
    fi
    dir="route53-zones-$(date -u +%FT%H-%M-%SZ)"
    mkdir -p "$dir"
    for zone in $zones; do
      aws route53 get-hosted-zone --id "$zone" --output json > "$dir/$zone.json"
      domain=$(jq -Mr .HostedZone.Name < "$dir/$zone.json")
      echo "$dir/$zone-${domain}json"
      aws route53 list-resource-record-sets --hosted-zone-id "$zone" --output json > "$dir/$zone-${domain}json"
    done
  }; f

# aws route53-ssl-check ZNUZH3NNCNAUQ
route53-ssl-check =
  !f() {
    if [ $# -lt 1 ]; then
      echo "Please supply zone id."
      exit 1
    fi
    zone="$1"
    temp=$(mktemp)
    aws route53 list-resource-record-sets --hosted-zone-id $zone --query 'ResourceRecordSets[].[Name,Type]' --output text | sort | while read -r domain type; do
      [[ "$type" != "A" && "$type" != "AAAA" && "$type" != "CNAME" ]] && continue
      nc -z -G 1 "$domain" 443 2>/dev/null || continue
      dates=$(echo Q | openssl s_client -connect "$domain:443" -servername "$domain" 2>$temp | openssl x509 -noout -dates 2>/dev/null || echo "error")
      issuer=$(cat $temp | grep -m 1 depth=1 || cat $temp | grep -m 1 depth=0)
      [[ "$issuer" == *"O = Let's Encrypt"* ]] && continue
      rm $temp
      printf "%-50s %-75s %s\n" "$domain" "${dates//$'\n'/ }" "${issuer:0:100}"
    done
  }; f

cw-dump =
  !f() {
    if [ $# -eq 0 ]; then
      echo "Available dashboards:"
      aws cloudwatch list-dashboards --query DashboardEntries[].[DashboardName] --output text
      exit 1
    fi
    if [ $# -gt 1 ]; then
      export AWS_DEFAULT_REGION=$2
    fi
    fn="$1-$(date +%F).json"
    aws cloudwatch get-dashboard --dashboard-name "$1" --query DashboardBody --output text > "$fn"
    echo "Created file $fn"
  }; f

```
