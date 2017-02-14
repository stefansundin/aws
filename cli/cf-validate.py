#!/usr/bin/env python
# Validate a CloudFormation template.
import sys, boto3, botocore

if __name__ == "__main__":
    client = boto3.client("cloudformation")
    for fn in sys.argv[1:]:
        with open(fn, "r") as f:
            body = f.read()
            if body.find("AWSTemplateFormatVersion") == -1:
                sys.stderr.write("The file %s doesn't look like a CloudFormation template!\n" % fn)
            try:
                client.validate_template(TemplateBody=body)
            except botocore.exceptions.ClientError as e:
                sys.stderr.write("%s: %s\n" % (fn, e.response["Error"]["Message"]))
            else:
                sys.stderr.write("The file %s is valid!\n" % fn)
