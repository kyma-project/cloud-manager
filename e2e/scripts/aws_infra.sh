#!/usr/bin/env bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/.env
source $SCRIPT_DIR/_common.sh
source $SCRIPT_DIR/_common-aws.sh

awsInit

createUser() {
  local user_name=$1

  if ! aws iam get-user --user-name "$user_name" > /dev/null 2>&1 ; then
    log "User $user_name does not exist, creating it now..."
    aws iam create-user --user-name "$user_name" > /dev/null
    log "User $user_name is created"
  else
    log "User $user_name already exists"
  fi

  return 0
}

createPolicy() {
  local policy_name=$1
  local policy_file=$2
  local policy_arn
  policy_arn=$(aws iam list-policies --scope Local --query "Policies[?PolicyName=='$policy_name'].Arn" --output text)
  if [[ -z "$policy_arn" ]]; then
    log "Policy $policy_name does not exist, creating it now from $policy_file ..."
    policy_arn=$(aws iam create-policy --policy-name "$policy_name" --policy-document "file://$policy_file" --query 'Policy.Arn' --output text)
    log "Policy $policy_name is created with ARN $policy_arn"
  else
    log "Policy $policy_name exists with ARN $policy_arn"
    local version_count
    version_count=$(aws iam list-policy-versions --policy-arn "$policy_arn" | jq -r ".Versions | length")
    if [[ "$version_count" -ge 4 ]]; then
      local OLDEST_VERSION_ID
      OLDEST_VERSION_ID=$(aws iam list-policy-versions --policy-arn "$policy_arn" | jq -r '.Versions | sort_by(.CreateDate) | .[0] | .VersionId')
      log "Policy $policy_name has $version_count versions, deleting oldest version $OLDEST_VERSION_ID ..."
      aws iam delete-policy-version --policy-arn "$policy_arn" --version-id "$OLDEST_VERSION_ID" > /dev/null
    fi

    log "Updating policy $policy_name from $policy_file..."
    aws iam create-policy-version --policy-arn "$policy_arn" --policy-document "file://$policy_file" --set-as-default > /dev/null
  fi

  return 0
}

createCMRole() {
  local role_name=$1
  local sa_name=$2
  local trust_template=$3
  local file
  file=$(mktemp)
  trap "rm -f \"$file\"" EXIT
  jq ".Statement[0].Principal.AWS |= \"arn:aws:iam::${AWS_ACCOUNT}:user/${sa_name}\"" $trust_template > "$file"

  createRole "$role_name" "$file"
  rm -f "$file"

  return 0
}

createRole() {
  local role_name=$1
  local trust_file=$2

  local role_arn
  role_arn=$(aws iam get-role --role-name "$role_name" --query 'Role.Arn' --output text 2>/dev/null || true)

  if [[ -z "$role_arn" ]]; then
    log "Role $role_name does not exist, creating it now from $trust_file ..."
    role_arn=$(aws iam create-role --role-name "$role_name" --assume-role-policy-document "file://$trust_file" --query 'Role.Arn' --output text)
    log "Role $role_name is created with ARN $role_arn"
  else
    log "Role $role_name exists with ARN $role_arn, updating it now from $trust_file..."
    aws iam update-assume-role-policy --role-name "$role_name" --policy-document "file://$trust_file" > /dev/null
  fi

  return 0
}

attachPolicyToUser() {
  local user_name=$1
  local policy_name=$2
  local policy_arn="arn:aws:iam::${AWS_ACCOUNT}:policy/${policy_name}"

  log "Attaching policy $policy_arn to user $user_name"

  aws iam attach-user-policy --user-name "$user_name" --policy-arn "$policy_arn" > /dev/null

  return 0
}

attachPolicyToRole() {
  local role_name=$1
  local policy_name=$2
  local policy_arn
  case "$policy_name" in
  (arn:*) policy_arn="$policy_name" ;; # arm already given
  (*) policy_arn="arn:aws:iam::${AWS_ACCOUNT}:policy/${policy_name}" ;; # policy name given, create arm out of it
  esac

  log "Attaching policy $policy_arn to role $role_name"

  aws iam attach-role-policy --role-name "$role_name" --policy-arn "$policy_arn" > /dev/null

  return 0
}


createUser ${SA_NAME_DEFAULT}
createUser ${SA_NAME_PEERING}
createPolicy ${POLICY_NAME_DEFAULT} ${POLICY_FILE_DEFAULT}
createPolicy ${POLICY_NAME_PEERING} ${POLICY_FILE_PEERING}
createPolicy ${POLICY_NAME_CM_ASSUME_ROLE} ${POLICY_FILE_CM_ASSUME_ROLE}

createCMRole ${ROLE_NAME_DEFAULT} ${SA_NAME_DEFAULT} ${TRUST_FILE_ALLOW_CLOUD_MANAGER_ASSUME_ROLE}
createCMRole ${ROLE_NAME_PEERING} ${SA_NAME_PEERING} ${TRUST_FILE_ALLOW_CLOUD_MANAGER_ASSUME_ROLE}

createRole ${ROLE_NAME_BACKUP_SERVICE} ${TRUST_FILE_BACKUP_SERVICE}

attachPolicyToUser ${SA_NAME_DEFAULT} ${POLICY_NAME_CM_ASSUME_ROLE}
attachPolicyToUser ${SA_NAME_PEERING} ${POLICY_NAME_CM_ASSUME_ROLE}

attachPolicyToRole ${ROLE_NAME_DEFAULT} ${POLICY_NAME_DEFAULT}
attachPolicyToRole ${ROLE_NAME_PEERING} ${POLICY_NAME_PEERING}

attachPolicyToRole ${ROLE_NAME_BACKUP_SERVICE} "arn:aws:iam::aws:policy/service-role/AWSBackupServiceRolePolicyForBackup"
attachPolicyToRole ${ROLE_NAME_BACKUP_SERVICE} "arn:aws:iam::aws:policy/service-role/AWSBackupServiceRolePolicyForRestores"
