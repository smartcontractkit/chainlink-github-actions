#!/usr/bin/env bash

set -euo pipefail

##
# Check if git tag is a release or pre-release.
#
# Examples:
#   1. v1.2.3-beta.0 -> pre-release
#   2. v1.2.3 -> release
##

if [[ -z "${GITHUB_REF:-}" ]]; then
    echo "ERROR: GITHUB_REF environment variable is required"
    exit 1
fi

TAG_REF="${GITHUB_REF}"
TAG_NAME=${TAG_REF:10} # remove "refs/tags/" prefix
echo "The tag name is $TAG_NAME".
echo "Checking if $TAG_NAME is a release or pre-release tag..."
IS_RELEASE=false
IS_PRE_RELEASE=false
if [[ $TAG_NAME =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    IS_RELEASE="true"
elif [[ $TAG_NAME =~ ^v[0-9]+\.[0-9]+\.[0-9]+-(.+)$ ]]; then
    IS_PRE_RELEASE="true"
fi
echo "is-release=${IS_RELEASE}" | tee -a "$GITHUB_OUTPUT"
echo "is-pre-release=${IS_PRE_RELEASE}" | tee -a "$GITHUB_OUTPUT"
