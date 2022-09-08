# Example usage for starting and polling a workflow

``` yaml
    jobs:
      tester:
        name: Start and poll the workflow
        runs-on: ubuntu-latest
        environment: integration
        steps:
          - name: Checkout the repo
            uses: actions/checkout@v3
          - uses: ./public-to-private/run-workflow
            id: poll
            with:
              token: ${{ github.token }}
              repository: chainlink
              ref: example_branch_name
              workflow_file: integration-tests.yml
              timeout: 30m
              inputs: "{\"example_input\":\"example value\"}"
          - name: get outputs
            run: |
              echo "status = ${{ steps.poll.outputs.status }}"
              echo "conclusion = ${{ steps.poll.outputs.conclusion }}"
              echo "workflow_id = ${{ steps.poll_pass.outputs.workflow_id }}"
```

Requires a token that has action:write as well as the ability to read workflows for the type of repository you try to use. 

If you need to do more with the results of the pipline you can go to that repos actions and past the workflow_id to see the results. You can also use the workflow_id with the github api to get whatever it is you need so long as the provided github token has permissions to do so.
