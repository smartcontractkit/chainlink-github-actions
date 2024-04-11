# [smartcontractkit/.github](https://github.com/smartcontractkit/.github/) is Chainlink's new monorepo for reusable public actions.

The plan is to eventually migrate all actions in this repository to the `.github` repository. Please refrain from making new actions here if possible.

## chainlink-github-actions

Place your action in a logically named folder with the action.yml and README.md and anything else required for your action.

Currently we will be versioning all actions at once with github releases/tags. In the future we can look into using something like `changesets` which is used by external-adapters-js to version independent actions.