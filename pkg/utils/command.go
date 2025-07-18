package utils

import (
	"fmt"
	"os/exec"
)

// OutputValidator is a simple type alias for functions
// that validate the output in the shell struct
type OutputValidator func(string) bool

// We generalize the shell to use it with different validators
type Shell struct {
	validators        []OutputValidator
	onValidationError func(string) error
}

// Functional options pattern:
// https://golang.cafe/blog/golang-functional-options-pattern.html
type ShellCommandOpt func(*Shell)

// NewShell instantiates a new shell object using the
// functional options pattern and returns a pointer to it
func NewShell(opts ...ShellCommandOpt) *Shell {
	validators := []OutputValidator{}
	onValidationError := func(out string) error {
		return fmt.Errorf("validation error, output: %s", string(out))
	}

	shell := &Shell{
		validators:        validators,
		onValidationError: onValidationError,
	}

	for _, opt := range opts {
		opt(shell)
	}

	return shell
}

// WithValidators is used to provide an arbitrary number
// of output validators to the new shell constructor
func WithValidators(validators ...OutputValidator) ShellCommandOpt {
	return func(s *Shell) {
		s.validators = validators
	}
}

// WithOnValidationError is used to provide a custom
// onValidationError function
func WithOnValidationError(onValidationError func(string) error) ShellCommandOpt {
	return func(s *Shell) {
		s.onValidationError = onValidationError
	}
}

// Run runs a command in the shell.
// If an error raised from running the command or any of the shell validators
// fails validating the command output, an error is returned from this method
func (s *Shell) Run(cmd string, args ...string) ([]byte, error) {
	command := exec.Command(cmd, args...)

	output, err := command.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running %s command: %w, output: %s", cmd, err, string(output))
	}

	// Each validator is tested on the output
	for _, validator := range s.validators {
		if !validator(string(output)) {
			return nil, s.onValidationError(string(output))
		}
	}

	return output, nil
}
