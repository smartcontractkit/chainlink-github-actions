# Action

- [Action](#action)
- [Usage](#usage)
- [Running locally](#running-locally)
  - [Assuming a test role](#assuming-a-test-role)
  - [Running the action](#running-the-action)

This action lets fetch a github installation access token based on your current IAM role. Your IAM role determines the following:

- What repositories you have access to
- What permissions you have across the repositories

# Usage

```yml
jobs:
  create-pr:
    runs-on: ubuntu-latest
    steps:
      - name: Assume role
        uses: aws-actions/configure-aws-credentials@v1
        with:
          aws-access-key-id: ${{ env.AWS_ACCESS_KEY_ID }}
          aws-region: ${{ inputs.aws-region }}
          aws-secret-access-key: ${{ env.AWS_SECRET_ACCESS_KEY }}
          aws-session-token: ${{ env.AWS_SESSION_TOKEN }}
          role-duration-seconds: ${{ inputs.role-duration-seconds }}
          role-to-assume: ${{ inputs.role-to-assume }}
          role-skip-session-tagging: true

      - name: Checkout repository
        uses: actions/checkout@v3

      - name: Setup node
        uses: actions/setup-node@v3
        with:
          node-version-file: .tool-versions

      - name: Get github installation access token
        id: get-gh-token
        uses: ./.github/actions/github-token-issuer
        with:
          url: ${{ inputs.url }}

      - name: Create random files
        shell: bash
        run: |
          for n in {1..5}; do
              dd if=/dev/urandom of=file$( printf %03d "$n" ).bin bs=1 count=$(( RANDOM + 1024 ))
          done

      - name: Create Pull Request
        uses: peter-evans/create-pull-request@v4
        with:
          token: ${{ steps.get-gh-token.outputs.access-token }}
          branch-suffix: timestamp
          title: ${{ inputs.pr-title }}
```

# Running locally

For testing purposes, you can run the action locally. You'll need to have be authenticated with AWS as one of the roles configured within the GATI service to be issued a token.

## Assuming a test role

```bash
# Within e2e directory
# Get the root test user to use for the GATI service
pulumi stack -C deploy/ output -s e2e --json --show-secrets | jq '.lambdas[0].test'

# Prints out the test user information and their mappings
{
  "roleIdMapping": {
    "AROAVVLBPM2KALALYL4EI": {
      "permissions": {
        "actions": "write"
      },
      "repositories": [
        "gati-test-0-e8897fb",
        "gati-test-1-90d91e3"
      ]
    },
    "AROAVVLBPM2KEHAANLSSZ": {
      "permissions": {
        "secrets": "write"
      },
      "repositories": [
        "gati-test-0-e8897fb",
        "gati-test-1-90d91e3",
        "gati-test-2-5dd95d3"
      ]
    },
    "AROAVVLBPM2KGRLWZF6DW": {
      "permissions": {
        "contents": "write",
        "pull_requests": "write"
      },
      "repositories": [
        "gati-test-0-e8897fb",
        "gati-test-1-90d91e3",
        "gati-test-2-5dd95d3"
      ]
    },
    "AROAVVLBPM2KO2E5AR74O": {
      "permissions": {
        "actions": "write",
        "checks": null,
        "contents": "write",
        "pull_requests": "write"
      },
      "repositories": [
        "gati-test-0-e8897fb",
        "gati-test-1-90d91e3",
        "gati-test-2-5dd95d3"
      ]
    },
    ....
  },
  "roles": [
    {
      "arn": "arn:aws:iam::900438541234:role/app/gati-test/test/gati-test-test-0-role-f0d8ca1",
      "id": "AROAVVLBPM2KAUVKARZOB"
    },
    {
      "arn": "arn:aws:iam::900438541234:role/app/gati-test/test/gati-test-test-1-role-f5fa836",
      "id": "AROAVVLBPM2KGMTUEVN6W"
    },
    {
      "arn": "arn:aws:iam::900438541234:role/app/gati-test/test/gati-test-test-2-role-29c2483",
      "id": "AROAVVLBPM2KGRLWZF6DW"
    },
    {
      "arn": "arn:aws:iam::900438541234:role/app/gati-test/test/gati-test-test-3-role-c941f3d",
      "id": "AROAVVLBPM2KO2E5AR74O"
    },
    {
      "arn": "arn:aws:iam::900438541234:role/app/gati-test/test/gati-test-test-4-role-7e2fa68",
      "id": "AROAVVLBPM2KDMQWXBJZC"
    },
    ....
  ],
  "rootRoleArn": "arn:aws:iam::900438541234:role/app/gati-test/test/gati-test-root-role-4a4cefa",
  "rootTestUser": {
    "accessKey": "AKIAVVLBPM2KETLNPGC3",
    "arn": "arn:aws:iam::900438541234:user/app/gati-test/test/gati-test-root-test-user-3b73f9c",
    "name": "gati-test-root-test-user-3b73f9c",
    "repoSecrets": [
      {
        "accessKeyGithubSecret": "TEST_USER_ACCESS_KEY",
        "secretKeyGithubSecret": "TEST_USER_SECRET_KEY"
      },
      {
        "accessKeyGithubSecret": "TEST_USER_ACCESS_KEY",
        "secretKeyGithubSecret": "TEST_USER_SECRET_KEY"
      },
      {
        "accessKeyGithubSecret": "TEST_USER_ACCESS_KEY",
        "secretKeyGithubSecret": "TEST_USER_SECRET_KEY"
      }
    ],
    "secretKey": "[secret]"
  }
}
```

If we want to find a role that has write for `actions`/`checks`/`contents`/`pull_requests` we can use the following command role based on the printed role mappings:

```json
 "AROAVVLBPM2KO2E5AR74O": {
      "permissions": {
        "actions": "write",
        "checks": null,
        "contents": "write",
        "pull_requests": "write"
      },
      "repositories": [
        "gati-test-0-e8897fb",
        "gati-test-1-90d91e3",
        "gati-test-2-5dd95d3"
      ]
    }
```

Then, we can edit our aws config to use the role we want to use:

```bash
# in ~/.aws/config, assuming you have a chainlink-sandbox profile
# which is the user that you deployed the GATI service with

# First assume the root test role that lets us assume all other test roles
[profile gati-test]
role_arn = arn:aws:iam::900438541234:role/app/gati-test/test/gati-test-root-role-4a4cefa
source_profile = chainlink-sandbox

# Then assume the role we want to use
[profile gati-test-3]
role_arn = arn:aws:iam::900438541234:role/app/gati-test/test/gati-test-test-3-role-c941f3d
source_profile = gati-test
```

Then, we can use the `aws-vault` tool to spawn a shell with the role we want to use:

```bash
# Assume our role using aws-vault to spawn a shell
aws-vault exec gati-test-3 -- bash
```

## Running the action

```bash
# Get the lambda url from the pulumi stack output
pulumi stack -C deploy/ output -s e2e --json | jq '.lambdas[0].lambda.url'
> https://asdf3s90823fwaf.lambda-url.us-west-2.on.aws/

# Run the action
LAMBDA_URL=https://<your-lambda-url> pnpm nx run action:start

# Test using the token locally using `gh`
echo <token> | gh auth login --with-token

gh auth status

gh repo ls
```
