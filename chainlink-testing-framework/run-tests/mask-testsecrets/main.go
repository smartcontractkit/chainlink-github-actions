package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type DotenvParseError struct {
	Line int // Line where the error occurred
}

// Error implements the error interface for DotenvParseError.
func (e *DotenvParseError) Error() string {
	return fmt.Sprintf("could not parse dotenv due to invalid KEY=\"VALUE\" in line: %d. Only KEY=VALUE or KEY=\"VALUE\" are supported. To include special characters such as an apostrophe in the VALUE, enclose the VALUE in double quotes (e.g., KEY=\"'VALUE'\")", e.Line)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <base64_dot_env_string>")
		return
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not decode provided base64 string")
		os.Exit(1)
	}

	envVars, err := parseEnvVars(string(decodedBytes))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for key, value := range envVars {
		mustMaskSecret(key, value)
		mustAddToGithubEnv(key, value)
	}
}

func parseEnvVars(envStr string) (map[string]string, error) {
	envVars := make(map[string]string)
	// Regular expression to match configurations in the form of KEY=VALUE or KEY="VALUE".
	// To include special characters such as an apostrophe in the VALUE, enclose the VALUE in double quotes (e.g., KEY="'VALUE'").
	// Restrictions:
	// - KEY must not be enclosed in any type of quotes.
	// - VALUE can be unquoted or enclosed in double quotes, but not single quotes.
	re := regexp.MustCompile(`^[^" '=]+=".*"$|^[^" '=]+=[^" ']+$`)

	lines := strings.Split(envStr, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line) // Trim space from start and end
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}
		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export"))
		}

		// Validate the line with regex before processing
		if re.MatchString(line) {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := removeSurroundingQuotes(strings.TrimSpace(parts[1]))
				envVars[key] = value
			}
		} else {
			return nil, &DotenvParseError{Line: i + 1}
		}
	}
	return envVars, nil
}

func removeSurroundingQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
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
