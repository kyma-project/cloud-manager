
validateAwsAccount() {
  if [ -z "${AWS_ACCOUNT+x}" ]; then
    echo "AWS_ACCOUNT is not set"
    exit 1
  fi
  local ACTUAL_ACCOUNT=$(aws sts get-caller-identity --query Account --output text)
  if [ "$AWS_ACCOUNT" != "$ACTUAL_ACCOUNT" ]; then
    echo "AWS_ACCOUNT mismatch: env defines to work on $AWS_ACCOUNT, but cli is logged into $ACTUAL_ACCOUNT"
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

initSAPolicyName() {
  :
}

awsInit() {
  checkRequiredCommands 'aws'

  validateAwsAccount

  initSANameDefault 'cloud-manager-e2e'
  initSANamePeering 'cloud-manager-peering-e2e'

  initRoleNameDefault 'CloudManagerRole'
  initRoleNamePeering 'CloudManagerRole'

  echo "AWS_ACCOUNT=$AWS_ACCOUNT"
  echo "SA_NAME_DEFAULT=$SA_NAME_DEFAULT"
  echo "ROLE_NAME_DEFAULT=$ROLE_NAME_DEFAULT"
#  echo "ROLE_FILE_DEFAULT=$ROLE_FILE_DEFAULT"
  echo "SA_NAME_PEERING=$SA_NAME_PEERING"
  echo "ROLE_NAME_PEERING=$ROLE_NAME_PEERING"
#  echo "ROLE_FILE_PEERING=$ROLE_FILE_PEERING"
  echo ""

  log "Running on AWS account $AWS_ACCOUNT"

}
