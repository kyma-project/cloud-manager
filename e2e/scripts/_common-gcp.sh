
initRoleFileDefault() {
  local default_value=$1
  ROLE_FILE_DEFAULT="${ROLE_FILE_DEFAULT:-$default_value}"
  case $ROLE_FILE_DEFAULT in (/*) : ;; (*) ROLE_FILE_DEFAULT="${SCRIPT_DIR}/${ROLE_FILE_DEFAULT}" ;; esac
  if [[ ! -f "$ROLE_FILE_DEFAULT" ]]; then
    echo "ROLE_FILE_DEFAULT $ROLE_FILE_DEFAULT not found"
    exit 1
  fi
  return 0
}

initRoleFilePeering() {
  local default_value=$1
  ROLE_FILE_PEERING="${ROLE_FILE_PEERING:-$default_value}"
  case $ROLE_FILE_PEERING in (/*) : ;; (*) ROLE_FILE_PEERING="${SCRIPT_DIR}/${ROLE_FILE_PEERING}" ;; esac
  if [[ ! -f "$ROLE_FILE_PEERING" ]]; then
    echo "ROLE_FILE_PEERING $ROLE_FILE_PEERING not found"
    exit 1
  fi
  return 0
}

initSANameDefault() {
  local default_value=$1
  SA_NAME_DEFAULT="${SA_NAME_DEFAULT:-$default_value}"
  return 0
}

initSANamePeering() {
  local default_value=$1
  SA_NAME_PEERING="${SA_NAME_PEERING:-$default_value}"
  return 0
}

initRoleNameDefault() {
  local default_value=$1
  ROLE_NAME_DEFAULT="${ROLE_NAME_DEFAULT:-$default_value}"
  return 0
}

initRoleNamePeering() {
  local default_value=$1
  ROLE_NAME_PEERING="${ROLE_NAME_PEERING:-$default_value}"
  return 0
}

# GCP Specific

gcpValidateProject() {
  if [[ -z "${GCP_PROJECT+x}" ]]; then
    echo "GCP_PROJECT is not set"
    exit 1
  fi
  return 0
}

gcpInit() {
  checkRequiredCommands 'gcloud jq tee'

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

  return 0
}
