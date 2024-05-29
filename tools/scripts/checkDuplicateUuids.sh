#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
PROJECT_ROOT_DIR=`realpath "$SCRIPT_DIR/../.."`

DUPS=`find $PROJECT_ROOT_DIR/internal -name *_test.go \
  | xargs cat \
  | grep -o -E "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}" \
  | sort \
  | uniq -d \
  | grep -v '^ *1 '`

CNT=`echo "$DUPS" | wc -l | tr -d " "`

if [ "$CNT" != "0" ]; then
  echo "Found duplicate UUIDs in go test files"
  echo "$DUPS"
  exit 1
fi


