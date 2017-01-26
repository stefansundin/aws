Handy awscli aliases:
- federate: Assume a role and create a sign-in link to the AWS console.
- s3-cat: Output the contents of a file on S3.
- cf-validate: Validate a CloudFormation template.
- cf-diff: Diff a stack against a template file.
- cf-dump: Download info about a stack (useful to "backup" a stack along with its parameters before you delete it).

# Python 3 setup

Install Python pre-requisites.

```bash
brew install python3
python3 -m virtualenv venv -p python3
(source venv/bin/activate && pip3 install --upgrade -r requirements.txt)
```

Add awscli alias. `cat ~/.aws/cli/alias`

```
[toplevel]

whoami = sts get-caller-identity

federate =
  !f() {
    DIR=~/src/aws/cli
    source "$DIR/venv/bin/activate"
    "$DIR/federate.py" $*
  }; f

s3-cat =
  !f() {
    FILE=$(mktemp)
    aws s3 cp "$1" "$FILE"
    cat "$FILE"
    rm "$FILE"
  }; f

cf-validate =
  !f() {
    DIR=~/src/aws/cli
    source "$DIR/venv/bin/activate"
    "$DIR/cf-validate.py" $*
  }; f

cf-diff =
  !f() {
    DIR=~/src/aws/cli
    source "$DIR/venv/bin/activate"
    "$DIR/cf-diff.py" $*
  }; f

cf-dump =
  !f() {
    DIR=~/src/aws/cli
    "$DIR/cf-dump.sh" $*
  }; f
```

Example commands:

```bash
aws federate admin
aws s3-cat http://s3.amazonaws.com/myrandombucket/logs/build.log
aws cf-validate webservers.yml
aws cf-diff prod-webservers webservers.yml
AWS_REGION=us-west-2 aws cf-diff stage-webservers webservers.yml
AWS_PROFILE=test aws cf-diff stage-webservers webservers.yml
aws cf-dump prod-webservers
```

Example federate bash aliases:

```bash
alias aws-admin="aws federate admin"
alias aws-admin="aws federate arn:aws:iam::123456789012:role/AdministratorRole arn:aws:iam::123456789012:mfa/username"
```
