#!/usr/bin/env bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/.env
source $SCRIPT_DIR/_common.sh

gcpInit

# Functions

usage() {
  echo "Usage: ${BASH_SOURCE[0]} [default|peering] [create|delete|list]"
  echo "  list - list SA keys"
  echo "  create - create new key"
  echo "  delete - delete oldest key"
  exit $1
}

listKeys() {
  az ad app credential list --id $APP_ID --output table
}

create() {
  log "Creating new key for APP_ID $APP_ID"
  az ad app credential reset --id "$APP_ID" --append --query password  -o tsv
}


delete() {
  log "Deleting oldest key of APP_ID $APP_ID"

  listKeys

  ID=$(az ad app credential list --id $APP_ID --query "sort_by([],&startDateTime) | [0].keyId" --output tsv)
  log "Oldest key id is $ID"

  az ad app credential delete --id "$APP_ID" --key-id "$ID"
  log "The key is deleted"
}

list() {
  log "Listing keys of APP_ID $APP_ID"
  listKeys
}

SA_TYPE=$1
CMD=$2

case "$1" in
default)
  APP_ID=$AZURE_DEFAULT_APP_ID
  ;;
peering)
  APP_ID=$AZURE_PEERING_APP_ID
  ;;
*)
  echo "Unknown SA type '$SA_TYPE'"
  usage 1
esac


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
