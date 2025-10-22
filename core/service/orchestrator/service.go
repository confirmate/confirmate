// Copyright 2025 Fraunhofer AISEC:
// This code is licensed under the terms of the Apache License, Version 2.0.
// See the LICENSE file in this project for details.

package orchestrator

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/api/orchestrator/orchestratorconnect"
	"confirmate.io/core/db"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm/schema"
)

type service struct {
	orchestratorconnect.UnimplementedOrchestratorHandler
	storage *db.Storage
}

func NewService() (orchestratorconnect.OrchestratorHandler, error) {
	var (
		svc = &service{}
		err error
	)

	svc.storage, err = db.NewStorage(db.WithAutoMigration(types))
	if err != nil {
		return nil, fmt.Errorf("could not create storage: %w", err)
	}

	// Register custom serializers
	schema.RegisterSerializer("timestamppb", &TimestampSerializer{})

	// Setup Join Table

	if err = svc.storage.DB.SetupJoinTable(orchestrator.TargetOfEvaluation{}, "ConfiguredMetrics", assessment.MetricConfiguration{}); err != nil {
		return nil, fmt.Errorf("error during join-table: %w", err)
	}
	// Create table
	err = svc.storage.DB.AutoMigrate(
		orchestrator.TargetOfEvaluation{})
	if err != nil {
		return nil, fmt.Errorf("could not migrate TargetOfEvaluation: %w", err)
	}

	err = svc.storage.Create(&orchestrator.TargetOfEvaluation{
		Id:   "1",
		Name: "TOE1",
	})
	if err != nil {
		return nil, fmt.Errorf("could not create TOE: %w", err)
	}

	return svc, nil
}

func (svc *service) ListTargetsOfEvaluation(context.Context, *connect.Request[orchestrator.ListTargetsOfEvaluationRequest]) (*connect.Response[orchestrator.ListTargetsOfEvaluationResponse], error) {
	var (
		toes = []*orchestrator.TargetOfEvaluation{}
		err  error
	)

	err = svc.storage.List(&toes, "name", true, 0, -1, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to query targets of evaluation: %w", err)
	}

	return connect.NewResponse(&orchestrator.ListTargetsOfEvaluationResponse{
		TargetsOfEvaluation: toes,
	}), nil
}

// TimestampSerializer is a GORM serializer that allows the serialization and deserialization of the
// google.protobuf.Timestamp protobuf message type.
type TimestampSerializer struct{}

// Value implements https://pkg.go.dev/gorm.io/gorm/schema#SerializerValuerInterface to indicate
// how this struct will be saved into an SQL database field.
func (TimestampSerializer) Value(_ context.Context, _ *schema.Field, _ reflect.Value, fieldValue interface{}) (interface{}, error) {
	var (
		t  *timestamppb.Timestamp
		ok bool
	)

	if isNil(fieldValue) {
		return nil, nil
	}

	if t, ok = fieldValue.(*timestamppb.Timestamp); !ok {
		return nil, fmt.Errorf("unsupported type")
	}

	return t.AsTime(), nil
}

// Scan implements https://pkg.go.dev/gorm.io/gorm/schema#SerializerInterface to indicate how
// this struct can be loaded from an SQL database field.
func (TimestampSerializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	var t *timestamppb.Timestamp

	if dbValue != nil {
		switch v := dbValue.(type) {
		case time.Time:
			t = timestamppb.New(v)
		default:
			return fmt.Errorf("unsupported type")
		}

		field.ReflectValueOf(ctx, dst).Set(reflect.ValueOf(t))
	}

	return
}

// isNil checks if an interface value is nil or if the value nil is assigned to it.
// TODO(lebogg): Goes to util package, eventually
func isNil(value any) bool {
	if value == nil || (reflect.ValueOf(value).Kind() == reflect.Pointer &&
		reflect.ValueOf(value).IsNil()) {
		return true
	}

	return false
}
