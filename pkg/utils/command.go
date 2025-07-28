package utils

import (
	"fmt"
	"os/exec"
	"strings"

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
	command := exec.Command(cmd, args...)

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
