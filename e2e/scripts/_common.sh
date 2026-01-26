
log() {
  DT=$(date "+%Y-%m-%dT%H:%M:%S")
  echo "$DT $1"
}

checkRequiredCommands() {
  local REQUIRED_COMMANDS=$1
  for cmd in $REQUIRED_COMMANDS; do
    if ! command -v $cmd > /dev/null ; then
      echo "Command $cmd not found"
      exit 1
    fi
  done
}

initFileVar() {
  local VAR_NAME=$1
  local DEFAULT_VALUE=$2
  local VAR_VALUE=${!VAR_NAME:-$DEFAULT_VALUE}
  case $VAR_VALUE in (/*) : ;; (*) VAR_VALUE=$(realpath "${SCRIPT_DIR}/${VAR_VALUE}") ;; esac
  if [ ! -f "$VAR_VALUE" ]; then
    echo "File $VAR_NAME with path $VAR_VALUE not found"
    echo "VAR_NAME=$VAR_NAME"
    echo "DEFAULT_VALUTE=$DEFAULT_VALUE"
    echo "VAR_VALUE=$VAR_VALUE"
    exit 1
  fi
  eval "$VAR_NAME=\"$VAR_VALUE\""
}

TMP_CREDS=$(mktemp)
trap "rm -f \"$TMP_CREDS\"" EXIT

putCredentialKeyVal() {
  local key=$1
  local value=$2
  local FN
  FN=$(mktemp)
  trap "rm -f \"$FN\"" EXIT
  echo -n "$value" > "$FN"
  value=$(base64 < "$FN")
  echo "  ${key}: ${value}" >> "$TMP_CREDS"
  rm -f "$FN"
}

saveCredentialsToGarden() {
  checkRequiredCommands "kubectl"

  if [ -z "$GARDEN_KUBECONFIG" ]; then
    return 0
  fi
  if [ ! -f "$GARDEN_KUBECONFIG" ]; then
    echo "GARDEN_KUBECONFIG '$$GARDEN_KUBECONFIG' is not a valid file"
    exit 1
  fi

  local TXT
  TXT=$(cat "$TMP_CREDS")

  if [ "${#TXT}" -eq 0 ]; then
    return 0
  fi
  local SECRET_NAME="$1"
  if [ -z "$SECRET_NAME" ]; then
    echo "Error: saveCredentialsToGarden called w/out secret name parameter"
    exit 1
  fi

  local FN
  FN=$(mktemp)
  trap "rm -f \"$FN\"" EXIT

  cat << EOF > $FN
apiVersion: v1
kind: Secret
metadata:
  name: ${SECRET_NAME}
type: Opaque
data:
${TXT}
EOF

  echo ""
  cat "$FN"

  KUBECONFIG="$GARDEN_KUBECONFIG" kubectl apply -f "$FN"
  rm -f "$FN"
}
