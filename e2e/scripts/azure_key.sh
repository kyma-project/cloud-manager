#!/usr/bin/env bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/.env
source $SCRIPT_DIR/_common.sh
source $SCRIPT_DIR/_common-azure.sh

azureInit

# Functions

usage() {
  echo "Usage: ${BASH_SOURCE[0]} [create|delete|list]"
  echo "  list - list SA keys"
  echo "  create - create new key"
  echo "  delete - delete oldest key"
  local exit_code
  exit_code=${1:-0}
  exit $exit_code
}

listKeys() {
  az ad app credential list --id $APP_ID --output table
  return 0
}

create() {
  log "Creating new key for APP_ID $APP_ID"
  local secret
  local tenant
  local subscription
  secret=$(az ad app credential reset --id "$APP_ID" --append --query password --only-show-errors -o tsv)
  tenant=$(az account show --query tenantId --output tsv)
  subscription=$(az account show --query id --output tsv)
  putCredentialKeyVal "clientID" "$APP_ID"
  putCredentialKeyVal "clientSecret" "$secret"
  putCredentialKeyVal "subscriptionID" "$subscription"
  putCredentialKeyVal "tenantID" "$tenant"
  saveCredentialsToGarden "$AZURE_GARDEN_SECRET"

  return 0
}


delete() {
  log "Deleting oldest key of APP_ID $APP_ID"

  listKeys

  ID=$(az ad app credential list --id $APP_ID --query "sort_by([],&startDateTime) | [0].keyId" --output tsv)
  log "Oldest key id is $ID"

  az ad app credential delete --id "$APP_ID" --key-id "$ID"
  log "The key is deleted"

  return 0
}

list() {
  log "Listing keys of APP_ID $APP_ID"
  listKeys
  return 0
}

APP_ID=$AZURE_DEFAULT_APP_ID
CMD=$1

case "$CMD" in
create)
  create
  ;;
delete)
  delete
  ;;
list)
  list
  ;;
*)
  echo "Unknown command '$CMD'"
  usage 1
esac
