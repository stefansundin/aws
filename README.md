Change EBS "Delete on Termination" flag after launching instance:
```shell
aws ec2 modify-instance-attribute --instance-id i-01234567890abcdef --block-device-mappings '[{"DeviceName":"/dev/sda1","Ebs":{"DeleteOnTermination":false}}]'
```

Get key identity:
```shell
AWS_ACCESS_KEY_ID=AKIA.. AWS_SECRET_ACCESS_KEY=... aws sts get-caller-identity
```

```shell
sudo apt install graphviz
brew install gprof2dot
terraform graph | dot -Tpng > graph.png
```
