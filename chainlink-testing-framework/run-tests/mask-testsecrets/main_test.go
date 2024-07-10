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
	}{
		{
			name:     "Simple Key-Value",
			envStr:   "KEY=VALUE\nANOTHER_KEY=VALUE",
			expected: map[string]string{"KEY": "VALUE", "ANOTHER_KEY": "VALUE"},
		},
		{
			name:     "Export Prefix",
			envStr:   "export KEY=VALUE\nexport ANOTHER_KEY=VALUE",
			expected: map[string]string{"KEY": "VALUE", "ANOTHER_KEY": "VALUE"},
		},
		{
			name:     "Values with Quotes",
			envStr:   "KEY=\"some value\"\nANOTHER_KEY='another value'",
			expected: map[string]string{"KEY": "some value", "ANOTHER_KEY": "another value"},
		},
		{
			name: "With Comments",
			envStr: `
# This is a comment
export VAR1="Value1"
#VAR2='Value2'
VAR3=Value3
# export VAR4="Value4"
`,
			expected: map[string]string{"VAR1": "Value1", "VAR3": "Value3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseEnvVars(tt.envStr)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseEnvVars(%q) = %v, want %v", tt.envStr, result, tt.expected)
			}
		})
	}
}
