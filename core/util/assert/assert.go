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

// Package assert contains helpful assertion helpers. Under the hood it currently uses
// github.com/stretchr/testify/assert, but this might change in the future. In order to keep this transparent to the
// tests, unit tests should exclusively use this package. This also helps us keep track how often the individual assert
// functions are used and whether we can reduce the API surface of this package.
package assert

import (
	"reflect"
	"testing"

	"connectrpc.com/connect"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/testing/protocmp"
)

var True = assert.True
var False = assert.False
var NotEmpty = assert.NotEmpty
var Contains = assert.Contains
var ErrorContains = assert.ErrorContains
var Error = assert.Error
var ErrorIs = assert.ErrorIs
var Fail = assert.Fail
var Same = assert.Same

type TestingT = assert.TestingT
type ErrorAssertionFunc = assert.ErrorAssertionFunc

// Want is a function type that can hold asserts in order to check the validity of "got".
type Want[T any] func(t *testing.T, got T, msgAndArgs ...any) bool

var _ Want[any] = AnyValue[any]
var _ Want[any] = Nil[any]

// WantErr is a function type that can hold asserts in order to check the error of "err".
type WantErr func(t *testing.T, err error, msgAndArgs ...any) bool

var _ WantErr = AnyValue
var _ WantErr = NoError

// NoError is a [WantErr] that asserts that no error occurred.
var NoError = Nil[error]

// CompareAllUnexported is a [cmp.Option] that allows the introspection of all un-exported fields in order to use them in
// [Equal] or [NotEqual].
func CompareAllUnexported() cmp.Option {
	return cmp.Exporter(func(reflect.Type) bool { return true })
}

// Equal asserts that [got] and [want] are Equal. Under the hood, this uses the go-cmp package in combination with
// protocmp and also supplies a diff, in case the messages to do not match.
//
// Note: By default the option protocmp.Transform() will be used. This can cause problems with structs that are NOT
// protobuf messages and contain un-exported fields. In this case, [CompareAllUnexported] can be used instead.
func Equal[T any](t TestingT, want T, got T, opts ...cmp.Option) bool {
	tt, ok := t.(*testing.T)
	if ok {
		tt.Helper()
	}

	opts = append(opts, protocmp.Transform())

	if cmp.Equal(got, want, opts...) {
		return true
	}

	return assert.Fail(t, "Not equal, but expected to be equal", cmp.Diff(got, want, opts...))
}

// NotEqual is similar to [Equal], but inverse.
func NotEqual[T any](t TestingT, want T, got T, opts ...cmp.Option) bool {
	tt, ok := t.(*testing.T)
	if ok {
		tt.Helper()
	}

	opts = append(opts, protocmp.Transform())

	if !cmp.Equal(got, want, opts...) {
		return true
	}

	return assert.Fail(t, "Equal, but expected to be not equal", cmp.Diff(got, want, opts...))
}

func Nil[T any](t *testing.T, obj T, msgAndArgs ...any) bool {
	t.Helper()

	return assert.Nil(t, obj, msgAndArgs...)
}

func NotNil[T any](t *testing.T, obj T, msgAndArgs ...any) bool {
	t.Helper()

	return assert.NotNil(t, obj, msgAndArgs...)
}

func Empty[T any](t *testing.T, obj T, msgAndArgs ...any) bool {
	t.Helper()

	return assert.Empty(t, obj, msgAndArgs...)
}

// Is asserts that a certain incoming object a (of type [any]) is of type T. It will return a type casted variant of
// that object in the return value obj, if it succeeded.
func Is[T any](t TestingT, a any) (obj T) {
	var ok bool

	obj, ok = a.(T)
	assert.True(t, ok)

	return
}

// AnyValue is a [Want] that accepts any value of T.
func AnyValue[T any](*testing.T, T, ...any) bool {
	return true
}

// Optional asserts the [want] function, if it is not nil. Otherwise, the assertion is ignored. This is helpful if an extra [Want] func is specified only for a select sub-set of table tests.
func Optional[T any](t *testing.T, want Want[T], got T) bool {
	if want != nil {
		return want(t, got)
	}

	return true
}

// WantResponse is a helper to assert a [connect.Response] (including error) in tests.
func WantResponse[T any](t *testing.T, got *connect.Response[T], gotErr error, want Want[*T], wantErr WantErr) bool {
	t.Helper()

	// Assert error first, if we "want" an error
	if wantErr != nil {
		cErr := Is[*connect.Error](t, gotErr)
		if !wantErr(t, cErr) {
			return false
		}
	} else {
		if !Nil(t, gotErr) {
			return false
		}
	}

	if want == nil {
		return assert.Nil(t, got)
	} else {
		return want(t, got.Msg)
	}
}

// Getter is an interface for database operations that can retrieve objects by ID.
type Getter interface {
	Get(dest any, conds ...any) error
}

// InDB retrieves an object from the database by ID and returns it.
// If the object cannot be retrieved, the test fails and returns the zero value.
func InDB[T any](t *testing.T, db Getter, id string) *T {
	t.Helper()

	var obj T
	err := db.Get(&obj, "id = ?", id)
	if !NoError(t, err) {
		return nil
	}

	return &obj
}
