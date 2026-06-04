#!/usr/bin/env bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/.env
source $SCRIPT_DIR/_common.sh
source $SCRIPT_DIR/_common-gcp.sh

gcpInit

# Functions

usage() {
  echo "Usage: ${BASH_SOURCE[0]} [default|peering] [create|delete|list]"
  echo "  list - list SA keys"
  echo "  create - create new key"
  echo "  delete - delete oldest key"
  local exit_code
  exit_code=${1:-0}
  exit $exit_code
}

listKeys() {
  KEYS_FILE=$(mktemp)
  trap "rm -f \"$KEYS_FILE\"" EXIT
  gcloud iam service-accounts keys list --iam-account="$SA" --project "$GCP_PROJECT" --format=json --filter='keyType: USER_MANAGED' \
    | jq -r "sort_by(.validAfterTime)" > $KEYS_FILE

  return 0
}

create() {
  log "Creating new key for SA $SA"
  local fn
  fn=$(mktemp)
  trap "rm -f \"$fn\"" EXIT
  gcloud iam service-accounts keys create "$fn" --iam-account="$SA" --project $GCP_PROJECT

  echo ""
  cat $fn

  local val
  val=$(cat "$fn")
  putCredentialKeyVal "serviceaccount.json" "$val"
  if [[ "$SA_TYPE" = "default" ]]; then
    saveCredentialsToGarden "$GCP_GARDEN_DEFAULT_SECRET"
  else
    saveCredentialsToGarden "$GCP_GARDEN_PEERING_SECRET"
  fi

  rm -r "$fn"

  return 0
}


delete() {
  log "Deleting oldest key of SA $SA"
  listKeys
  local id
  id=$(cat $KEYS_FILE | jq -r "sort_by(.validAfterTime) | .[0] | .name")
  id=$(basename "$id")
  log "Oldest key id is $id"
  gcloud iam service-accounts keys delete $id --iam-account="$SA" --project "$GCP_PROJECT"
  log "The key with id $id is deleted"

  return 0
}

list() {
  log "Listing keys of SA $SA"
  listKeys
  cat $KEYS_FILE
  return 0
}

SA_TYPE=$1
CMD=$2

case "$1" in
default)
  SA_NAME="$SA_NAME_DEFAULT"
  SA="${SA_NAME_DEFAULT}@${GCP_PROJECT}.iam.gserviceaccount.com"
  ;;
peering)
  SA_NAME="$SA_NAME_PEERING"
  SA="${SA_NAME_PEERING}@${GCP_PROJECT}.iam.gserviceaccount.com"
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

