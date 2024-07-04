package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run script.go <base64_dot_env_string>")
		return
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(os.Args[1])
	if err != nil {
		fmt.Println("Error:", err)
	}

	envVars := parseEnvVars(string(decodedBytes))

	for key, value := range envVars {
		mustMaskSecret(key, value)
		mustAddToGithubEnv(key, value)
	}
}

func parseEnvVars(envStr string) map[string]string {
	envVars := make(map[string]string)
	lines := strings.Split(envStr, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := strings.TrimSpace(parts[1]) // Trim spaces from the value
			value = removeDoubleQuotes(value)
			value = removeSingleQuotes(value)
			envVars[key] = value
		}
	}
	return envVars
}

func mustAddToGithubEnv(key, value string) {
	command := fmt.Sprintf("echo \"%s=%s\" >> $GITHUB_ENV", key, value)
	cmd := exec.Command("bash", "-c", command)
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add '%s' to GITHUB_ENV", key)
		os.Exit(1)
	}
}

func mustMaskSecret(description string, secret string) {
	if secret != "" {
		fmt.Printf("Mask '%s'\n", description)
		fmt.Printf("::add-mask::%s\n", secret)
		cmd := exec.Command("bash", "-c", "echo ::add-mask::$0", "_", secret)
		err := cmd.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to mask secret '%s'", description)
			os.Exit(1)
		}
	}
}

// removeDoubleQuotes checks if a string is enclosed in double quotes and removes them.
func removeDoubleQuotes(value string) string {
	if len(value) >= 2 && strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
		return value[1 : len(value)-1]
	}
	return value
}

// removeSingleQuotes checks if a string is enclosed in single quotes and removes them.
func removeSingleQuotes(value string) string {
	if len(value) >= 2 && strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
		return value[1 : len(value)-1]
	}
	return value
}
