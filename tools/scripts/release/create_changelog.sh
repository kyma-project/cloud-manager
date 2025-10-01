#!/usr/bin/env bash

# This script will create a changelog for a release based on the latest commits.


# standard bash error handling
set -o nounset  # treat unset variables as an error and exit immediately.
set -o errexit  # exit immediately when a command fails.
set -E          # needs to be set if we want the ERR trap
set -o pipefail # prevents errors in a pipeline from being masked

REPOSITORY=$1
RELEASE_TAG=$2

GITHUB_URL=https://api.github.com/repos/${REPOSITORY}
GITHUB_AUTH_HEADER="Authorization: token ${GITHUB_TOKEN}"
CHANGELOG_FILE="CHANGELOG.md"

PREVIOUS_RELEASE=$(git describe --tags --abbrev=0)

echo -e "\n**Full changelog**: https://github.com/$REPOSITORY/compare/${PREVIOUS_RELEASE}...${RELEASE_TAG}" >>${CHANGELOG_FILE}

# cleanup
echo "CHANGELOG.md created"
