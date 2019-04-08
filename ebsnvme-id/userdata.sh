#!/bin/bash -ex
update-alternatives --install /usr/bin/python python /usr/bin/python3 1
curl -fsS -o /sbin/ebsnvme-id https://raw.githubusercontent.com/stefansundin/aws/master/ebsnvme-id/ebsnvme-id
chmod +x /sbin/ebsnvme-id
mkdir -p /usr/lib/udev/
curl -fsS -o /usr/lib/udev/ec2nvme-nsid https://raw.githubusercontent.com/stefansundin/aws/master/ebsnvme-id/ec2nvme-nsid
chmod +x /usr/lib/udev/ec2nvme-nsid
curl -fsS -o /etc/udev/rules.d/70-ec2-nvme-devices.rules https://raw.githubusercontent.com/stefansundin/aws/master/ebsnvme-id/70-ec2-nvme-devices.rules
udevadm trigger
