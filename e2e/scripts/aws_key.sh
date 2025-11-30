#!/usr/bin/env bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/.env
source $SCRIPT_DIR/_common.sh
source $SCRIPT_DIR/_common-aws.sh

awsInit

usage() {
  echo "Usage: ${BASH_SOURCE[0]} [default|peering] [create|delete|list]"
  echo "  list - list SA keys"
  echo "  create - create new key"
  echo "  delete - delete oldest key"
  exit $1
}

listKeys() {
  KEYS_FILE=$(mktemp)
  trap "rm -f \"$KEYS_FILE\"" EXIT
  aws iam list-access-keys --user-name "$SA" \
    | jq -r ".AccessKeyMetadata | sort_by(.CreateDate)" > $KEYS_FILE
}

create() {
  log "Creating new key for SA $SA"
  local FN=$(mktemp)
  trap "rm -f \"$FN\"" EXIT

  aws iam create-access-key --user-name "$SA" | tee "$FN"

  local KEY=$(jq -r '.AccessKey.AccessKeyId' "$FN")
  local SECRET=$(jq -r '.AccessKey.SecretAccessKey' "$FN")
  putCredentialKeyVal "accessKeyID" "$KEY"
  putCredentialKeyVal "secretAccessKey" "$SECRET"
  if [ "$SA_TYPE" = "default" ]; then
    saveCredentialsToGarden "$AWS_GARDEN_DEFAULT_SECRET"
  else
    saveCredentialsToGarden "$AWS_GARDEN_PEERING_SECRET"
  fi
}


delete() {
  log "Deleting oldest key of SA $SA"
  listKeys
  local ID=$(cat $KEYS_FILE | jq -r "sort_by(.CreateDate) | .[0] | .AccessKeyId")
  log "Oldest key id is $ID"
  if [ "$ID" == "null" ]; then
    log "The SA $SA has no keys that can be deleted"
    exit 1
  fi
  aws iam delete-access-key --user-name "$SA" --access-key-id "$ID"
  log "The key with id $ID is deleted"
}

list() {
  log "Listing keys of SA $SA"
  listKeys
  cat $KEYS_FILE
}

SA_TYPE=$1
CMD=$2

case "$1" in
default)
  SA="$SA_NAME_DEFAULT"
  ;;
peering)
  SA="$SA_NAME_PEERING"
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
