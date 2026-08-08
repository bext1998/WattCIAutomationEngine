package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	os.Exit(execute(newRootCommand()))
}

func execute(command *cobra.Command) int {
	if err := command.Execute(); err != nil {
		fmt.Fprintln(command.ErrOrStderr(), err)

		var commandError *exitError
		if errors.As(err, &commandError) {
			return commandError.code
		}
		return EXIT_INTERNAL_ERROR
	}

	return EXIT_SUCCESS
}
