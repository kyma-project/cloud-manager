
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

initRoleConditionDefault() {
  local default_value=$1
  ROLE_CONDITION_DEFAULT="${ROLE_CONDITION_DEFAULT:-$default_value}"
  case $ROLE_CONDITION_DEFAULT in (/*) : ;; (*) ROLE_CONDITION_DEFAULT="${SCRIPT_DIR}/${ROLE_CONDITION_DEFAULT}" ;; esac
  if [[ ! -f "$ROLE_CONDITION_DEFAULT" ]]; then
    echo "ROLE_CONDITION_DEFAULT $ROLE_CONDITION_DEFAULT not found"
    exit 1
  fi
  return 0
}

# Azure specific
azureValidateSubscription(){
    if [[ -z "${AZURE_SUBSCRIPTION_ID+x}" ]]; then
      echo "AZURE_SUBSCRIPTION_ID is not set"
      exit 1
    fi
  return 0
}

azureValidateDefaultAppId(){
    if [[ -z "${AZURE_DEFAULT_APP_ID+x}" ]]; then
      echo "AZURE_DEFAULT_APP_ID is not set"
      exit 1
    fi
  return 0
}


azureInit() {
  checkRequiredCommands 'az jq tee'

  azureValidateSubscription
  azureValidateDefaultAppId

  initRoleFileDefault '../../docs/contributor/permissions/azure_default.json'
  initRoleFilePeering '../../docs/contributor/permissions/azure_peering.json'
  initRoleConditionDefault '../../docs/contributor/permissions/azure_default_condition.txt'

  initRoleNameDefault 'cloud_manager_e2e'
  initRoleNamePeering 'cloud_manager_peering_e2e'


  echo "ROLE_NAME_DEFAULT=$ROLE_NAME_DEFAULT"
  echo "ROLE_FILE_DEFAULT=$ROLE_FILE_DEFAULT"
  echo "ROLE_CONDITION_DEFAULT=$ROLE_CONDITION_DEFAULT"
  echo "ROLE_NAME_PEERING=$ROLE_NAME_PEERING"
  echo "ROLE_FILE_PEERING=$ROLE_FILE_PEERING"
  echo ""

  echo "=== Azure ==="
  echo "AZURE_SUBSCRIPTION_ID=$AZURE_SUBSCRIPTION_ID"
  echo "AZURE_DEFAULT_APP_ID=$AZURE_DEFAULT_APP_ID"
  echo ""

  return 0
}
