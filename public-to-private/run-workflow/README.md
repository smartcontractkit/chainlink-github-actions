# Example usage for both starting and polling a workflow
    jobs:
      tester:
        name: Start and poll the test
        runs-on: ubuntu-latest
        environment: integration
        steps:
          - name: Checkout the repo
            uses: actions/checkout@2541b1294d2704b0964813337f33b291d3f8596b # v3.0.2
          - uses: ./public-to-private/run-workflow
            id: poll
            with:
              token: ${{ secrets.TMP_TEST_TOKEN }}
              repository: chainlink
              ref: qa_test_workflow_run
              workflow_file: integration-tests.yml
              timeout: 30m
          - name: get outputs
            run: |
              echo "status = ${{ steps.poll.outputs.status }}"
              echo "conclusion = ${{ steps.poll.outputs.conclusion }}"

Requires a token that has action:write as well as the ability to read workflows for the type of repository you try to use