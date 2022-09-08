package helpers

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestInputsOrOutputsMatchAction(t *testing.T, actionFilePath string, structToTest interface{}, fieldToValidateAgainst string) {
	yamlFile, err := ioutil.ReadFile(actionFilePath)
	assert.Nil(t, err, "Should have no error reading the action.yml")

	var c map[string]interface{}
	err = yaml.Unmarshal(yamlFile, &c)
	assert.Nil(t, err, "Should have no error parsing the action.yml")

	expectedInputs := c[fieldToValidateAgainst].(map[interface{}]interface{})

	// Get all the keys in the inputs of the action file
	expectedKeys := []string{}
	for k, _ := range expectedInputs {
		ks := fmt.Sprintf("%v", k)
		expectedKeys = append(expectedKeys, ks)
	}

	// Verify all the input fields of the action file line up with the envconfigs of the Inputs type
	val := reflect.ValueOf(structToTest)
	typeOfS := val.Type()
	for i := 0; i < val.NumField(); i++ {
		k := fmt.Sprintf("%v", typeOfS.Field(i).Tag.Get("envconfig"))
		fmt.Print(k)
		assert.Contains(t, expectedKeys, k, "The key did not exist in the expected keys slice")
	}

	// Verify the number of inputs is the same length as the inputs struct
	assert.Equal(t, len(expectedInputs), val.NumField(), "Number of '%s' did not match the number of elements in the struct", fieldToValidateAgainst)
}
