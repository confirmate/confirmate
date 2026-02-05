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

package core

// //go:generate buf generate
//go:generate buf generate --exclude-path policies
//go:generate buf generate --template buf.openapi.gen.yaml --path api/evidence -o api/evidence
//go:generate buf generate --template buf.openapi.gen.yaml --path api/assessment -o api/assessment
//go:generate buf generate --template buf.openapi.gen.yaml --path api/orchestrator -o api/orchestrator
//go:generate buf generate --template buf.gen.ontology.yaml --path policies/security-metrics/ontology/v1/ontology.proto -o api/ontology
// Keep the gotag generation at the end to make sure that it isn't overridden by the other generators.
//go:generate buf generate --template buf.gotag.gen.yaml --exclude-path policies
