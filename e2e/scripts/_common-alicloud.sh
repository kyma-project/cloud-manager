alicloudInit() {
  checkRequiredCommands 'aliyun jq'

  initFileVar "POLICY_FILE_DEFAULT" "../../docs/contributor/permissions/alicloud/policy-CloudManagerAccess.json"

  SA_NAME_DEFAULT="${SA_NAME_DEFAULT:-cloud-manager}"
  POLICY_NAME_DEFAULT="${POLICY_NAME_DEFAULT:-CloudManagerAccess}"

  echo "SA_NAME_DEFAULT=$SA_NAME_DEFAULT"
  echo "POLICY_NAME_DEFAULT=$POLICY_NAME_DEFAULT"
  echo "POLICY_FILE_DEFAULT=$POLICY_FILE_DEFAULT"
  echo ""

  return 0
}
