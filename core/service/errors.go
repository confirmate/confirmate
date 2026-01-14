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

package service

import (
	"errors"
	"fmt"

	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/util"

	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// validator is reused for all validation calls.
var validator protovalidate.Validator

func init() {
	var err error

	validator, err = protovalidate.New()
	if err != nil {
		panic(fmt.Sprintf("failed to create protovalidate validator: %v", err))
	}
}

var (
	// ErrEmptyRequest is returned when a nil request is passed.
	ErrEmptyRequest = errors.New("empty request")

	// IgnoreIDFilter is a validation filter that skips validation of "id" fields.
	// Useful for Create operations where the ID is auto-generated after initial validation.
	IgnoreIDFilter = protovalidate.FilterFunc(func(msg protoreflect.Message, desc protoreflect.Descriptor) bool {
		// Return false to skip validation of fields named "id"
		// Return true to validate all other fields
		if fd, ok := desc.(protoreflect.FieldDescriptor); ok {
			return fd.Name() != "id"
		}
		return true
	})
)

// ErrNotFound returns a plain error with the given entity name.
// This error is meant to be wrapped by [HandleDatabaseError] which converts it to a [connect.CodeNotFound] error.
func ErrNotFound(entity string) error {
	return fmt.Errorf("%s not found", entity)
}

// Validate validates an incoming request using protovalidate.
// The type parameter T should be a protobuf message type where *T implements [proto.Message].
//   - If the request or request message is nil, it returns an [ErrEmptyRequest] error.
//   - Accepts optional validation options (e.g., WithFilter to ignore specific fields).
//   - If the request fails validation, it returns a [connect.CodeInvalidArgument] error.
func Validate[T any](req *connect.Request[T], opts ...protovalidate.ValidationOption) error {
	if util.IsNil(req) || util.IsNil(req.Msg) {
		return connect.NewError(connect.CodeInvalidArgument, ErrEmptyRequest)
	}

	// req.Msg is expected to be a proto.Message
	msg, ok := any(req.Msg).(proto.Message)
	if !ok {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("request message does not implement proto.Message"))
	}

	if err := validator.Validate(msg, opts...); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid request: %w", err))
	}

	return nil
}

// ValidateWithPrep validates a request with a preparation function that runs after
// nil checks but before validation. This is useful when the request needs modification
// (e.g., setting auto-generated UUIDs) before validation can pass.
// The type parameter T should be a protobuf message type where *T implements [proto.Message].
//   - If the request or request message is nil, it returns an [ErrEmptyRequest] error.
//   - If the prep function is not nil, it is called before validation.
//   - Accepts optional validation options (e.g., WithFilter to ignore specific fields).
//   - If the request fails validation, it returns a [connect.CodeInvalidArgument] error.
func ValidateWithPrep[T any](req *connect.Request[T], prep func(), opts ...protovalidate.ValidationOption) error {
	if util.IsNil(req) || util.IsNil(req.Msg) {
		return connect.NewError(connect.CodeInvalidArgument, ErrEmptyRequest)
	}

	// Execute preparation function if provided
	if prep != nil {
		prep()
	}

	// req.Msg is expected to be a proto.Message
	msg, ok := any(req.Msg).(proto.Message)
	if !ok {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("request message does not implement proto.Message"))
	}

	if err := validator.Validate(msg, opts...); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid request: %w", err))
	}

	return nil
}

// HandleDatabaseError translates database errors into appropriate connect errors.
//   - If the error is [persistence.ErrRecordNotFound], it returns a [connect.CodeNotFound]
//     error with the provided notFoundErr (or a default error if not provided).
//   - If err is already a [connect.Error], it returns it as-is.
//   - For other errors, it returns a [connect.CodeInternal] error. If err is nil, it returns nil.
func HandleDatabaseError(err error, notFoundErr ...error) error {
	if err == nil {
		return nil
	}

	// If it's already a [connect.Error], return it as-is
	var connectErr *connect.Error
	if errors.As(err, &connectErr) {
		return err
	}

	if errors.Is(err, persistence.ErrRecordNotFound) {
		if len(notFoundErr) == 0 {
			notFoundErr = append(notFoundErr, ErrNotFound("entity"))
		}

		return connect.NewError(connect.CodeNotFound, notFoundErr[0])
	}

	return connect.NewError(connect.CodeInternal, fmt.Errorf("database error: %w", err))
}

// ValidateEvent validates a ChangeEvent using the shared validator.
// Returns a connect error with CodeInvalidArgument on validation failures.
func ValidateEvent(ce *orchestrator.ChangeEvent) error {
	if ce == nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("empty change event"))
	}

	// Use protovalidate for validation. The oneof is now marked as required
	// in the proto definition, so this will catch missing event fields.
	if err := validator.Validate(ce); err != nil {
		return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid message: %w", err))
	}

	return nil
}
