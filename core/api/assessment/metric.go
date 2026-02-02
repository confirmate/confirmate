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

package assessment

import (
	"encoding/base64"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
)

// Hash provides a simple string based hash for this metric configuration. It can be used
// to provide a key for a map or a cache.
func (x *MetricConfiguration) Hash() string {
	return base64.RawURLEncoding.EncodeToString([]byte(fmt.Sprintf("%v-%v", x.Operator, x.TargetValue)))
}

func (x *MetricConfiguration) MarshalJSON() (b []byte, err error) {
	return protojson.Marshal(x)
}

func (x *MetricConfiguration) UnmarshalJSON(b []byte) (err error) {
	return protojson.Unmarshal(b, x)
}
