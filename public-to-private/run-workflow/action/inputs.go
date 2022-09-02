package action

import (
	"encoding/json"
	"time"

	"github.com/kelseyhightower/envconfig"
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

func (i *Inputs) InputsToMap() (map[string]interface{}, error) {
	inputsMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(i.Inputs), &inputsMap)
	return inputsMap, err
}

// getInputs Loads inputs from the environment
func GetInputs() (*Inputs, error) {
	var inputs Inputs
	err := envconfig.Process("", &inputs)
	return &inputs, err
}
