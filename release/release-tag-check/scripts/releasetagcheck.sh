#!/usr/bin/env bash

set -euo pipefail

# Configurable regex patterns with defaults
RELEASE_REGEX=${RELEASE_REGEX:-"^v[0-9]+\.[0-9]+\.[0-9]+$"}
PRE_RELEASE_REGEX=${PRE_RELEASE_REGEX:-"^v[0-9]+\.[0-9]+\.[0-9]+-(.+)$"}

# Configurable prefix removal with default
VERSION_PREFIX=${VERSION_PREFIX:-"v"} 

if [[ -z "${GITHUB_REF:-}" ]]; then
    echo "ERROR: GITHUB_REF environment variable is required"
    exit 1
fi

TAG_REF="${GITHUB_REF}"
TAG_NAME=${TAG_REF:10} # remove "refs/tags/" prefix

# Remove specified prefix from the version tag
VERSION_TAG=${TAG_NAME#"${VERSION_PREFIX}"}

echo "Tag: $TAG_NAME"
echo "Checking if $TAG_NAME is a release or pre-release tag..."

IS_RELEASE=false
IS_PRE_RELEASE=false
RELEASE_VERSION="null"
PRE_RELEASE_VERSION="null"

if [[ $TAG_NAME =~ $RELEASE_REGEX ]]; then
    echo "Release tag detected. Tag: $TAG_NAME - Version: $VERSION_TAG"
    IS_RELEASE=true
    RELEASE_VERSION=$VERSION_TAG
elif [[ $TAG_NAME =~ $PRE_RELEASE_REGEX ]]; then
    echo "Pre-release tag detected. Tag: $TAG_NAME - Version: $VERSION_TAG"
    IS_PRE_RELEASE=true
    PRE_RELEASE_VERSION=$VERSION_TAG
else
    echo "No release or pre-release tag detected. Tag: $TAG_NAME"
fi

echo "is-release=$IS_RELEASE" | tee -a "$GITHUB_OUTPUT"
echo "release-version=$RELEASE_VERSION" | tee -a "$GITHUB_OUTPUT"

echo "is-pre-release=$IS_PRE_RELEASE" | tee -a "$GITHUB_OUTPUT"
echo "pre-release-version=$PRE_RELEASE_VERSION" | tee -a "$GITHUB_OUTPUT"
