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

package assert

import (
	"errors"
	"testing"

	"buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go/buf/validate"
	"buf.build/go/protovalidate"
	"connectrpc.com/connect"
)

func TestIsConnectError(t *testing.T) {
	type args struct {
		err  error
		code connect.Code
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "correct connect error",
			args: args{
				err:  connect.NewError(connect.CodeInvalidArgument, errors.New("test error")),
				code: connect.CodeInvalidArgument,
			},
			want: true,
		},
		{
			name: "wrong code",
			args: args{
				err:  connect.NewError(connect.CodeNotFound, errors.New("test error")),
				code: connect.CodeInvalidArgument,
			},
			want: false,
		},
		{
			name: "not a connect error",
			args: args{
				err:  errors.New("regular error"),
				code: connect.CodeInvalidArgument,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsConnectError(&fakeT{}, tt.args.err, tt.args.code)
			if got != tt.want {
				t.Errorf("IsConnectError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidationError(t *testing.T) {
	type args struct {
		err   error
		field string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "validation error with correct field",
			args: args{
				err: connect.NewError(connect.CodeInvalidArgument, &protovalidate.ValidationError{
					Violations: []*protovalidate.Violation{
						{
							Proto: &validate.Violation{
								Field:   &validate.FieldPath{Elements: []*validate.FieldPathElement{{FieldName: stringPtr("id")}}},
								Message: stringPtr("field is required"),
							},
						},
					},
				}),
				field: "id",
			},
			want: true,
		},
		{
			name: "validation error with different field",
			args: args{
				err: connect.NewError(connect.CodeInvalidArgument, &protovalidate.ValidationError{
					Violations: []*protovalidate.Violation{
						{
							Proto: &validate.Violation{
								Field:   &validate.FieldPath{Elements: []*validate.FieldPathElement{{FieldName: stringPtr("name")}}},
								Message: stringPtr("field is required"),
							},
						},
					},
				}),
				field: "id",
			},
			want: false,
		},
		{
			name: "validation error with multiple fields",
			args: args{
				err: connect.NewError(connect.CodeInvalidArgument, &protovalidate.ValidationError{
					Violations: []*protovalidate.Violation{
						{
							Proto: &validate.Violation{
								Field:   &validate.FieldPath{Elements: []*validate.FieldPathElement{{FieldName: stringPtr("name")}}},
								Message: stringPtr("field is required"),
							},
						},
						{
							Proto: &validate.Violation{
								Field:   &validate.FieldPath{Elements: []*validate.FieldPathElement{{FieldName: stringPtr("id")}}},
								Message: stringPtr("field is required"),
							},
						},
					},
				}),
				field: "id",
			},
			want: true,
		},
		{
			name: "wrong error code",
			args: args{
				err: connect.NewError(connect.CodeNotFound, &protovalidate.ValidationError{
					Violations: []*protovalidate.Violation{
						{
							Proto: &validate.Violation{
								Field:   &validate.FieldPath{Elements: []*validate.FieldPathElement{{FieldName: stringPtr("id")}}},
								Message: stringPtr("not found"),
							},
						},
					},
				}),
				field: "id",
			},
			want: false,
		},
		{
			name: "not a connect error",
			args: args{
				err:   errors.New("regular error"),
				field: "id",
			},
			want: false,
		},
		{
			name: "connect error without validation error",
			args: args{
				err:   connect.NewError(connect.CodeInvalidArgument, errors.New("some other error")),
				field: "id",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidationError(&fakeT{}, tt.args.err, tt.args.field)
			if got != tt.want {
				t.Errorf("IsValidationError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
