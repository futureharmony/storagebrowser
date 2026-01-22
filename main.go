package main

import (
	"os"

	"github.com/futureharmony/storagebrowser/v2/cmd"
	"github.com/futureharmony/storagebrowser/v2/errors"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(errors.GetExitCode(err))
	}
}
