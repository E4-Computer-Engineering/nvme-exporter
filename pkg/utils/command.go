package utils

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/tidwall/gjson"
)

func getStringCmd(cmd string, args ...string) string {
	cmdSlice := []string{cmd}
	cmdSlice = append(cmdSlice, args...)

	return strings.Join(cmdSlice, " ")
}

func ExecuteCommand(cmd string, args ...string) (string, error) {
	command := exec.Command(cmd, args...)

	out, err := command.CombinedOutput()
	if err != nil {
		cmdString := getStringCmd(cmd, args...)

		return string(out), fmt.Errorf("error running command %s: %w", cmdString, err)
	}

	return string(out), nil
}

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
