#!/usr/bin/env bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/.env
source $SCRIPT_DIR/_common.sh

azureInit

# Functions

createRole() {
  ROLE_NAME=$1
  ROLE_FILE=$2
  ROLE_EXISTS=$(az role definition list --name "$ROLE_NAME" --subscription $AZURE_SUBSCRIPTION_ID --query "[0].id" -o tsv 2>/dev/null || echo "")



  TEMP_PERMISSION_FILE="/tmp/$ROLE_NAME.json"
  TEMP_PERMISSION_FILE=$(mktemp)

  log "Crating temp permission file $TEMP_PERMISSION_FILE"

  jq -n \
  --arg name "$ROLE_NAME" \
  --arg scope "/subscriptions/$AZURE_SUBSCRIPTION_ID" \
  --slurpfile permissions "$ROLE_FILE" \
'{
  "name": $name,
  "assignableScopes": [$scope],
  "actions": $permissions[0].actions,
  "dataActions": $permissions[0].dataActions
}' > $TEMP_PERMISSION_FILE


  if [[ -z "$ROLE_EXISTS" ]]; then
    log "Role $ROLE_NAME does not exist, creating it now from $ROLE_FILE ..."
    az role definition create  --role-definition $TEMP_PERMISSION_FILE  --subscription $AZURE_SUBSCRIPTION_ID
  else
    log "Role $ROLE_NAME exist, updating it now from $ROLE_FILE ..."
    az role definition update  --role-definition $TEMP_PERMISSION_FILE  --subscription $AZURE_SUBSCRIPTION_ID
  fi
}

createRoleAssignment() {
  SA_NAME=$1
  ROLE_NAME=$2

  az role assignment create --assignee $SA_NAME \
  --role $ROLE_NAME \
  --scope "/subscriptions/$AZURE_SUBSCRIPTION_ID"
}

# Main

createRole "$ROLE_NAME_DEFAULT" "$ROLE_FILE_DEFAULT"
createRole "$ROLE_NAME_PEERING" "$ROLE_FILE_PEERING"

createRoleAssignment "$AZURE_DEFAULT_APP_ID" "$ROLE_NAME_DEFAULT"
createRoleAssignment "$AZURE_PEERING_APP_ID" "$ROLE_NAME_PEERING"