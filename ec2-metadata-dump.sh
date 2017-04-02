#!/bin/bash -e
# https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html
# curl -sfL https://raw.githubusercontent.com/stefansundin/aws/master/ec2-metadata-dump.sh | bash -e
# Also print user-data:
# curl -sfL https://raw.githubusercontent.com/stefansundin/aws/master/ec2-metadata-dump.sh | bash -e -s user-data

function get {
  curl -s http://169.254.169.254/2016-09-02/$1
}

>&2 echo "Fetching metadata..."

AZ=$(get meta-data/placement/availability-zone)
INSTANCE_ID=$(get meta-data/instance-id)
INSTANCE_TYPE=$(get meta-data/instance-type)
PROFILE=$(get meta-data/profile)
AMI_ID=$(get meta-data/ami-id)
HOSTNAME=$(get meta-data/hostname)
LOCAL_HOSTNAME=$(get meta-data/local-hostname)
LOCAL_IPV4=$(get meta-data/local-ipv4)
PUBLIC_HOSTNAME=$(get meta-data/public-hostname)
PUBLIC_IPV4=$(get meta-data/public-ipv4)
MAC=$(get meta-data/mac)
INTERFACE_ID=$(get meta-data/network/interfaces/macs/$MAC/interface-id)
PUBLIC_KEY=$(get meta-data/public-keys/0/openssh-key)
IAM_INFO=$(get meta-data/iam/info)
IAM_ROLE=$(get meta-data/iam/security-credentials/)
IAM_CREDENTIALS=$(get meta-data/iam/security-credentials/$IAM_ROLE)
INSTANCE_IDENTITY=$(get dynamic/instance-identity/document)

echo "availability-zone: $AZ"
echo "instance-id: $INSTANCE_ID"
echo "instance-type: $INSTANCE_TYPE"
echo "profile: $PROFILE"
echo "ami-id: $AMI_ID"
echo "hostname: $HOSTNAME"
echo "local-hostname: $LOCAL_HOSTNAME"
echo "local-ipv4: $LOCAL_IPV4"
echo "public-hostname: $PUBLIC_HOSTNAME"
echo "public-ipv4: $PUBLIC_IPV4"
echo "mac: $MAC"
echo "interface-id: $INTERFACE_ID"
echo "ssh key: $PUBLIC_KEY"
echo
echo "iam info: $IAM_INFO"
echo "iam credentials: $IAM_CREDENTIALS"
echo
echo "instance-identity: $INSTANCE_IDENTITY"

if [ $# -gt 0 ]; then echo; fi
for k in "$@"; do
  echo -n "$k: "
  get "$k"
done

echo
echo "https://${AZ::-1}.console.aws.amazon.com/ec2/v2/home?region=${AZ::-1}#Instances:instanceId=$INSTANCE_ID;sort=instanceId"
