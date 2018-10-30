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
