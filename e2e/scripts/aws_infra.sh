#!/usr/bin/env bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/.env
source $SCRIPT_DIR/_common.sh
source $SCRIPT_DIR/_common-aws.sh

awsInit

createUser() {
  local USER_NAME=$1

  if ! aws iam get-user --user-name "$USER_NAME" > /dev/null 2>&1 ; then
    log "User $USER_NAME does not exist, creating it now..."
    aws iam create-user --user-name "$USER_NAME" > /dev/null
    log "User $USER_NAME is created"
  else
    log "User $USER_NAME already exists"
  fi
}

createPolicy() {
  local POLICY_NAME=$1
  local POLICY_FILE=$2
  local POLICY_ARN
  POLICY_ARN=$(aws iam list-policies --scope Local --query "Policies[?PolicyName=='$POLICY_NAME'].Arn" --output text)
  if [ -z "$POLICY_ARN" ]; then
    log "Policy $POLICY_NAME does not exist, creating it now from $POLICY_FILE ..."
    POLICY_ARN=$(aws iam create-policy --policy-name "$POLICY_NAME" --policy-document "file://$POLICY_FILE" --query 'Policy.Arn' --output text)
    log "Policy $POLICY_NAME is created with ARN $POLICY_ARN"
  else
    log "Policy $POLICY_NAME exists with ARN $POLICY_ARN"
    local VERSION_COUNT
    VERSION_COUNT=$(aws iam list-policy-versions --policy-arn "$POLICY_ARN" | jq -r ".Versions | length")
    if [ "$VERSION_COUNT" -ge 4 ]; then
      local OLDEST_VERSION_ID
      OLDEST_VERSION_ID=$(aws iam list-policy-versions --policy-arn "$POLICY_ARN" | jq -r '.Versions | sort_by(.CreateDate) | .[0] | .VersionId')
      log "Policy $POLICY_NAME has $VERSION_COUNT versions, deleting oldest version $OLDEST_VERSION_ID ..."
      aws iam delete-policy-version --policy-arn "$POLICY_ARN" --version-id "$OLDEST_VERSION_ID" > /dev/null
    fi

    log "Updating policy $POLICY_NAME from $POLICY_FILE..."
    aws iam create-policy-version --policy-arn "$POLICY_ARN" --policy-document "file://$POLICY_FILE" --set-as-default > /dev/null
  fi
}

createCMRole() {
  local ROLE_NAME=$1
  local SA_NAME=$2
  local TRUST_TEMPLATE=$3
  local FILE
  FILE=$(mktemp)
  trap "rm -f \"$FILE\"" EXIT
  jq ".Statement[0].Principal.AWS |= \"arn:aws:iam::${AWS_ACCOUNT}:user/${SA_NAME}\"" $TRUST_TEMPLATE > "$FILE"

  createRole "$ROLE_NAME" "$FILE"
  rm -f "$FILE"
}

createRole() {
  local ROLE_NAME=$1
  local TRUST_FILE=$2

  local ROLE_ARN
  ROLE_ARN=$(aws iam get-role --role-name "$ROLE_NAME" --query 'Role.Arn' --output text 2>/dev/null || true)

  if [ -z "$ROLE_ARN" ]; then
    log "Role $ROLE_NAME does not exist, creating it now from $TRUST_FILE ..."
    ROLE_ARN=$(aws iam create-role --role-name "$ROLE_NAME" --assume-role-policy-document "file://$TRUST_FILE" --query 'Role.Arn' --output text)
    log "Role $ROLE_NAME is created with ARN $ROLE_ARN"
  else
    log "Role $ROLE_NAME exists with ARN $ROLE_ARN, updating it now from $TRUST_FILE..."
    aws iam update-assume-role-policy --role-name "$ROLE_NAME" --policy-document "file://$TRUST_FILE" > /dev/null
  fi
}

attachPolicyToUser() {
  local USER_NAME=$1
  local POLICY_NAME=$2
  local POLICY_ARN="arn:aws:iam::${AWS_ACCOUNT}:policy/${POLICY_NAME}"

  log "Attaching policy $POLICY_ARN to user $USER_NAME"

  aws iam attach-user-policy --user-name "$USER_NAME" --policy-arn "$POLICY_ARN" > /dev/null
}

attachPolicyToRole() {
  local ROLE_NAME=$1
  local POLICY_NAME=$2
  local POLICY_ARN
  case "$POLICY_NAME" in
  (arn:*) POLICY_ARN="$POLICY_NAME" ;; # arm already given
  (*) POLICY_ARN="arn:aws:iam::${AWS_ACCOUNT}:policy/${POLICY_NAME}" ;; # policy name given, create arm out of it
  esac

  log "Attaching policy $POLICY_ARN to role $ROLE_NAME"

  aws iam attach-role-policy --role-name "$ROLE_NAME" --policy-arn "$POLICY_ARN" > /dev/null
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
