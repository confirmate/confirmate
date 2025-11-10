// Copyright 2016-2025 Fraunhofer AISEC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commands

import (
	"context"

	"confirmate.io/core/api/evidence/evidenceconnect"
	"confirmate.io/core/server"
	"confirmate.io/core/service/evidence"

	"github.com/mfridman/cli"
)

var EvidenceCommand = &cli.Command{
	Name: "evidence store",
	Exec: func(ctx context.Context, s *cli.State) error {
		svc, err := evidence.NewService()
		if err != nil {
			return err
		}

		return server.RunConnectServer(evidenceconnect.NewEvidenceStoreHandler(svc))
	},
}
