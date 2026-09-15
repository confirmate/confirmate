// Copyright 2016-2026 Fraunhofer AISEC
//
// SPDX-License-Identifier: Apache-2.0

package pointer

// Deref returns the zero value when collector SDK pointers are nil.
func Deref[T any](pointer *T) T {
	var zero T
	if pointer == nil {
		return zero
	}

	return *pointer
}
