package orchestrator

import (
	"context"
	"testing"
	"time"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/persistence"
	"confirmate.io/core/persistence/persistencetest"
	"confirmate.io/core/service/orchestrator/orchestratortest"
	"confirmate.io/core/util/assert"
	"connectrpc.com/connect"
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
		assert.Equal(t, orchestrator.ChangeEvent_TYPE_METRIC_CHANGE, event.Type)
		metricChange := event.GetMetricChange()
		assert.NotNil(t, metricChange)
		assert.Equal(t, orchestrator.MetricChangeEvent_TYPE_METADATA_CHANGED, metricChange.Type)
		assert.Equal(t, orchestratortest.MockMetric1.Id, metricChange.MetricId)
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

	// Register a subscriber with filter for METRIC_CHANGE only
	ch, id := svc.RegisterSubscriber(&orchestrator.SubscribeRequest_Filter{
		Types: []orchestrator.ChangeEvent_Type{orchestrator.ChangeEvent_TYPE_METRIC_CHANGE},
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
		assert.Equal(t, orchestrator.ChangeEvent_TYPE_METRIC_CHANGE, event.Type)
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for event")
	}
}
