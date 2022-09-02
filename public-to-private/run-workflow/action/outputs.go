package action

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sethvargo/go-githubactions"
)

type Outputs struct {
	Status     string
	Conclusion string
	WorkflowID int64
}

// setOutputs Sets the outputs in the format that a docker action can parse
func (o *Outputs) SetOutputs(githubAction *githubactions.Action) {
	ao := *o
	val := reflect.ValueOf(ao)
	typeOfS := val.Type()

	for i := 0; i < val.NumField(); i++ {
		k := strings.ToLower(typeOfS.Field(i).Name)
		v := fmt.Sprintf("%v", val.Field(i).Interface())
		fmt.Printf("Setting output: %s = %s\n", k, v)
		githubAction.SetOutput(k, v)
	}
}
