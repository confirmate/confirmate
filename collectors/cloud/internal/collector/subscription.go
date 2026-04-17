package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type azAccountShow struct {
	ID string `json:"id"`
}

// ResolveSubscriptionIDFromAzureCLI resolves the active subscription from the current az CLI login context.
func ResolveSubscriptionIDFromAzureCLI(ctx context.Context) (subscriptionID string, err error) {
	cmd := exec.CommandContext(ctx, "az", "account", "show", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("reading subscription from az account show: %w", err)
	}

	var account azAccountShow
	if err = json.Unmarshal(output, &account); err != nil {
		return "", fmt.Errorf("parsing az account show output: %w", err)
	}

	subscriptionID = strings.TrimSpace(account.ID)
	if subscriptionID == "" {
		return "", fmt.Errorf("subscription id not found in az account show output")
	}

	return subscriptionID, nil
}
