Change EBS "Delete on Termination" flag after launching instance:
```shell
aws ec2 modify-instance-attribute --instance-id i-01234567890abcdef --block-device-mappings '[{"DeviceName":"/dev/sda1","Ebs":{"DeleteOnTermination":false}}]'
```

Get key identity:
```shell
AWS_ACCESS_KEY_ID=AKIA.. AWS_SECRET_ACCESS_KEY=... aws sts get-caller-identity
```

Find and abort stale S3 multipart uploads:
```shell
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

```shell
sudo apt install graphviz
brew install gprof2dot
terraform graph | dot -Tpng > graph.png
```
