package commands

import (
	"context"
	"os"
	"testing"
)

func TestToolsListCommand(t *testing.T) {
	cmd := ToolsListCommand()
	ctx := context.Background()
	os.Args = []string{"cf", "tools", "list"}
	if err := cmd.Action(ctx, cmd); err != nil {
		t.Errorf("tools list command failed: %v", err)
	}
}
