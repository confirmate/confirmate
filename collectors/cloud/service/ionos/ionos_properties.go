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

package ionos

import (
	"strings"

	"confirmate.io/collectors/cloud/internal/pointer"

	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

// labels converts IONOS Cloud label resources to the ontology label format.
func labels(lr ionoscloud.LabelResources) map[string]string {
	l := make(map[string]string)

	if lr.Items == nil {
		return l
	}

	for _, label := range pointer.Deref(lr.Items) {
		l[pointer.Deref(label.Properties.GetKey())] = pointer.Deref(label.Properties.GetValue())
	}

	return l
}

// hasEmptySegmentInURL checks if the URL contains an empty segment, which is indicated by two consecutive slashes
// (e.g., "http://example.com//path").
func hasEmptySegmentInURL(s string) bool {
	return strings.Contains(s, "//")
}
