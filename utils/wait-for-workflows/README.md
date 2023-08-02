# wait-for-workflows action

```yaml
name: example

on:
  merge_group:
  pull_request:

jobs:
  waitForWorkflows:
    name: Wait for workflows
    runs-on: ubuntu-latest
    if: always()
    steps:
      - name: Checkout repository
        uses: actions/checkout@v3
        with:
          ref: ${{ github.event.pull_request.head.sha || github.event.merge_group.head_sha }}

      - name: Wait for workflows
        id: wait
        uses: smartcontractkit/chainlink-github-actions/utils/wait-for-workflows@main
        with:
          max-timeout: "900"
          polling-interval: "30"
          exclude-workflow-names: ""
          exclude-workflow-ids: ""
          github-token: ${{ secrets.GITHUB_TOKEN }}
        env:
          DEBUG: "true"

  afterWait:
    name: after-wait
    needs: [waitForWorkflows]
    runs-on: ubuntu-latest
    if: always()
    steps:
      - name: Check needs results
        if: needs.waitForWorkflows.result != 'success'
        run: exit 1
```
