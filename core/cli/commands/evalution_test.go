package commands_test

import (
	"testing"

	"confirmate.io/core/cli/commandstest"
	"confirmate.io/core/util/assert"
)

func TestEvaluationCommands(t *testing.T) {
	t.Run("list-tools", func(t *testing.T) {
		// The cache starts empty on a fresh server, so the response is an
		// empty JSON object. We verify the command succeeds and produces output.
		output, err := commandstest.RunCLI(t, "evaluation", "list")
		assert.NoError(t, err)
		assert.NotNil(t, output)
	})
}
