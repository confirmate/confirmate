package commands_test

import (
	"testing"

	"confirmate.io/core/cli/commandstest"
	"confirmate.io/core/service/evaluation/evaluationtest"
	"confirmate.io/core/util/assert"
)

func TestEvaluationCommands(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		output, err := commandstest.RunCLI(t, "evaluation", "list")
		assert.NoError(t, err)
		assert.Contains(t, output, evaluationtest.MockEvaluationResult1.GetId())
	})
}
