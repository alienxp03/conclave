package provider

import "fmt"

// CLIError represents an error from a CLI provider.
type CLIError struct {
	Provider string
	Message  string
	Err      error
}

// Error implements the error interface.
func (e *CLIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s CLI error: %s (%v)", e.Provider, e.Message, e.Err)
	}
	return fmt.Sprintf("%s CLI error: %s", e.Provider, e.Message)
}

// Unwrap returns the underlying error.
func (e *CLIError) Unwrap() error {
	return e.Err
}
