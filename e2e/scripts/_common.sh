
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
