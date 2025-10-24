// Copyright 2025 Fraunhofer AISEC:
// This code is licensed under the terms of the Apache License, Version 2.0.
// See the LICENSE file in this project for details.

package db

import (
	"context"
	"reflect"
	"testing"
	"time"

	"confirmate.io/core/util/testutil/assert"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm/schema"
)

func TestDurationSerializer_Value(t *testing.T) {
	type args struct {
		ctx        context.Context
		field      *schema.Field
		dst        reflect.Value
		fieldValue interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "ok field",
			args: args{
				field:      &schema.Field{Name: "timestamp"},
				dst:        reflect.Value{},
				fieldValue: durationpb.New(time.Duration(4)),
			},
			want:    time.Duration(4),
			wantErr: assert.NoError,
		},
		{
			name: "nil field",
			args: args{
				field:      &schema.Field{Name: "duration"},
				dst:        reflect.Value{},
				fieldValue: nil,
			},
			want:    nil,
			wantErr: assert.NoError,
		},
		{
			name: "field wrong type",
			args: args{
				field:      &schema.Field{Name: "duration"},
				dst:        reflect.Value{},
				fieldValue: "string",
			},
			want: nil,
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, ErrUnsupportedType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := DurationSerializer{}

			got, err := tr.Value(tt.args.ctx, tt.args.field, tt.args.dst, tt.args.fieldValue)
			tt.wantErr(t, err, tt.args)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDurationSerializer_Scan(t *testing.T) {
	type args struct {
		ctx     context.Context
		field   *schema.Field
		dst     reflect.Value
		dbValue interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "db wrong type",
			args: args{
				field:   &schema.Field{},
				dbValue: "string",
			},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, ErrUnsupportedType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := DurationSerializer{}
			err := tr.Scan(tt.args.ctx, tt.args.field, tt.args.dst, tt.args.dbValue)

			tt.wantErr(t, err, tt.args)
		})
	}
}

func TestTimestampSerializer_Value(t *testing.T) {
	type args struct {
		ctx        context.Context
		field      *schema.Field
		dst        reflect.Value
		fieldValue interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "ok field",
			args: args{
				field:      &schema.Field{Name: "timestamp"},
				dst:        reflect.Value{},
				fieldValue: timestamppb.New(time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC)),
			},
			want:    time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC),
			wantErr: assert.NoError,
		},
		{
			name: "nil field",
			args: args{
				field:      &schema.Field{Name: "timestamp"},
				dst:        reflect.Value{},
				fieldValue: nil,
			},
			want:    nil,
			wantErr: assert.NoError,
		},
		{
			name: "field wrong type",
			args: args{
				field:      &schema.Field{Name: "timestamp"},
				dst:        reflect.Value{},
				fieldValue: "string",
			},
			want: nil,
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, ErrUnsupportedType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := TimestampSerializer{}

			got, err := tr.Value(tt.args.ctx, tt.args.field, tt.args.dst, tt.args.fieldValue)
			tt.wantErr(t, err, tt.args)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTimestampSerializer_Scan(t *testing.T) {
	type args struct {
		ctx     context.Context
		field   *schema.Field
		dst     reflect.Value
		dbValue interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "db wrong type",
			args: args{
				field:   &schema.Field{},
				dbValue: "string",
			},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, ErrUnsupportedType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := TimestampSerializer{}
			err := tr.Scan(tt.args.ctx, tt.args.field, tt.args.dst, tt.args.dbValue)

			tt.wantErr(t, err, tt.args)
		})
	}
}

/* func TestAnySerializer_Value(t *testing.T) {
	type args struct {
		ctx        context.Context
		field      *schema.Field
		dst        reflect.Value
		fieldValue interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    assert.Want[any]
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "ok field",
			args: args{
				field: &schema.Field{Name: "config"},
				dst:   reflect.Value{},
				fieldValue: func() *anypb.Any {
					a, _ := anypb.New(&orchestrator.TargetOfEvaluation{
						Id: "my-target",
					})
					return a
				}(),
			},
			want: func(t *testing.T, got any) bool {
				// output of protojson is randomized (see
				// https://github.com/protocolbuffers/protobuf-go/commit/582ab3de426ef0758666e018b422dd20390f7f26),
				// so we need to unmarshal it to compare the contents in a
				// stable way
				b := assert.Is[[]byte](t, got)
				if !assert.NotNil(t, b) {
					return false
				}

				var m map[string]interface{}
				err := json.Unmarshal(b, &m)
				assert.NoError(t, err)

				return assert.Equal(t, m, map[string]interface{}{
					"@type": "type.googleapis.com/clouditor.orchestrator.v1.TargetOfEvaluation",
					"id":    "my-target",
				})
			},
			wantErr: assert.NoError,
		},
		{
			name: "nil field",
			args: args{
				field:      &schema.Field{Name: "config"},
				dst:        reflect.Value{},
				fieldValue: nil,
			},
			want:    assert.Nil[any],
			wantErr: assert.NoError,
		},
		{
			name: "field wrong type",
			args: args{
				field:      &schema.Field{Name: "config"},
				dst:        reflect.Value{},
				fieldValue: "string",
			},
			want: assert.Nil[any],
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorIs(t, err, ErrUnsupportedType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := AnySerializer{}

			got, err := tr.Value(tt.args.ctx, tt.args.field, tt.args.dst, tt.args.fieldValue)
			tt.wantErr(t, err, tt.args)
			tt.want(t, got)
		})
	}
} */

func TestAnySerializer_Scan(t *testing.T) {
	type args struct {
		ctx     context.Context
		field   *schema.Field
		dst     reflect.Value
		dbValue interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "db wrong type",
			args: args{
				field:   &schema.Field{},
				dbValue: "string",
			},
			wantErr: func(tt assert.TestingT, err error, i ...interface{}) bool {
				return assert.ErrorContains(t, err, "could not unmarshal JSONB value into protobuf message")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := AnySerializer{}
			err := tr.Scan(tt.args.ctx, tt.args.field, tt.args.dst, tt.args.dbValue)

			tt.wantErr(t, err, tt.args)

		})
	}
}
