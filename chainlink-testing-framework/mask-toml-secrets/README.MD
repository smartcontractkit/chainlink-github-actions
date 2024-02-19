# Mask TOML Secrets Action

This GitHub Action masks all secret values in Base64-encoded TOML configurations within GitHub Actions logs. It is designed to ensure that sensitive information denoted by keys ending with `_secret` is not exposed in Github logs.

## Inputs

### `base64_toml_content`

- **Description**: Base64-encoded TOML configurations separated by spaces that contain "_secret" keys to be masked.
- **Required**: Yes

## How to Use

To use this action in your workflow, follow the steps below:

1. Ensure you have a Base64-encoded TOML configuration string that you wish to mask. The keys for any secrets in the TOML should end with `_secret`.
2. Add a step in your GitHub Actions workflow to use this action. Do it **before** any other action related to the TOML 

Example: 

```yaml
steps:
  - name: Mask TOML Secrets
    uses: smartcontractkit/chainlink-github-actions/chainlink-testing-framework/mask-toml-secrets@main
    with:
      base64_toml_content: >
        ${{ secrets.DOCKER_TESTS_BASE64_TOML_CONTENT_1 }}
        ${{ secrets.DOCKER_TESTS_BASE64_TOML_CONTENT_2 }}
```