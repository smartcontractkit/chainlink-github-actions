# Release Tag Check

Checks if git tag is a release or pre-release, and tells you the version.

## Inputs

These are passed by setting environment variables.

- GITHUB_REF
    - Automatically available in a Github workflow. Will only work with `tag` pushes, otherwise the extracted ref will have an extra `/`
    - If a tag is the git ref, the prefix will be `refs/tag/`, if a branch is the git ref, the prefix will be `refs/head/` (9 characters vs 10 characters).
- RELEASE_REGEX
    - Used to determine if the tag pushed is the expected format of a release
    - Defaults to: `^v[0-9]+\.[0-9]+\.[0-9]+$`
- PRE_RELEASE_REGEX
    - Used to determine if the tag pushed is the expected format of a pre-release
    - Defaults to: `^v[0-9]+\.[0-9]+\.[0-9]+-(.+)$`
- VERSION_PREFIX
    - Used for determining the `release-version` and `pre-release-version` outputs only. This will not affect how the release/pre-release regexes determine the output.
    - Defaults to: `v`


## Outputs

- `is-release` - whether the tag name conformed to the release regex (`refs/tag/<tag name>`)
    - If yes, `release-version` should be set to the version. Without the `$VERSION_PREFIX` on the tag name
- `is-pre-release` whether the tag name conformed to the pre-release regex (`refs/tag/<tag name>`)
    - If yes, `pre-release-version` should be set to the version. Without the `$VERSION_PREFIX` on the tag name


## Examples

1. Ref: refs/tag/v1.2.3-beta.0
    - is-pre-release: true
    - is-release: false
    - pre-release-version: 1.2.3-beta.0
    - release-version: null
2. Ref: refs/tag/v1.2.3
    - is-pre-release: false
    - is-release: true
    - pre-release-version: null
    - release-version: 1.2.3
3. Ref: refs/tag/release-v1.2.3 (must override release_regex, and VERSION_PREFIX)
    - is-pre-release: false
    - is-release: true
    - pre-release-version: null
    - release-version: 1.2.3
4. Ref: refs/head/v1.2.3
    - is-pre-release: false
    - is-release: false
    - pre-release-version: null
    - release-version: false
