Use the following scripts to map Nitro NVME device names to EC2 device names and EBS volume ids.
- https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/nvme-ebs-volumes.html

Taken from:
- Amazon Linux 2 AMI
- ami-01e3b8c3a51e88954
- us-east-1

Files:
- /sbin/ebsnvme-id
  - Updated to be compatible with both Python 2 and 3.
- /usr/lib/udev/ec2nvme-nsid
- /etc/udev/rules.d/70-ec2-nvme-devices.rules
