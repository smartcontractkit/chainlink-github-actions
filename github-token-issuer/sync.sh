#!/bin/bash
set -e

# Dependencies:
# gh cli ^2.15.0 https://github.com/cli/cli/releases/tag/v2.15.0
# jq ^1.6 https://stedolan.github.io/jq/

repo=smartcontractkit/github-app-token-issuer
gitRoot=$(git rev-parse --show-toplevel)

msg() {
  echo -e "\033[32m$1\033[39m" >&2
}

err() {
  echo -e "\033[31m$1\033[39m" >&2
}

fail() {
  err "$1"
  exit 1
}
cd "$gitRoot/github-token-issuer"

echo "Getting latest release for tag for $repo"
action_releases=$(gh release list -R $repo | grep action | head -1 | awk '{ print $1 }')
release=$(gh release view -R $repo --json 'tagName,body' "$action_releases")
tag=$(echo "$release" | jq -r '.tagName')

echo "Getting release $tag for $repo"
release=$(gh release view "$tag" -R $repo --json 'assets')
asset_name=$(echo "$release" | jq -r '.assets | map(select(.contentType == "application/x-gtar"))[0].name')

echo "Downloading ${repo}:${tag} asset: $asset_name..."
echo ""
gh release download "$tag" -R "$repo" -p "$asset_name"

echo "Unpacking asset $asset_name"
tar -xvzf "$asset_name"

msg ""
cp -rf package/. "." || true

msg "Cleaning up"
rm -r package
rm "$asset_name"
