package provider

import "fmt"

// CLIError represents an error from a CLI provider.
type CLIError struct {
	// Provider is the name of the provider that encountered the error.
	Provider string

	// Message is a human-readable error message.
	Message string

	// Err is the underlying error (if any).
	Err error
}

// Error implements the error interface.
func (e *CLIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s provider error: %s: %v", e.Provider, e.Message, e.Err)
	}
	return fmt.Sprintf("%s provider error: %s", e.Provider, e.Message)
}

// Unwrap returns the underlying error.
func (e *CLIError) Unwrap() error {
	return e.Err
}
