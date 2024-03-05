#!/usr/bin/env bash

run_test_case() {
    echo -e "===================================================================="
    GITHUB_REF="$1"
    expected_release="$2"
    expected_pre_release="$3"
    expected_release_version="$4"
    expected_pre_release_version="$5"
    release_regex="${6:-}" # Optional, use default if not provided
    pre_release_regex="${7:-}" # Optional, use default if not provided

    # Create a temporary file for GITHUB_OUTPUT
    GITHUB_OUTPUT=$(mktemp)

    # Set environment variables for the test case, including optional regex overrides
    env_vars=(
        "GITHUB_REF=$GITHUB_REF"
        "GITHUB_OUTPUT=$GITHUB_OUTPUT"
        "VERSION_PREFIX=${VERSION_PREFIX:-v}"
    )

    # Add regex overrides to environment variables if provided
    [[ -n "$release_regex" ]] && env_vars+=("RELEASE_REGEX=$release_regex")
    [[ -n "$pre_release_regex" ]] && env_vars+=("PRE_RELEASE_REGEX=$pre_release_regex")

    # Run the script with the environment variables set for this test case
    output=$(env "${env_vars[@]}" ./releasetagcheck.sh)

    # Read outputs from GITHUB_OUTPUT file and trim potential trailing newlines
    is_release=$(grep "^is-release" "$GITHUB_OUTPUT" | cut -d= -f2 | tr -d '\n')
    is_pre_release=$(grep "^is-pre-release" "$GITHUB_OUTPUT" | cut -d= -f2 | tr -d '\n')
    release_version=$(grep "^release-version" "$GITHUB_OUTPUT" | cut -d= -f2 | tr -d '\n')
    pre_release_version=$(grep "^pre-release-version" "$GITHUB_OUTPUT" | cut -d= -f2 | tr -d '\n')

    # Verify the outputs
    if [[ "$is_release" == "$expected_release" && \
        "$is_pre_release" == "$expected_pre_release" && \
        "$release_version" == "$expected_release_version" && \
        "$pre_release_version" == "$expected_pre_release_version" ]]; then
        echo "Test case $GITHUB_REF passed."
    else
        echo "Test case $GITHUB_REF failed."
        echo "Expected: is-release=$expected_release, is-pre-release=$expected_pre_release, release-version=$expected_release_version, pre-release-version=$expected_pre_release_version"
        echo "Got: is-release=$is_release, is-pre-release=$is_pre_release, release-version=$release_version, pre-release-version=$pre_release_version"
    fi

    # Clean up the temporary GITHUB_OUTPUT file
    rm -f "$GITHUB_OUTPUT"

    echo -e "====================================================================\n"
}

# Test cases
run_test_case "refs/tags/v1.2.3" "true" "false" "1.2.3" "null"
run_test_case "refs/tags/v1.2.3-beta.1" "false" "true" "null" "1.2.3-beta.1"
run_test_case "refs/tags/v1.2.3.4" "false" "false" "null" "null" # Invalid tag
run_test_case "refs/tags/release-1.2.3" "false" "false" "null" "null" # Custom tag not matching default regex

# Standard release version
run_test_case "refs/tags/v1.2.4" "true" "false" "1.2.4" "null"

# Pre-release with multiple identifiers
run_test_case "refs/tags/v1.2.5-alpha.1.beta" "false" "true" "null" "1.2.5-alpha.1.beta"

# Release version without 'v' prefix (requires changing VERSION_PREFIX)
VERSION_PREFIX=""
run_test_case "refs/tags/1.2.6" "true" "false" "1.2.6" "null" "^[0-9]+\.[0-9]+\.[0-9]+$"
VERSION_PREFIX="v" 

# Tag with a non-standard prefix "release-v"
VERSION_PREFIX="release-v"
run_test_case "refs/tags/release-v1.3.0" "true" "false" "1.3.0" "null" "^release-v[0-9]+\.[0-9]+\.[0-9]+$"
VERSION_PREFIX="v"

# Tag with a non-standard prefix "release-" (no 'v')
VERSION_PREFIX="release-"
run_test_case "refs/tags/release-1.3.0" "true" "false" "1.3.0" "null" "^release-[0-9]+\.[0-9]+\.[0-9]+$"
VERSION_PREFIX="v"

# Tag with a non-standard prefix "release-v" (prerelease)
VERSION_PREFIX="release-v"
run_test_case "refs/tags/release-v1.3.0-beta.5" "false" "true" "null" "1.3.0-beta.5" "^release-v[0-9]+\.[0-9]+\.[0-9]+$" "^release-v[0-9]+\.[0-9]+\.[0-9]+-(.+)$"
VERSION_PREFIX="v"

# Tag with a non-standard prefix "release-" (no 'v') (prereslease)
VERSION_PREFIX="release-"
run_test_case "refs/tags/release-1.3.0-beta.0" "false" "true" "null" "1.3.0-beta.0" "^release-[0-9]+\.[0-9]+\.[0-9]+$" "^release-[0-9]+\.[0-9]+\.[0-9]+-(.+)$"
VERSION_PREFIX="v"

# Tag with complex pre-release and build metadata
run_test_case "refs/tags/v1.3.1-rc.1+build.123" "false" "true" "null" "1.3.1-rc.1+build.123"

# Edge case: version with leading zeros
run_test_case "refs/tags/v0.0.9" "true" "false" "0.0.9" "null"

# Edge case: version with extended numeric identifiers
run_test_case "refs/tags/v1.2.3.4" "false" "false" "null" "null"

# Invalid tag not following semantic versioning
run_test_case "refs/tags/v1.2" "false" "false" "null" "null"
