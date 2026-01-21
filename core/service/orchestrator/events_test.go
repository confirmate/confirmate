package orchestrator

import (
	"context"
	"testing"
	"time"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util"
	"confirmate.io/core/util/assert"
	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// types contains all Orchestrator types that we need to auto-migrate into database tables
var testTypes = []any{
	&orchestrator.TargetOfEvaluation{},
	&orchestrator.Certificate{},
	&orchestrator.State{},
	&orchestrator.Catalog{},
	&orchestrator.Category{},
	&orchestrator.Control{},
	&orchestrator.AuditScope{},
	&orchestrator.AssessmentTool{},
	&assessment.MetricConfiguration{},
	&assessment.AssessmentResult{},
	&assessment.Metric{},
	&assessment.MetricImplementation{},
	&assessment.AssessmentResult{},
}

var testJoinTables = []persistence.CustomJoinTable{
	{
		Model:     &orchestrator.TargetOfEvaluation{},
		Field:     "ConfiguredMetrics",
		JoinTable: &assessment.MetricConfiguration{},
	},
}

func TestService_RegisterSubscriber(t *testing.T) {
	// Initialize service with in-memory DB
	db := persistencetest.NewInMemoryDB(t, testTypes, testJoinTables)
	svc := &Service{
		db:          db,
		subscribers: make(map[int64]*subscriber),
	}

	// Register a subscriber
	ch, id := svc.RegisterSubscriber(nil)
	defer svc.UnregisterSubscriber(id)

	// Create a metric to trigger an event
	go func() {
		_, err := svc.CreateMetric(context.Background(), connect.NewRequest(&orchestrator.CreateMetricRequest{
			Metric: orchestratortest.MockMetric1,
		}))
		assert.NoError(t, err)
	}()

	// Wait for event
	select {
	case event := <-ch:
		assert.NotNil(t, event)
		assert.Equal(t, orchestrator.EventCategory_EVENT_CATEGORY_METRIC, event.Category)
		assert.Equal(t, orchestrator.ChangeType_CHANGE_TYPE_CREATED, event.ChangeType)
		assert.Equal(t, orchestratortest.MockMetric1.Id, event.EntityId)
		metric := event.GetMetric()
		assert.NotNil(t, metric)
		assert.Equal(t, orchestratortest.MockMetric1.Id, metric.Id)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestService_RegisterSubscriber_WithFilter(t *testing.T) {
	// Initialize service with in-memory DB
	db := persistencetest.NewInMemoryDB(t, testTypes, testJoinTables)
	svc := &Service{
		db:          db,
		subscribers: make(map[int64]*subscriber),
	}

	// Register a subscriber with filter for METRIC category only
	ch, id := svc.RegisterSubscriber(&orchestrator.SubscribeRequest_Filter{
		Categories: []orchestrator.EventCategory{orchestrator.EventCategory_EVENT_CATEGORY_METRIC},
	})
	defer svc.UnregisterSubscriber(id)

	// Create a metric (should be received)
	go func() {
		_, err := svc.CreateMetric(context.Background(), connect.NewRequest(&orchestrator.CreateMetricRequest{
			Metric: orchestratortest.MockMetric1,
		}))
		assert.NoError(t, err)
	}()

	// Wait for event
	select {
	case event := <-ch:
		assert.NotNil(t, event)
		assert.Equal(t, orchestrator.EventCategory_EVENT_CATEGORY_METRIC, event.Category)
		assert.NotNil(t, event.GetMetric())
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}

func TestValidateMessage_ChangeEvent(t *testing.T) {
	tests := []struct {
		name    string
		event   *orchestrator.ChangeEvent
		wantErr bool
	}{
		{
			name: "valid metric change",
			event: &orchestrator.ChangeEvent{
				Timestamp:            timestamppb.Now(),
				Category:             orchestrator.EventCategory_EVENT_CATEGORY_METRIC,
				ChangeType:           orchestrator.ChangeType_CHANGE_TYPE_CREATED,
				EntityId:             "metric-1",
				TargetOfEvaluationId: util.Ref("11111111-1111-1111-1111-111111111111"),
			},
			wantErr: false,
		},
		{
			name: "missing required fields",
			event: &orchestrator.ChangeEvent{
				Timestamp: timestamppb.Now(),
				// Missing category, change_type, entity_id - should fail validation
			},
			wantErr: true,
		},
		{
			name: "invalid entity id",
			event: &orchestrator.ChangeEvent{
				Timestamp:  timestamppb.Now(),
				Category:   orchestrator.EventCategory_EVENT_CATEGORY_METRIC,
				ChangeType: orchestrator.ChangeType_CHANGE_TYPE_UPDATED,
				EntityId:   "", // Empty entity ID should fail validation
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateEvent(tt.event)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
