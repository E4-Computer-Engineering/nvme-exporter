package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

// getStringCmd produces a single string from the cmd, args... format.
func getStringCmd(cmd string, args ...string) string {
	cmdSlice := []string{cmd}
	cmdSlice = append(cmdSlice, args...)

	return strings.Join(cmdSlice, " ")
}

// ExecuteCommand executes a command and returns a nicely-formatted error if it fails.
func ExecuteCommand(cmd string, args ...string) (string, error) {
	_, err := exec.LookPath(cmd)
	if err != nil {
		return "", fmt.Errorf("error looking for %s cli command in path: %w", cmd, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, cmd, args...)

	out, err := command.CombinedOutput()
	if err != nil {
		cmdString := getStringCmd(cmd, args...)

		return string(out), fmt.Errorf("error running command %s: %w", cmdString, err)
	}

	return string(out), nil
}

// ExecuteJSONCommand executes a command, validates the JSON output, and returns
// the parsed gjson.Result object.
func ExecuteJSONCommand(cmd string, args ...string) (gjson.Result, error) {
	output, err := ExecuteCommand(cmd, args...)
	if err != nil {
		return gjson.Result{}, err
	}

	if !gjson.Valid(output) {
		cmdString := getStringCmd(cmd, args...)

		return gjson.Result{}, fmt.Errorf("invalid JSON output from %s command: %s", cmdString, output)
	}

	ret := gjson.Parse(output)

	return ret, nil
}

// CheckRoot verifies the process is running with root privileges by
// checking the effective user ID. This is more reliable than comparing
// the username, which misses uid-0 accounts not literally named "root".
func CheckRoot() error {
	if euid := os.Geteuid(); euid != 0 {
		return fmt.Errorf("effective uid is %d, not 0 (root)", euid)
	}

	return nil
}

// MapToJSONString converts a map to a JSON string.
func MapToJSONString(data map[string]interface{}) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return "{}"
	}

	return string(jsonBytes)
}
