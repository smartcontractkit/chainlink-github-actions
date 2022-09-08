package action

import (
	"fmt"
	"reflect"

	"github.com/sethvargo/go-githubactions"
)

type Outputs struct {
	Status     string `envconfig:"status"`
	Conclusion string `envconfig:"conclusion"`
	WorkflowID int64  `envconfig:"workflow_id"`
}

// setOutputs Sets the outputs in the format that a docker action can parse
func (o *Outputs) SetOutputs(githubAction *githubactions.Action) {
	ao := *o
	val := reflect.ValueOf(ao)
	typeOfS := val.Type()

	for i := 0; i < val.NumField(); i++ {
		k := typeOfS.Field(i).Tag.Get("envconfig")
		v := fmt.Sprintf("%v", val.Field(i).Interface())
		fmt.Printf("Setting output: %s = %s\n", k, v)
		githubAction.SetOutput(k, v)
	}
}
