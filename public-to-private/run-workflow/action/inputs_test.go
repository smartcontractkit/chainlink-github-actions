package action_test

import (
	"testing"

	"github.com/smartcontractkit/chainlink-github-actions/public-to-private/run-workflow/action"
	"github.com/stretchr/testify/assert"
)

func TestGetInputsSetsAllValues(t *testing.T) {
	a := &action.Inputs{}
	if len(a.Inputs) == 0 {
		t.Fail()
	}
	assert.NotEqual(t, "", a.Inputs)
}
