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

package api

import (
	"google.golang.org/protobuf/proto"
)

// PayloadRequest describes any kind of requests that carries a certain payload.
// This is for example a Create/Update request carrying an embedded message,
// which should be created or updated.
type PayloadRequest interface {
	GetPayload() proto.Message
	proto.Message
}

// HasId interface for messages that have an ID field.
type HasId interface {
	GetId() string
}

// HasTargetOfEvaluationId interface for messages that have a target_of_evaluation_id field.
type HasTargetOfEvaluationId interface {
	GetTargetOfEvaluationId() string
}

// HasToolId interface for messages that have a tool_id field.
type HasToolId interface {
	GetToolId() string
}
