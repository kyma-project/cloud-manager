
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

initRoleNameDefault() {
  local DEFAULT_VALUE=$1
  ROLE_NAME_DEFAULT="${ROLE_NAME_DEFAULT:-$DEFAULT_VALUE}"
}

initRoleNamePeering() {
  local DEFAULT_VALUE=$1
  ROLE_NAME_PEERING="${ROLE_NAME_PEERING:-$DEFAULT_VALUE}"
}

# Azure specific
azureValidateSubscription(){
    if [ -z "${AZURE_SUBSCRIPTION_ID+x}" ]; then
      echo "AZURE_SUBSCRIPTION_ID is not set"
      exit 1
    fi
}

azureValidateDefaultAppId(){
    if [ -z "${AZURE_DEFAULT_APP_ID+x}" ]; then
      echo "AZURE_DEFAULT_APP_ID is not set"
      exit 1
    fi
}

azureValidatePeeringAppId(){
    if [ -z "${AZURE_PEERING_APP_ID+x}" ]; then
      echo "AZURE_PEERING_APP_ID is not set"
      exit 1
    fi
}



azureInit() {
  checkRequiredCommands 'az jq'

  azureValidateSubscription
  azureValidateDefaultAppId
  azureValidatePeeringAppId

  initRoleFileDefault '../../docs/contributor/permissions/azure_default.json'
  initRoleFilePeering '../../docs/contributor/permissions/azure_peering.json'

  initRoleNameDefault 'cloud_manager_e2e'
  initRoleNamePeering 'cloud_manager_peering_e2e'


  echo "ROLE_NAME_DEFAULT=$ROLE_NAME_DEFAULT"
  echo "ROLE_FILE_DEFAULT=$ROLE_FILE_DEFAULT"
  echo "ROLE_NAME_PEERING=$ROLE_NAME_PEERING"
  echo "ROLE_FILE_PEERING=$ROLE_FILE_PEERING"
  echo ""

  echo "=== Azure ==="
  echo "AZURE_SUBSCRIPTION_ID=$AZURE_SUBSCRIPTION_ID"
  echo "AZURE_DEFAULT_APP_ID=$AZURE_DEFAULT_APP_ID"
  echo "AZURE_PEERING_APP_ID=$AZURE_PEERING_APP_ID"
  echo ""

}

