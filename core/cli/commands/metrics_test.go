// Copyright 2016-2025 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0
//
//                                 /$$$$$$  /$$                                     /$$
//                               /$$__  $$|__/                                    | $$
//   /$$$$$$$  /$$$$$$  /$$$$$$$ | $$  \__/ /$$  /$$$$$$  /$$$$$$/$$$$   /$$$$$$  /$$$$$$    /$$$$$$
//  /$$_____/ /$$__  $$| $$__  $$| $$$$    | $$ /$$__  $$| $$_  $$_  $$ |____  $$|_  $$_/   /$$__  $$
// | $$      | $$  \ $$| $$  \ $$| $$_/    | $$| $$  \__/| $$ \ $$ \ $$  /$$$$$$$  | $$    | $$$$$$$$
// | $$      | $$  | $$| $$  | $$| $$      | $$| $$      | $$ | $$ | $$ /$$__  $$  | $$ /$$| $$_____/
// |  $$$$$$$|  $$$$$$/| $$  | $$| $$      | $$| $$      | $$ | $$ | $$|  $$$$$$$  |  $$$$/|  $$$$$$$
// \_______/ \______/ |__/  |__/|__/      |__/|__/      |__/ |__/ |__/ \_______/   \___/   \_______/
//
// This file is part of Confirmate Core.

package commands_test

import (
	"testing"

	"confirmate.io/core/cli/commandstest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"
)

func TestMetricsCommands(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		output, err := commandstest.RunCLI(t, "metrics", "list")
		assert.NoError(t, err)
		assert.Contains(t, output, orchestratortest.MockMetricId1)
		assert.Contains(t, output, orchestratortest.MockMetricId2)
	})

	t.Run("get", func(t *testing.T) {
		output, err := commandstest.RunCLI(t, "metrics", "get", orchestratortest.MockMetricId1)
		assert.NoError(t, err)
		assert.Contains(t, output, orchestratortest.MockMetricId1)
	})

	t.Run("remove", func(t *testing.T) {
		output, err := commandstest.RunCLI(t, "metrics", "remove", orchestratortest.MockMetricId2)
		assert.NoError(t, err)
		assert.Contains(t, output, "Metric "+orchestratortest.MockMetricId2+" deleted successfully")
	})
}
