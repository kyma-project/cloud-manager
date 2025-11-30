
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



awsInit() {
  checkRequiredCommands 'aws jq tee'

  validateAwsAccount

  initFileVar "POLICY_FILE_DEFAULT" "../../docs/contributor/permissions/aws/policy-CloudManagerAccess.json"
  initFileVar "POLICY_FILE_PEERING" "../../docs/contributor/permissions/aws/policy-CloudManagerPeeringAccess.json"
  initFileVar "POLICY_FILE_CM_ASSUME_ROLE" "../../docs/contributor/permissions/aws/policy-CloudManagerAssumeRole.json"
  initFileVar "TRUST_FILE_ALLOW_CLOUD_MANAGER_ASSUME_ROLE" "../../docs/contributor/permissions/aws/trust-AllowCloudManagerAssumeRole.json"
  initFileVar "TRUST_FILE_BACKUP_SERVICE" "../../docs/contributor/permissions/aws/trust-CloudManagerBackupServiceRole.json"

  SA_NAME_DEFAULT="${SA_NAME_DEFAULT:-cloud-manager-e2e}"
  SA_NAME_PEERING="${SA_NAME_PEERING:-cloud-manager-peering-e2e}"
  ROLE_NAME_DEFAULT="${ROLE_NAME_DEFAULT:-CloudManagerRole}"
  ROLE_NAME_PEERING="${ROLE_NAME_PEERING:-CloudManagerPeeringRole}"
  ROLE_NAME_BACKUP_SERVICE="${ROLE_NAME_BACKUP_SERVICE:-CloudManagerBackupServiceRole}"

  POLICY_NAME_DEFAULT="${POLICY_NAME_DEFAULT:-CloudManagerAccess}"
  POLICY_NAME_PEERING="${POLICY_NAME_PEERING:-CloudManagerPeeringAccess}"

  POLICY_NAME_CM_ASSUME_ROLE="${POLICY_NAME_CM_ASSUME_ROLE:-CloudManagerAssumeRole}"

  if [ -z "$QUIET" ]; then
    echo "AWS_ACCOUNT=$AWS_ACCOUNT"
    echo "SA_NAME_DEFAULT=$SA_NAME_DEFAULT"
    echo "SA_NAME_PEERING=$SA_NAME_PEERING"
    echo "ROLE_NAME_DEFAULT=$ROLE_NAME_DEFAULT"
    echo "ROLE_NAME_PEERING=$ROLE_NAME_PEERING"
    echo "ROLE_NAME_BACKUP_SERVICE=$ROLE_NAME_BACKUP_SERVICE"
    echo "POLICY_NAME_CM_ASSUME_ROLE=$POLICY_NAME_CM_ASSUME_ROLE"
    echo "POLICY_FILE_CM_ASSUME_ROLE=$POLICY_FILE_CM_ASSUME_ROLE"
    echo "TRUST_FILE_ALLOW_CLOUD_MANAGER_ASSUME_ROLE=$TRUST_FILE_ALLOW_CLOUD_MANAGER_ASSUME_ROLE"
    echo "TRUST_FILE_BACKUP_SERVICE=$TRUST_FILE_BACKUP_SERVICE"
    echo ""
  fi

  log "Running on AWS account $AWS_ACCOUNT"

}
