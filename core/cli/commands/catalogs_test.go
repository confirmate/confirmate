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

func TestCatalogsCommands(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		output, err := commandstest.RunCLI(t, "catalogs", "list")
		assert.NoError(t, err)
		assert.Contains(t, output, orchestratortest.MockCatalogId1)
		assert.Contains(t, output, orchestratortest.MockCatalogId2)
	})

	t.Run("get", func(t *testing.T) {
		output, err := commandstest.RunCLI(t, "catalogs", "get", orchestratortest.MockCatalogId1)
		assert.NoError(t, err)
		assert.Contains(t, output, orchestratortest.MockCatalogId1)
	})

	t.Run("delete", func(t *testing.T) {
		output, err := commandstest.RunCLI(t, "catalogs", "delete", orchestratortest.MockCatalogId2)
		assert.NoError(t, err)
		assert.Contains(t, output, "Catalog "+orchestratortest.MockCatalogId2+" deleted successfully")
	})
}
