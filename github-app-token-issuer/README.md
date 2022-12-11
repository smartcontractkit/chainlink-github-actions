# Action

- [Action](#action)
- [Usage](#usage)


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
          node-version: 16

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
