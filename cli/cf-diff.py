#!/usr/bin/env python
# Diff a CloudFormation stack with a local file.
import argparse, sys, boto3, botocore, difflib, json
from clint.textui import colored

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-r", "--region", help="AWS region")
    parser.add_argument("stack", help="the CloudFormation stack to compare")
    parser.add_argument("file", help="the local file to compare")
    args = parser.parse_args()

    if args.region:
        boto3.setup_default_session(region_name=args.region)

    # get remote template
    try:
        client = boto3.client("cloudformation")
        templ = client.get_template(StackName=args.stack)
        remote_template = templ["TemplateBody"]
        if isinstance(remote_template, dict):
            remote_template = json.dumps(remote_template, indent=2, sort_keys=True)
    except botocore.exceptions.ClientError as e:
        print(e)
        sys.exit(1)

    # get local template
    with open(args.file, "r") as f:
        local_template = f.read()
        if args.file.endswith(".json"):
            local_template = json.dumps(json.loads(local_template), indent=2, sort_keys=True)

    # diff them!
    diff = difflib.unified_diff(
        remote_template.splitlines(),
        local_template.splitlines(),
        fromfile="AWS STACK: " + args.stack,
        tofile="LOCAL: " + args.file
    )
    for line in diff:
        if line.startswith("+"):
            print(colored.green(line))
        elif line.startswith("-"):
            print(colored.red(line))
        else:
            print(line)
