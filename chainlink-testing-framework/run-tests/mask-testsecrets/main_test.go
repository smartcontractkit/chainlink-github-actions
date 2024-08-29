package main

import (
	"reflect"
	"testing"
)

// TestParseEnvVars checks various scenarios for parsing environment variables.
func TestParseEnvVars(t *testing.T) {
	tests := []struct {
		name     string
		envStr   string
		expected map[string]string
		err      error
	}{
		{
			name: "Simple Key-Value",
			envStr: `KEY=VALUE
ANOTHER_KEY=VALUE2
KEY3="VALUE1 VALUE2"`,
			expected: map[string]string{"KEY": "VALUE", "ANOTHER_KEY": "VALUE2", "KEY3": "VALUE1 VALUE2"},
		},
		{
			name:     "Export Prefix",
			envStr:   "export KEY=VALUE\nexport ANOTHER_KEY=VALUE",
			expected: map[string]string{"KEY": "VALUE", "ANOTHER_KEY": "VALUE"},
		},
		{
			name: "Values with Quotes",
			envStr: `KEY="some value"
ANOTHER_KEY="another value"
KEY3="'value3'"
KEY4="'value4',"`,
			expected: map[string]string{"KEY": "some value", "ANOTHER_KEY": "another value", "KEY3": "'value3'", "KEY4": "'value4',"},
		},
		{
			name: "With Comments",
			envStr: `# This is a comment
export VAR1="Value 1"
#VAR2='Value2'
VAR3=Value3
# export VAR4="Value4"`,
			expected: map[string]string{"VAR1": "Value 1", "VAR3": "Value3"},
		},
		{
			name: "Comma after Value",
			envStr: `KEY=VALUE,
KEY2="VALUE2",`,
			err: &DotenvParseError{Line: 2},
		},
		{
			name: "Fail on Single Quotes",
			envStr: `KEY1=VALUE1
KEY2="VALUE2"
ANOTHER_KEY='VALUE'
		`,
			err: &DotenvParseError{Line: 3},
		},
		{
			name: "Key with Double Quotes",
			envStr: `KEY="VALUE"
"ANOTHER_KEY"=VALUE`,
			err: &DotenvParseError{Line: 2},
		},
		{
			name: "Key with Single Quotes",
			envStr: `KEY="VALUE"
'ANOTHER_KEY'=VALUE`,
			err: &DotenvParseError{Line: 2},
		},
		{
			name: "Value with Spaces",
			envStr: `KEY="VALUE"
KEY2=VALUE 2`,
			err: &DotenvParseError{Line: 2},
		},
		{
			name: "Empty Line",
			envStr: `KEY="VALUE"

KEY2=VALUE2`,
			expected: map[string]string{"KEY": "VALUE", "KEY2": "VALUE2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseEnvVars(tt.envStr)
			// Check for error type and line number
			if err != nil && tt.err == nil {
				t.Errorf("parseEnvVars(%q) error = %v, want nil", tt.envStr, err)
			}
			if err != nil && tt.err != nil {
				if err.Error() != tt.err.Error() {
					t.Errorf("parseEnvVars(%q) error = %v, want %v", tt.envStr, err, tt.err)
				}
			}
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseEnvVars(%q) = %v, want %v", tt.envStr, result, tt.expected)
			}
		})
	}
}
