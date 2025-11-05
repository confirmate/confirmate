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

package persistence

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm/schema"

	"confirmate.io/core/util"
)

// DurationSerializer is a GORM serializer that allows the serialization and deserialization of the
// google.protobuf.Duration protobuf message type.
type DurationSerializer struct{}

// Value implements https://pkg.go.dev/gorm.io/gorm/schema#SerializerValuerInterface to indicate how
// this struct will be saved into an SQL database field.
func (DurationSerializer) Value(_ context.Context, _ *schema.Field, _ reflect.Value, fieldValue interface{}) (interface{}, error) {
	var (
		t  *durationpb.Duration
		ok bool
	)

	if util.IsNil(fieldValue) {
		return nil, nil
	}

	if t, ok = fieldValue.(*durationpb.Duration); !ok {
		return nil, ErrUnsupportedType
	}

	return t.AsDuration(), nil
}

// Scan implements https://pkg.go.dev/gorm.io/gorm/schema#SerializerInterface to indicate how this
// struct can be loaded from an SQL database field.
func (DurationSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	var t *durationpb.Duration

	if dbValue != nil {
		switch v := dbValue.(type) {
		case time.Duration:
			t = durationpb.New(v)
		default:
			return ErrUnsupportedType
		}

		field.ReflectValueOf(ctx, dst).Set(reflect.ValueOf(t))
	}

	return
}

// TimestampSerializer is a GORM serializer that allows the serialization and deserialization of the
// google.protobuf.Timestamp protobuf message type.
type TimestampSerializer struct{}

// Value implements https://pkg.go.dev/gorm.io/gorm/schema#SerializerValuerInterface to indicate how
// this struct will be saved into an SQL database field.
func (TimestampSerializer) Value(_ context.Context, _ *schema.Field, _ reflect.Value, fieldValue interface{}) (interface{}, error) {
	var (
		t  *timestamppb.Timestamp
		ok bool
	)

	if util.IsNil(fieldValue) {
		return nil, nil
	}

	if t, ok = fieldValue.(*timestamppb.Timestamp); !ok {
		return nil, ErrUnsupportedType
	}

	return t.AsTime(), nil
}

// Scan implements https://pkg.go.dev/gorm.io/gorm/schema#SerializerInterface to indicate how this
// struct can be loaded from an SQL database field.
func (TimestampSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	var t *timestamppb.Timestamp

	if dbValue != nil {
		switch v := dbValue.(type) {
		case time.Time:
			t = timestamppb.New(v)
		default:
			return ErrUnsupportedType
		}

		field.ReflectValueOf(ctx, dst).Set(reflect.ValueOf(t))
	}

	return
}

// AnySerializer is a GORM serializer that allows the serialization and deserialization of the
// google.protobuf.Any protobuf message type using a JSONB field.
type AnySerializer struct{}

// Value implements https://pkg.go.dev/gorm.io/gorm/schema#SerializerValuerInterface to indicate how
// this struct will be saved into an SQL database field.
func (AnySerializer) Value(_ context.Context, _ *schema.Field, _ reflect.Value, fieldValue interface{}) (interface{}, error) {
	var (
		a  *anypb.Any
		ok bool
	)

	if util.IsNil(fieldValue) {
		return nil, nil
	}

	if a, ok = fieldValue.(*anypb.Any); !ok {
		return nil, ErrUnsupportedType
	}

	return protojson.Marshal(a)
}

// Scan implements https://pkg.go.dev/gorm.io/gorm/schema#SerializerInterface to indicate how this
// struct can be loaded from an SQL database field.
func (AnySerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	var (
		a anypb.Any
	)

	if dbValue != nil {
		var bytes []byte
		switch v := dbValue.(type) {
		case []byte:
			bytes = v
		case string:
			bytes = []byte(v)
		default:
			return ErrUnsupportedType
		}

		err = protojson.Unmarshal(bytes, &a)
		if err != nil {
			return fmt.Errorf("could not unmarshal JSONB value into protobuf message: %w", err)
		}
	}

	field.ReflectValueOf(ctx, dst).Set(reflect.ValueOf(&a))
	return
}

// ValueSerializer is a GORM serializer that allows the serialization and deserialization of the
// google.protobuf.Value protobuf message type.
type ValueSerializer struct{}

// Value implements https://pkg.go.dev/gorm.io/gorm/schema#SerializerValuerInterface to indicate how
// this struct will be saved into an SQL database field.
func (ValueSerializer) Value(_ context.Context, _ *schema.Field, _ reflect.Value, fieldValue interface{}) (interface{}, error) {
	var (
		v  *structpb.Value
		ok bool
	)

	if util.IsNil(fieldValue) {
		return nil, nil
	}

	if v, ok = fieldValue.(*structpb.Value); !ok {
		return nil, ErrUnsupportedType
	}

	return v.MarshalJSON()
}

// Scan implements https://pkg.go.dev/gorm.io/gorm/schema#SerializerInterface to indicate how this
// struct can be loaded from an SQL database field.
func (ValueSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	v := new(structpb.Value)

	if dbValue != nil {
		switch d := dbValue.(type) {
		case []byte:
			err = v.UnmarshalJSON(d)
			if err != nil {
				return err
			}
		default:
			return ErrUnsupportedType
		}

		field.ReflectValueOf(ctx, dst).Set(reflect.ValueOf(v))
	}

	return
}
