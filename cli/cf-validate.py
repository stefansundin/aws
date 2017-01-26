#!/usr/bin/env python
# Validate a CloudFormation template.
import sys, boto3, botocore

if __name__ == "__main__":
    client = boto3.client("cloudformation")
    for fn in sys.argv[1:]:
        with open(fn, "r") as f:
            body = f.read()
            try:
                client.validate_template(TemplateBody=body)
            except botocore.exceptions.ClientError as e:
                sys.stderr.write(e.response["Error"]["Message"] + "\n")
            else:
                sys.stderr.write("The file %s is valid!\n" % fn)
