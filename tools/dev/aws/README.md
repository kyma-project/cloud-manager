# AWS

> [!CAUTION]
> Internal teams should not create new subscriptions but share already created ones!

## AWS Accounts

Ensure two AWS accounts:
* CloudResources - where CloudResources principal is defined
* SKR - where Garden created SKR cluster


## AWS Azure SSO

Ensure you are using AWS Azure SSO 
https://wiki.one.int.sap/wiki/pages/viewpage.action?pageId=3398468150
https://github.com/aws-azure-login/aws-azure-login

### The aws-azure-login Installation

The aws-azure-login prompt in *Option B: Install Only for Current User* asks you to set the npm prefix to your home directory.
Note that using npm prefix is not compatible with nvm. It's recommended not to use nvm and set the npm prefix as aws-azure-login suggests. Using nvm will probably get you the error:
```
TypeError: Cannot read properties of undefined (reading 'load')
    at Object._parseRolesFromSamlResponse (..../node_modules/aws-azure-login/lib/login.js:663:40)
```
related to `cheerio.default` not defined. If you find a way to use aws-azure-login with nvm, please update this doc. Thanks!

### How to create the AWS profile

Go to the https://myapps.microsoft.com/ and find the AWS app URL that is in the format 
like `https://launcher.myapps.microsoft.com/api/signin/<app_id_uri>?tenantId=<tenant_id>`.

Based on that app URL make the AWS config (usuallt at `~/.aws/config`) profile like this:
```
[profile <name_of_profile>]
azure_tenant_id=<tenant_id>
azure_app_id_uri=<app_id_uri>
azure_default_username=<sap_email_id>
azure_default_role_arn=
azure_default_duration_hours=8 # Range between 4-8 hours
azure_default_remember_me=true
region=eu-central-1
```

Choose that profile
```shell
export AWS_PROFILE=<name_of_profile>
```

Run 
```shell
aws-azure-login --mode gui --no-sandbox
```

A Chromium window will open, login with your SAP credentials, do the SFA, and window should close and aws-azure-login process terminate w/out errors. 

Test the access:
```shell
aws sts get-caller-identity 
```

## AWS CLI Profiles

Create separate AWS CLI profiles, one for CloudResources, other for SKR account. Name them by the MC subscription name (do not use CloudResources and SKR since those will be profile names for the technical principal used by the operator).

## Security setup

In the CloudResources account create user `cloudresources` and obtain its access key and secret. Use it for debugging and local runs of the operator. 

In SKR account create role allowing `sts:AssumeRole` for the CloudResources:user/cloudresources principal:

```shell
skr_trust_policy.json
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::CLOUDRESOURCES_AWS_ACCOUNT_ID:user/cloudresources"
      },
    "Action": "sts:AssumeRole"
  }]
}

AWS_PROFILE=skr aws iam create-role \
  --role-name CrossAccountPowerUser \
  --assume-role-policy-document file://./skr_trust_policy.json 
```

Attach that newly created `CrossAccountPowerUser` appropriate access in the SKR account, ie:

```shell
AWS_PROFILE=skr aws iam attach-role-policy 
  --role-name CrossAccountPowerUser 
  --policy-arn arn:aws:iam::aws:policy/AdministratorAccess 
```

In the CloudResources account allow user `cloudresources` to assume SKR's roles `CrossAccountPowerUser`:
```shell
cloudresources_assume_role_prod.json
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "sts:AssumeRole",
    "Resource": "arn:aws:iam::*:role/CrossAccountPowerUser"
  }
}

AWS_PROFILE=cloudresources aws iam create-policy 
  --policy-name SkrAccountAccess 
  --policy-document file://./cloudresources_assume_role_prod.json
```

Attach newly created `SkrAccountAccess` role to the `cloudresources` principal:
```shell
AWS_PROFILE=cloudresources aws iam attach-user-policy 
  --user-name bjwagner 
  --policy-arn arn:aws:iam::CLOUDRESOURCES_AWS_ACCOUNT_ID:policy/SkrAccountAccess
```

Create AWS CLI profiles for these two accounts using the service principal

```shell
cat ~/.aws/config

[profile cloudresources]
region = SOME-REGION

[profile skr]
role_arn = arn:aws:iam::SKR_AWS_ACCOUNT_ID:role/CrossAccountPowerUser
source_profile = cloudresources
region = SOME-REGION
```

Verify it works
```shell
% AWS_PROFILE=cloudresources aws sts get-caller-identity
{
    "UserId": "XXXXXXXXXX",
    "Account": "CLOUDRESOURCES_AWS_ACCOUNT_ID",
    "Arn": "arn:aws:iam::CLOUDRESOURCES_AWS_ACCOUNT_ID:user/cloudresources"
}

% AWS_PROFILE=skr aws sts get-caller-identity
{
    "UserId": "XXXXXXXXXX:botocore-session-YYYYYYYY",
    "Account": "SKR_AWS_ACCOUNT_ID",
    "Arn": "arn:aws:sts::SKR_AWS_ACCOUNT_ID:assumed-role/CrossAccountPowerUser/botocore-session-YYYYYYYY"
}
```
