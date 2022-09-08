package action_test

import (
	"os"
	"testing"

	"github.com/sethvargo/go-githubactions"
	"github.com/smartcontractkit/chainlink-github-actions/public-to-private/run-workflow/action"
	"github.com/smartcontractkit/chainlink-github-actions/public-to-private/run-workflow/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/zenizh/go-capturer"
)

func TestSetOutputs(t *testing.T) {
	o := &action.Outputs{
		Status:     "s",
		Conclusion: "c",
		WorkflowID: 123,
	}

	githubAction := githubactions.New(githubactions.WithWriter(os.Stdout))
	out := capturer.CaptureOutput(func() {
		o.SetOutputs(githubAction)
	})
	assert.Contains(t, out, "Setting output: status = s")
	assert.Contains(t, out, "Setting output: conclusion = c")
	assert.Contains(t, out, "Setting output: workflow_id = 123")
}

func TestOutputsMatchAction(t *testing.T) {
	output := &action.Outputs{}
	helpers.TestInputsOrOutputsMatchAction(t, "../action.yml", *output, "outputs")
}
