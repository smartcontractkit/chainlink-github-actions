package action

import (
	"encoding/json"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/sethvargo/go-githubactions"
)

type Inputs struct {
	GithubToken   string        `envconfig:"token" required:"true"`
	Owner         string        `envconfig:"owner" default:"smartcontractkit"`
	Repository    string        `envconfig:"repository" required:"true"`
	Branch        string        `envconfig:"ref" required:"true"`
	WorkflowFile  string        `envconfig:"workflow_file" required:"true"`
	Timeout       time.Duration `envconfig:"timeout" default:"5m"`
	RetryInterval time.Duration `envconfig:"retry_interval" default:"15s"`
	Inputs        string        `envconfig:"inputs" default:"{}"`
}

func (i *Inputs) InputsToMap(action *githubactions.Action) map[string]interface{} {
	inputsMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(i.Inputs), &inputsMap)
	if err != nil {
		action.Fatalf("could not unmarshall inputs json: %v", err)
	}
	return inputsMap
}

// getInputs Loads inputs from the environment
func GetInputs(action *githubactions.Action) *Inputs {
	var inputs Inputs
	if err := envconfig.Process("", &inputs); err != nil {
		action.Fatalf("could not load inputs from environment: %v", err)
	}

	return &inputs
}
