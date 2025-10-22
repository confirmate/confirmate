// Copyright 2025 Fraunhofer AISEC:
// This code is licensed under the terms of the Apache License, Version 2.0.
// See the LICENSE file in this project for details.

package util

import "reflect"

// Deref dereferences pointer values
func Deref[T any](p *T) T {
	var result T
	if p != nil {
		return *p
	}

	return result
}

// Ref references pointer values
func Ref[T any](p T) *T {
	return &p
}

// IsNil checks if an interface value is nil or if the value nil is assigned to it.
func IsNil(value any) bool {
	if value == nil || (reflect.ValueOf(value).Kind() == reflect.Pointer &&
		reflect.ValueOf(value).IsNil()) {
		return true
	}

	return false
}
