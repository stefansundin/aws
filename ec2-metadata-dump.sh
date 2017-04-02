#!/bin/bash -e
AZ=$(curl http://169.254.169.254/2016-09-02/meta-data/placement/availability-zone)
INSTANCE_ID=$(curl http://169.254.169.254/2016-09-02/meta-data/instance-id)
INSTANCE_TYPE=$(curl http://169.254.169.254/2016-09-02/meta-data/instance-type)
PROFILE=$(curl http://169.254.169.254/2016-09-02/meta-data/profile)
AMI_ID=$(curl http://169.254.169.254/2016-09-02/meta-data/ami-id)
HOSTNAME=$(curl http://169.254.169.254/2016-09-02/meta-data/hostname)
LOCAL_HOSTNAME=$(curl http://169.254.169.254/2016-09-02/meta-data/local-hostname)
LOCAL_IPV4=$(curl http://169.254.169.254/2016-09-02/meta-data/local-ipv4)
PUBLIC_HOSTNAME=$(curl http://169.254.169.254/2016-09-02/meta-data/public-hostname)
PUBLIC_IPV4=$(curl http://169.254.169.254/2016-09-02/meta-data/public-ipv4)
MAC=$(curl http://169.254.169.254/2016-09-02/meta-data/mac)
INTERFACE_ID=$(curl http://169.254.169.254/2016-09-02/meta-data/network/interfaces/macs/$MAC/interface-id)
PUBLIC_KEY=$(curl http://169.254.169.254/2016-09-02/meta-data/public-keys/0/openssh-key)
IAM_INFO=$(curl http://169.254.169.254/2016-09-02/meta-data/iam/info)
IAM_ROLE=$(curl http://169.254.169.254/2016-09-02/meta-data/iam/security-credentials/)
IAM_CREDENTIALS=$(curl http://169.254.169.254/2016-09-02/meta-data/iam/security-credentials/$IAM_ROLE)
INSTANCE_IDENTITY=$(curl http://169.254.169.254/2016-09-02/dynamic/instance-identity/document)

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
echo
echo "https://${AZ::-1}.console.aws.amazon.com/ec2/v2/home?region=${AZ::-1}#Instances:instanceId=$INSTANCE_ID;sort=instanceId"
