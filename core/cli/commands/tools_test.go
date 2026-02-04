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

func TestToolsCommands(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		output, err := commandstest.RunCLI(t, "tools", "list")
		assert.NoError(t, err)
		assert.Contains(t, output, orchestratortest.MockToolId1)
		assert.Contains(t, output, orchestratortest.MockToolId2)
	})

	t.Run("get", func(t *testing.T) {
		output, err := commandstest.RunCLI(t, "tools", "get", orchestratortest.MockToolId1)
		assert.NoError(t, err)
		assert.Contains(t, output, orchestratortest.MockToolId1)
	})

	t.Run("deregister", func(t *testing.T) {
		output, err := commandstest.RunCLI(t, "tools", "deregister", orchestratortest.MockToolId2)
		assert.NoError(t, err)
		assert.Contains(t, output, "Tool "+orchestratortest.MockToolId2+" deleted successfully")
	})
}
