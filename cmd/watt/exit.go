package main

const (
	EXIT_SUCCESS          = 0
	EXIT_INVALID_PIPELINE = 2
	EXIT_USAGE            = EXIT_INVALID_PIPELINE
	EXIT_INTERNAL_ERROR   = 5
)

type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string {
	return e.err.Error()
}

func (e *exitError) Unwrap() error {
	return e.err
}
