
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
