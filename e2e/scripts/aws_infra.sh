#!/usr/bin/env bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/.env
source $SCRIPT_DIR/_common.sh
source $SCRIPT_DIR/_common-aws.sh

awsInit

#AWS_ACCOUNT=$(aws sts get-caller-identity --query Account --output text)
#
#log "Running on AWS account $AWS_ACCOUNT"

