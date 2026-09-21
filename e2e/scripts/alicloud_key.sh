#!/usr/bin/env bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/.env
source $SCRIPT_DIR/_common.sh
source $SCRIPT_DIR/_common-alicloud.sh

alicloudInit

usage() {
  echo "Usage: ${BASH_SOURCE[0]} [create|delete|list]"
  echo "  list   - list access keys for the RAM user"
  echo "  create - create a new access key"
  echo "  delete - delete the oldest access key"
  local exit_code
  exit_code=${1:-0}
  exit $exit_code
}

listKeys() {
  KEYS_FILE=$(mktemp)
  trap "rm -f \"$KEYS_FILE\"" EXIT
  aliyun ram ListAccessKeys --UserName "$SA_NAME_DEFAULT" \
    | jq -r '.AccessKeys.AccessKey | sort_by(.CreateDate)' > "$KEYS_FILE"

  return 0
}

create() {
  log "Creating new access key for RAM user $SA_NAME_DEFAULT"
  local fn
  fn=$(mktemp)
  trap "rm -f \"$fn\"" EXIT

  aliyun ram CreateAccessKey --UserName "$SA_NAME_DEFAULT" | tee "$fn"

  local key
  key=$(jq -r '.AccessKey.AccessKeyId' "$fn" | tr -d '\n')
  local secret
  secret=$(jq -r '.AccessKey.AccessKeySecret' "$fn" | tr -d '\n')

  putCredentialKeyVal "accessKeyID" "$key"
  putCredentialKeyVal "accessKeySecret" "$secret"
  saveCredentialsToGarden "$ALICLOUD_GARDEN_DEFAULT_SECRET"

  return 0
}

delete() {
  log "Deleting oldest access key of RAM user $SA_NAME_DEFAULT"
  listKeys
  local id
  id=$(jq -r '.[0].AccessKeyId' "$KEYS_FILE")
  log "Oldest access key id is $id"
  if [[ "$id" == "null" ]]; then
    log "The RAM user $SA_NAME_DEFAULT has no access keys that can be deleted"
    exit 1
  fi
  aliyun ram DeleteAccessKey --UserName "$SA_NAME_DEFAULT" --UserAccessKeyId "$id"
  log "The access key with id $id is deleted"

  return 0
}

list() {
  log "Listing access keys of RAM user $SA_NAME_DEFAULT"
  listKeys
  cat "$KEYS_FILE"

  return 0
}

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
