#!/usr/bin/env python
# https://github.com/stefansundin/aws/blob/master/cli/federate.py

# This script lets you assume a role in the AWS console with a session duration
# that is longer than one hour (max 12 hours).

# Example bash aliases:
# alias aws-admin="aws federate admin"
# alias aws-admin="aws federate arn:aws:iam::123456789012:role/ReadOnlyRole"
# alias aws-admin="aws federate arn:aws:iam::123456789012:role/AdministratorRole arn:aws:iam::123456789012:mfa/username"
# alias aws-admin="~/src/aws/cli/federate.py arn:aws:iam::123456789012:role/AdministratorRole arn:aws:iam::123456789012:mfa/username"

# Docs:
# http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_enable-console-custom-url.html
# https://aws.amazon.com/blogs/security/enable-your-federated-users-to-work-in-the-aws-management-console-for-up-to-12-hours/
# http://boto.cloudhackers.com/en/latest/ref/sts.html

import sys, os, urllib, json, requests
from boto.sts import STSConnection

dest = "https://console.aws.amazon.com/console/home"

source_profile = None
mfa_serial = None

if len(sys.argv) == 2:
    if sys.argv[1].startswith("arn:aws:iam:"):
        role_arn = sys.argv[1]
    else:
        import configparser
        config = configparser.ConfigParser()
        config.read([os.environ["HOME"]+"/.aws/credentials"])
        if not config.has_option(sys.argv[1], "source_profile"):
            print("Profile "+sys.argv[1]+" not found in ~/.aws/credentials.")
            sys.exit(1)
        source_profile = config.get(sys.argv[1], "source_profile")
        role_arn = config.get(sys.argv[1], "role_arn")
        mfa_serial = config.get(sys.argv[1], "mfa_serial", fallback=None)
        region = config.get(sys.argv[1], "region", fallback=None)
        if region:
            dest += "?region=" + region
elif len(sys.argv) == 3:
    role_arn = sys.argv[1]
    mfa_serial = sys.argv[2]
else:
    print("Insufficient arguments.")
    print("Usage: %s <profile>" % sys.argv[0])
    print("Usage: %s <role_arn>" % sys.argv[0])
    print("Usage: %s <role_arn> <mfa_arn>" % sys.argv[0])
    sys.exit(1)

# This is what will show up as the username in the ConsoleLogin event in CloudTrail
if mfa_serial:
    session_name = mfa_serial.split("/")[-1]
else:
    session_name = os.environ["USER"]

kwargs = {
    "role_session_name": session_name,
    "role_arn": role_arn,
}

if mfa_serial:
    kwargs["mfa_serial_number"] = mfa_serial
    kwargs["mfa_token"] = raw_input("Enter MFA code: ")

# Call AssumeRole to get temporary access keys for the federated user
if source_profile:
    sts_connection = STSConnection(profile_name=source_profile)
else:
    sts_connection = STSConnection()
assumed_role_object = sts_connection.assume_role(**kwargs)

# Make request to AWS federation endpoint to get sign-in token
r = requests.get("https://signin.aws.amazon.com/federation", params={
    "Action": "getSigninToken",
    "SessionDuration": "43200",  # 12 hours
    "Session": json.dumps({
        "sessionId": assumed_role_object.credentials.access_key,
        "sessionKey": assumed_role_object.credentials.secret_key,
        "sessionToken": assumed_role_object.credentials.session_token,
    }),
})
data = r.json()
# print(json.dumps(data, indent=2))

# Create URL where users can sign-in to the console
# This URL must be used within 15 minutes
url = "https://signin.aws.amazon.com/federation"
url += "?Action=login"
url += "&Issuer=https://github.com/stefansundin/aws/blob/master/cli/federate.py"
url += "&Destination=" + urllib.quote_plus(dest)
url += "&SigninToken=" + data["SigninToken"]

# Print URL
print(url)
