#!/usr/bin/env bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source $SCRIPT_DIR/.env
source $SCRIPT_DIR/_common.sh

gcpInit

# Functions

createServiceAccount() {
  SA_NAME=$1
  SA="${SA_NAME}@${GCP_PROJECT}.iam.gserviceaccount.com"
  if ! gcloud iam service-accounts describe $SA --project $GCP_PROJECT > /dev/null 2>&1 ; then
    log "SA $SA does not exist, creating it now..."
    gcloud iam service-accounts create $SA_NAME --project $GCP_PROJECT > /dev/null
    log "SA $SA is created"
  else
    log "SA $SA already exists"
  fi
}

setServiceAccountPermissions() {
  ROLE_NAME=$1
  ROLE_FILE=$2
  if ! gcloud iam roles describe $ROLE_NAME --project $GCP_PROJECT > /dev/null 2>&1 ; then
    log "Role $ROLE_NAME does not exist, creating it now from $ROLE_FILE ..."
    gcloud iam roles create $ROLE_NAME --project $GCP_PROJECT --file $ROLE_FILE > /dev/null
  else
    IS_DELETED=$(gcloud iam roles describe $ROLE_NAME --project $GCP_PROJECT --format json | jq '.deleted')
    if [ $IS_DELETED = "true" ]; then
      log "Role $ROLE_NAME is deleted, undeleting..."
      gcloud iam roles undelete $ROLE_NAME --project $GCP_PROJECT > /dev/null
    fi
    log "Role $ROLE_NAME exist, updating it now from $ROLE_FILE..."
    gcloud iam roles update $ROLE_NAME --project $GCP_PROJECT --file $ROLE_FILE > /dev/null
  fi
}


setProjectIAM() {
  SA_NAME=$1
  SA="${SA_NAME}@${GCP_PROJECT}.iam.gserviceaccount.com"
  ROLE_NAME=$2

  log "Updating project IAM for SA $SA with role $ROLE_NAME"

  gcloud projects add-iam-policy-binding $GCP_PROJECT --project="$GCP_PROJECT" \
    --member="serviceAccount:$SA" --role="projects/$GCP_PROJECT/roles/$ROLE_NAME" > /dev/null
}

# Main

createServiceAccount "$SA_NAME_DEFAULT"
createServiceAccount "$SA_NAME_PEERING"

setServiceAccountPermissions "$ROLE_NAME_DEFAULT" "$ROLE_FILE_DEFAULT"
setServiceAccountPermissions "$ROLE_NAME_PEERING" "$ROLE_FILE_PEERING"

setProjectIAM "$SA_NAME_DEFAULT" "$ROLE_NAME_DEFAULT"
setProjectIAM "$SA_NAME_PEERING" $ROLE_NAME_PEERING
