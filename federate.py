#!/usr/bin/env python

# This script lets you assume a role in the AWS console with a session duration
# that is longer than one hour (max 12 hours).
# It is convenient to call this script from a script or bash function, e.g.:
# alias aws-admin="federate.py arn:aws:iam::123456789012:role/AdministratorRole arn:aws:iam::123456789012:mfa/username"

# Install prerequisites:
# pip install requests boto

# Docs:
# http://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_enable-console-custom-url.html
# https://aws.amazon.com/blogs/security/enable-your-federated-users-to-work-in-the-aws-management-console-for-up-to-12-hours/
# http://boto.cloudhackers.com/en/latest/ref/sts.html

import sys, urllib, json, requests
from boto.sts import STSConnection

if len(sys.argv) < 3:
    print("Insufficient arguments.")
    print("Usage: %s <role_arn> <mfa_arn>" % sys.argv[0])
    sys.exit(1)

# Call AssumeRole to get temporary access keys for the federated user
sts_connection = STSConnection()
assumed_role_object = sts_connection.assume_role(
    role_session_name="federate.py",
    role_arn=sys.argv[1],
    mfa_serial_number=sys.argv[2],
    mfa_token=raw_input("Enter MFA code: ")
)

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
url += "&Issuer=federate.py"
url += "&Destination=" + urllib.quote_plus("https://console.aws.amazon.com/")
url += "&SigninToken=" + data["SigninToken"]

# Print URL
print(url)
