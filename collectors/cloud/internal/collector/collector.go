// Copyright 2016-2026 Fraunhofer AISEC
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

package collector

import (
	"encoding/json"
	"reflect"

	"confirmate.io/core/api/ontology"
)

// Collector takes care of collecting provider resources and translating them into ontology objects.
type Collector interface {
	Name() string
	ID() string
	Collect() ([]ontology.IsResource, error)
	List() ([]ontology.IsResource, error)
	TargetOfEvaluationID() string
}

// Raw serializes provider-native objects into the ontology raw payload field.
func Raw(raws ...any) string {
	var rawMap = make(map[string][]any)

	for _, raw := range raws {
		typ := reflect.TypeOf(raw).String()

		rawMap[typ] = append(rawMap[typ], raw)
	}

	b, _ := json.Marshal(rawMap)
	return string(b)
}
