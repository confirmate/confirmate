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

package prototest

import (
	"testing"

	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/assert"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// NewAny creates a new [*anypb.Any] from a [proto.Message] with an assert that no error has been thrown.
func NewAny(t *testing.T, m proto.Message) *anypb.Any {
	a, err := anypb.New(m)
	assert.NoError(t, err)

	return a
}

// NewAny creates a new [*anypb.Any] from a [proto.Message] with a panic if an error has been thrown.
func NewAnyWithPanic(m proto.Message) *anypb.Any {
	a, err := anypb.New(m)
	if err != nil {
		panic(err)
	}

	return a
}

func NewProtobufResource(t *testing.T, or ontology.IsResource) *ontology.Resource {
	r := ontology.ProtoResource(or)
	assert.NotNil(t, r)

	return r
}
