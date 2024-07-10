package main

import (
	"fmt"
	"os"
	"regexp"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <base64_dot_env_string>")
		return
	}
	fmt.Println(mustResolveEnvPlaceholder(os.Args[1]))
}

// mustResolveEnvPlaceholder checks if the input string is an environment variable placeholder and resolves it.
func mustResolveEnvPlaceholder(input string) string {
	envVarName, hasEnvVar := lookupEnvVarName(input)
	if hasEnvVar {
		value, set := os.LookupEnv(envVarName)
		if !set {
			fmt.Fprintf(os.Stderr, "Error resolving '%s'. Environment variable '%s' not set or is empty\n", input, envVarName)
			os.Exit(1)
		}
		return value
	}
	return input
}

func lookupEnvVarName(input string) (string, bool) {
	re := regexp.MustCompile(`^{{ env\.([a-zA-Z_]+) }}$`)
	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		return matches[1], true
	}
	return "", false
}
