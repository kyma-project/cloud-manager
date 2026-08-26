#!/usr/bin/env bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/.env
source $SCRIPT_DIR/_common.sh
source $SCRIPT_DIR/_common-alicloud.sh

alicloudInit

createUser() {
  local user_name=$1

  local exists
  exists=$(aliyun ram GetUser --UserName "$user_name" 2>/dev/null | jq -r '.User.UserName // empty')
  if [[ -z "$exists" ]]; then
    log "RAM user $user_name does not exist, creating it now..."
    aliyun ram CreateUser --UserName "$user_name" > /dev/null
    log "RAM user $user_name is created"
  else
    log "RAM user $user_name already exists"
  fi

  return 0
}

createPolicy() {
  local policy_name=$1
  local policy_file=$2

  local exists
  exists=$(aliyun ram GetPolicy --PolicyType Custom --PolicyName "$policy_name" 2>/dev/null | jq -r '.Policy.PolicyName // empty')
  if [[ -z "$exists" ]]; then
    log "RAM policy $policy_name does not exist, creating it now from $policy_file..."
    local policy_doc
    policy_doc=$(cat "$policy_file")
    aliyun ram CreatePolicy --PolicyName "$policy_name" --PolicyDocument "$policy_doc" > /dev/null
    log "RAM policy $policy_name is created"
  else
    log "RAM policy $policy_name already exists, updating it..."
    local policy_doc
    policy_doc=$(cat "$policy_file")
    # AliCloud limits policies to 5 versions (1 default + 4 non-default).
    # Prune oldest non-default versions to ensure at most 3 exist before creating
    # a new one (keeping up to 4 total non-default is safe; head -n -3 deletes
    # the oldest entry only when 4 are already present).
    local versions
    versions=$(aliyun ram ListPolicyVersions --PolicyType Custom --PolicyName "$policy_name" \
      | jq -r '.PolicyVersions.PolicyVersion[] | select(.IsDefaultVersion == false) | .VersionId' \
      | sort | head -n -3)
    for v in $versions; do
      log "Deleting old policy version $v..."
      aliyun ram DeletePolicyVersion --PolicyName "$policy_name" --VersionId "$v" > /dev/null
    done
    aliyun ram CreatePolicyVersion --PolicyName "$policy_name" --PolicyDocument "$policy_doc" --SetAsDefault true > /dev/null
  fi

  return 0
}

attachPolicyToUser() {
  local user_name=$1
  local policy_name=$2

  log "Attaching policy $policy_name to RAM user $user_name"
  # AttachPolicyToUser errors if already attached; suppress to allow re-runs.
  aliyun ram AttachPolicyToUser --PolicyType Custom --PolicyName "$policy_name" --UserName "$user_name" > /dev/null 2>&1 || true

  return 0
}

createUser "$SA_NAME_DEFAULT"
createPolicy "$POLICY_NAME_DEFAULT" "$POLICY_FILE_DEFAULT"
attachPolicyToUser "$SA_NAME_DEFAULT" "$POLICY_NAME_DEFAULT"
