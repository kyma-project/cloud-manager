
initRoleFileDefault() {
  local DEFAULT_VALUE=$1
  ROLE_FILE_DEFAULT="${ROLE_FILE_DEFAULT:-$DEFAULT_VALUE}"
  case $ROLE_FILE_DEFAULT in (/*) : ;; (*) ROLE_FILE_DEFAULT="${SCRIPT_DIR}/${ROLE_FILE_DEFAULT}" ;; esac
  if [ ! -f "$ROLE_FILE_DEFAULT" ]; then
    echo "ROLE_FILE_DEFAULT $ROLE_FILE_DEFAULT not found"
    exit 1
  fi
}

initRoleFilePeering() {
  local DEFAULT_VALUE=$1
  ROLE_FILE_PEERING="${ROLE_FILE_PEERING:-$DEFAULT_VALUE}"
  case $ROLE_FILE_PEERING in (/*) : ;; (*) ROLE_FILE_PEERING="${SCRIPT_DIR}/${ROLE_FILE_PEERING}" ;; esac
  if [ ! -f "$ROLE_FILE_PEERING" ]; then
    echo "ROLE_FILE_PEERING $ROLE_FILE_PEERING not found"
    exit 1
  fi
}

initSANameDefault() {
  local DEFAULT_VALUE=$1
  SA_NAME_DEFAULT="${SA_NAME_DEFAULT:-$DEFAULT_VALUE}"
}

initSANamePeering() {
  local DEFAULT_VALUE=$1
  SA_NAME_PEERING="${SA_NAME_PEERING:-$DEFAULT_VALUE}"
}

initRoleNameDefault() {
  local DEFAULT_VALUE=$1
  ROLE_NAME_DEFAULT="${ROLE_NAME_DEFAULT:-$DEFAULT_VALUE}"
}

initRoleNamePeering() {
  local DEFAULT_VALUE=$1
  ROLE_NAME_PEERING="${ROLE_NAME_PEERING:-$DEFAULT_VALUE}"
}

# GCP Specific

gcpValidateProject() {
  if [ -z "${GCP_PROJECT+x}" ]; then
    echo "GCP_PROJECT is not set"
    exit 1
  fi
}

gcpInit() {
  checkRequiredCommands 'gcloud jq'

  gcpValidateProject

  initRoleFileDefault '../../docs/contributor/permissions/gcp/gcp_default.yaml'
  initRoleFilePeering '../../docs/contributor/permissions/gcp/gcp_peering.yaml'

  initSANameDefault 'cloud-manager-e2e'
  initSANamePeering 'cloud-manager-peering-e2e'

  initRoleNameDefault 'cloud_manager_e2e'
  initRoleNamePeering 'cloud_manager_peering_e2e'

  echo "GCP_PROJECT=$GCP_PROJECT"
  echo "SA_NAME_DEFAULT=$SA_NAME_DEFAULT"
  echo "ROLE_NAME_DEFAULT=$ROLE_NAME_DEFAULT"
  echo "ROLE_FILE_DEFAULT=$ROLE_FILE_DEFAULT"
  echo "SA_NAME_PEERING=$SA_NAME_PEERING"
  echo "ROLE_NAME_PEERING=$ROLE_NAME_PEERING"
  echo "ROLE_FILE_PEERING=$ROLE_FILE_PEERING"
  echo ""
}

