package clitest

import (
	"errors"
	"os"
)

// AutoChdir automatically guesses if we need to change the current working directory
// so that we can find the policies folder
func AutoChdir() {
	var (
		err error
	)

	// Check, if we can find the core folder
	_, err = os.Stat("policies")
	if errors.Is(err, os.ErrNotExist) {
		// Try again one level deeper
		err = os.Chdir("..")
		if err != nil {
			panic(err)
		}

		AutoChdir()
	}
}
