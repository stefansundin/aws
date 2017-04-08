#!/usr/bin/env python
# This script lets you easily assume a role in your terminal.

import sys, os, boto3

mfa_serial = None

if len(sys.argv) == 2:
    if sys.argv[1].startswith("arn:aws:iam:"):
        role_arn = sys.argv[1]
    else:
        import configparser
        config = configparser.ConfigParser()
        config.read([os.environ["HOME"]+"/.aws/credentials"])
        if config.has_option(sys.argv[1], "source_profile"):
            source_profile = config.get(sys.argv[1], "source_profile")
            boto3.setup_default_session(profile_name=source_profile)
            role_arn = config.get(sys.argv[1], "role_arn")
            mfa_serial = config.get(sys.argv[1], "mfa_serial", fallback=None)
        else:
            response = boto3.client("sts").get_caller_identity()
            role_arn = "arn:aws:iam::"+response["Account"]+":role/"+sys.argv[1]
elif len(sys.argv) == 3:
    role_arn = sys.argv[1]
    mfa_serial = sys.argv[2]
else:
    sys.stderr.write("Insufficient arguments.\n")
    sys.stderr.write("Usage: %s <profile>" % sys.argv[0]+"\n")
    sys.stderr.write("Usage: %s <role_name>" % sys.argv[0]+"\n")
    sys.stderr.write("Usage: %s <role_arn>" % sys.argv[0]+"\n")
    sys.stderr.write("Usage: %s <role_arn> <mfa_arn>" % sys.argv[0]+"\n")
    sys.exit(1)

# This is what will show up as the username CloudTrail events
if mfa_serial:
    session_name = mfa_serial.split("/")[-1]
else:
    session_name = os.environ["USER"]

kwargs = {
    "RoleArn": role_arn,
    "RoleSessionName": session_name,
    # "DurationSeconds": 43200,  # 12 hours
}

if mfa_serial:
    kwargs["SerialNumber"] = mfa_serial
    kwargs["TokenCode"] = raw_input("Enter MFA code: ")

sts = boto3.client("sts")
role = sts.assume_role(**kwargs)

print("export AWS_ACCESS_KEY_ID="+role["Credentials"]["AccessKeyId"])
print("export AWS_SECRET_ACCESS_KEY="+role["Credentials"]["SecretAccessKey"])
print("export AWS_SESSION_TOKEN="+role["Credentials"]["SessionToken"])

if sys.stdout.isatty():
    sys.stderr.write("\n")

sys.stderr.write("Assumed role: "+role["AssumedRoleUser"]["Arn"]+"\n")
sys.stderr.write("Expires at "+str(role["Credentials"]["Expiration"])+"\n")
sys.stderr.write("\n")

if sys.stdout.isatty():
    sys.stderr.write("You can quickly source these variables by running:\n")
    sys.stderr.write("eval \"$("+" ".join(sys.argv)+")\"\n")
    sys.stderr.write("\n")
sys.stderr.write("Undo with:\n")
sys.stderr.write("unset AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY AWS_SESSION_TOKEN\n")
sys.stderr.write("\n")
sys.stderr.write("Verify with:\n")
sys.stderr.write("aws sts get-caller-identity\n")
