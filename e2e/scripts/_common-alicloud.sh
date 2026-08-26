alicloudInit() {
  checkRequiredCommands 'aliyun jq'

  initFileVar "POLICY_FILE_DEFAULT" "../../docs/contributor/permissions/alicloud/policy-CloudManagerAccess.json"

  SA_NAME_DEFAULT="${SA_NAME_DEFAULT:-cloud-manager-e2e}"
  POLICY_NAME_DEFAULT="${POLICY_NAME_DEFAULT:-CloudManagerAccess}"

  if [[ -z "$QUIET" ]]; then
    echo "SA_NAME_DEFAULT=$SA_NAME_DEFAULT"
    echo "POLICY_NAME_DEFAULT=$POLICY_NAME_DEFAULT"
    echo ""
  fi

  return 0
}
