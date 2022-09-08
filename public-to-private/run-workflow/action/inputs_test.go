package action_test

import (
	"os"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink-github-actions/public-to-private/run-workflow/action"
	"github.com/smartcontractkit/chainlink-github-actions/public-to-private/run-workflow/helpers"
	"github.com/stretchr/testify/assert"
)

func TestGetInputsSetsAllValues(t *testing.T) {
	expectedDuration, _ := time.ParseDuration("1s")
	os.Setenv("TOKEN", "token")
	os.Setenv("OWNER", "owner")
	os.Setenv("REPOSITORY", "repo")
	os.Setenv("REF", "branch_name")
	os.Setenv("WORKFLOW_FILE", "bam.yml")
	os.Setenv("TIMEOUT", "1s")
	os.Setenv("RETRY_INTERVAL", "1s")
	os.Setenv("INPUTS", "{}")
	i, err := action.GetInputs()
	assert.NoError(t, err, "Received an error while getting inputs")
	assert.NotNil(t, i)
	assert.Equal(t, "token", i.GithubToken)
	assert.Equal(t, "owner", i.Owner)
	assert.Equal(t, "repo", i.Repository)
	assert.Equal(t, "branch_name", i.Branch)
	assert.Equal(t, "bam.yml", i.WorkflowFile)
	assert.Equal(t, expectedDuration, i.Timeout)
	assert.Equal(t, expectedDuration, i.RetryInterval)
	assert.Equal(t, "{}", i.Inputs)
}

func TestInputsToMap(t *testing.T) {
	i := &action.Inputs{Inputs: "{\"a\": \"b\"}"}
	m, _ := i.InputsToMap()
	assert.NotNil(t, m)
	assert.Equal(t, "b", m["a"], "failed to convert the inputs to a map %v", m)
}

func TestInputsMatchAction(t *testing.T) {
	input := &action.Inputs{}
	helpers.TestInputsOrOutputsMatchAction(t, "../action.yml", *input, "inputs")
}
