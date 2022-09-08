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
              # fail if the workflow did not return success
              if [ "${{ steps.poll_pass.outputs.conclusion }}" != "success" ]; then
                exit 1
              fi
```

Requires a token that has action:write as well as the ability to read workflows for the type of repository you try to use.

This action provides the final status, conclusion, and workflow_id of the started workflow. It does not fail if the external workflow failed, it will provide the information in outputs instead so you can do what you need with them, whether that be to fail the workflow or gather logs, or whatever you need to do based on the results.

## Workflow reruns

If you need to rerun the private workflow, do so with the workflow running this action and not the workflow this action starts otherwise you will not get an updated result.
