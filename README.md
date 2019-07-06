Get key identity:
```
AWS_ACCESS_KEY_ID=AKIA.. AWS_SECRET_ACCESS_KEY=... aws sts get-caller-identity
```

## S3

Find and abort stale S3 multipart uploads:
```
for bucket in $(aws s3api list-buckets --query Buckets[*].Name --output text); do
  echo "$bucket"
  aws s3api list-multipart-uploads --bucket "$bucket" --query Uploads[*].[Key,UploadId,Initiated,Initiator.DisplayName] --output text | while read key id date user; do
    [[ "$key" == "None" ]] && continue
    echo "Press enter to abort s3://$bucket/$key (initiated $date by $user)"
    read < /dev/tty
    aws s3api abort-multipart-upload --bucket "$bucket" --key "$key" --upload-id "$id"
  done
done
```

### Bucket policies

Even if you lock out the root user with a bucket policy, it is still able to edit/delete the bucket policy via the management console or aws cli.

- https://docs.aws.amazon.com/AmazonS3/latest/dev/example-bucket-policies.html
- https://aws.amazon.com/blogs/security/how-to-restrict-amazon-s3-bucket-access-to-a-specific-iam-role/

Get role id with:
```
aws iam get-role --role-name ROLE_NAME
```

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Deny",
            "Principal": "*",
            "Action": "s3:*",
            "Resource": [
                "arn:aws:s3:::bucketname",
                "arn:aws:s3:::bucketname/*"
            ],
            "Condition": {
                "StringNotLike": {
                    "aws:userId": [
                        "123456789012",
                        "AROAEXAMPLEID:*"
                    ]
                }
            }
        }
    ]
}
```

Deny access to dangerous things:
```
        {
            "Effect": "Deny",
            "Principal": "*",
            "Action": [
                "s3:DeleteBucket",
                "s3:DeleteBucketPolicy",
                "s3:DeleteBucketWebsite",
                "s3:PutBucketAcl",
                "s3:PutBucketCORS",
                "s3:PutBucketObjectLockConfiguration",
                "s3:PutBucketPolicy",
                "s3:PutBucketPublicAccessBlock",
                "s3:PutBucketWebsite",
                "s3:PutReplicationConfiguration"
            ],
            "Resource": "arn:aws:s3:::bucketname"
        },
        {
            "Effect": "Deny",
            "Principal": "*",
            "Action": [
                "s3:PutAccelerateConfiguration",
                "s3:PutAnalyticsConfiguration",
                "s3:PutBucketLogging",
                "s3:PutBucketNotification",
                "s3:PutBucketRequestPayment",
                "s3:PutBucketVersioning",
                "s3:PutEncryptionConfiguration",
                "s3:PutInventoryConfiguration",
                "s3:PutLifecycleConfiguration",
                "s3:PutMetricsConfiguration"
            ],
            "Resource": "arn:aws:s3:::bucketname",
            "Condition": {
                "StringNotLike": {
                    "aws:userId": "123456789012"
                }
            }
        },
        {
            "Effect": "Deny",
            "Principal": "*",
            "Action": [
                "s3:BypassGovernanceRetention",
                "s3:DeleteObject",
                "s3:DeleteObjectVersion",
                "s3:PutObjectAcl",
                "s3:PutObjectLegalHold",
                "s3:PutObjectRetention",
                "s3:PutObjectVersionAcl"
            ],
            "Resource": "arn:aws:s3:::bucketname/*",
            "Condition": {
                "StringNotLike": {
                    "aws:userId": "123456789012"
                }
            }
        },
        {
            "Effect": "Deny",
            "Principal": "*",
            "Action": [
                "s3:*"
            ],
            "Resource": "arn:aws:s3:::bucketname/*",
            "Condition": {
                "StringEquals": {
                    "s3:object-lock-mode": "COMPLIANCE"
                }
            }
        }
```

### Object Lock

`--object-lock-retain-until-date` is given in this format: `2019-01-01T12:00:00.000Z`

Calculate Content-MD5:
```
ruby -rbase64 -rdigest -e 'puts Base64.strict_encode64(Digest::MD5.digest(File.read("file.zip")))'
```

### MFA Delete

Enabling [MFA Delete](https://docs.aws.amazon.com/AmazonS3/latest/dev/Versioning.html#MultiFactorAuthenticationDelete) must be done by the root user and with the aws cli. U2F is not supported.

Log in with the root user and get the MFA serial number from https://console.aws.amazon.com/iam/home#/security_credentials

The MFA serial number is typically in this format:
```
arn:aws:iam::123456789012:mfa/root-account-mfa-device
```

```
# Enable MFA Delete:
aws s3api put-bucket-versioning --profile root --bucket bucketname --versioning-configuration Status=Enabled,MFADelete=Enabled --mfa "mfa_serial_number mfa_code"

# Delete an object version:
aws s3api delete-object --profile root --bucket bucketname --key path/to/file --version-id longversionid --mfa "mfa_serial_number mfa_code"

# Get versioning status:
aws s3api get-bucket-versioning --bucket bucketname

# Disable MFA Delete:
aws s3api put-bucket-versioning --profile root --bucket bucketname --versioning-configuration Status=Enabled,MFADelete=Disabled --mfa "mfa_serial_number mfa_code"
```

It would be great if this could be done with a U2F device. At least you can enable MFA Delete on a few buckets, and then switch back to U2F until you need to delete objects.

## EC2

Change EBS "Delete on Termination" flag after launching instance:
```
aws ec2 modify-instance-attribute --instance-id i-01234567890abcdef --block-device-mappings '[{"DeviceName":"/dev/sda1","Ebs":{"DeleteOnTermination":false}}]'
```

Userdata environment (Ubuntu 16.04). Note that HOME is missing.
```
+ env
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
PWD=/root
SHLVL=1
_=/usr/bin/env
OLDPWD=/
```

### Initialize EBS volume after restoring from snapshot

https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ebs-initialize.html

```
sudo dd if=/dev/xvdf of=/dev/null bs=1M

sudo apt-get install -y fio
sudo fio --filename=/dev/xvdf --rw=read --bs=128k --iodepth=32 --ioengine=libaio --direct=1 --name=volume-initialize
```

Read progress from dd with:
```
sudo pkill -USR1 -n -x dd
```

Faster parallel processing:
```
seq 0 $(($(cat /sys/block/xvdf/size) / (1 << 10))) | xargs -n1 -P32 -I {} sudo dd if=/dev/xvdf of=/dev/null skip={}k count=1 bs=512 > /tmp/initialize.log 2>&1

sudo fio --filename=/dev/xvdf --direct=1 --rw=randread --refill_buffers --norandommap --randrepeat=0 --ioengine=libaio --bs=128k --rwmixread=100 --iodepth=32 --numjobs=4 --group_reporting --name=initialize
```


## RDS

Enable binlog on RDS MySQL without having a replica:
- Enable automated backups (1 day is fine)
- Set `log_bin` and `binlog_format`
- Connect and run `CALL mysql.rds_set_configuration('binlog retention hours', 4);` (168 hours is max (one week))

Get default parameter groups:
```
aws rds describe-db-cluster-parameters --db-cluster-parameter-group-name default.aurora-mysql5.7 > default.aurora-mysql5.7-cluster.json
aws rds describe-db-parameters --db-parameter-group-name default.aurora-mysql5.7 > default.aurora-mysql5.7-instance.json
aws rds describe-db-parameters --db-parameter-group-name default.mysql5.7 > default.mysql5.7-instance.json

aws rds describe-engine-default-cluster-parameters --db-parameter-group-family aurora-mysql5.7 > aurora-mysql5.7-cluster.json
aws rds describe-engine-default-parameters --db-parameter-group-family aurora-mysql5.7 > aurora-mysql5.7-instance.json
aws rds describe-engine-default-parameters --db-parameter-group-family mysql5.7 > mysql5.7.json
```

To get the Aurora version, you have to connect with `mysql` and run `SELECT AURORA_VERSION();`.


## Terraform

```shell
sudo apt install graphviz
brew install gprof2dot
terraform graph | dot -Tpng > graph.png
```

## Acronyms

- PDT: GovCloud
- LCK: Rickenbacker International Airport in Columbus, Ohio.
- DCA: Ronald Reagan Washington National Airport in Washington, D.C.

## Billing

Cost Explorer buckets cost and abbreviates some API operations. For example:
- USW2-CW:MetricMonitorUsage
- USW2-CW:GMD-Metrics
- USW2-CW:Requests

In this case, GMD is short for GetMetricData.

## Swedish region

- Städer: Eskilstuna, Katrineholm, Västerås.
- Vi har höga ambitioner och mål för den här regionen. Vi vet att det finns mycket duktigt folk att anställa här. Dessutom kommer 53 procent av energin i Sverige från förnybara källor, och det passar bra med vår ambition att ha 100 procent förnybar energi.
- Det kommer att vara mellan 50 000 och 80 000 servrar vid varje Availabilty Zone, och det kommer att skapas många jobb. Någon exakt siffra har jag inte, men vi vet att det kommer att ge oss helt nya möjligheter till affärsutveckling och en möjlighet att skapa nya tjänster, säger Darren Mowry.
