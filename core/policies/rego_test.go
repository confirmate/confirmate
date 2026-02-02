// Copyright 2016-2026 Fraunhofer AISEC
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

package policies

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/api/orchestrator"
	"confirmate.io/core/util/assert"
	"confirmate.io/core/util/prototest"
	"confirmate.io/core/util/testdata"
	"github.com/open-policy-agent/opa/v1/rego"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_regoEval_Eval(t *testing.T) {

	type fields struct {
		qc   *queryCache
		mrtc *metricsCache
		pkg  string
	}
	type args struct {
		resource   ontology.IsResource
		related    map[string]ontology.IsResource
		evidenceID string
		src        MetricsSource
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		compliant map[string]bool
		wantErr   assert.WantErr
	}{
		{
			name: "ObjectStorage: Compliant Case",
			fields: fields{
				qc:   newQueryCache(),
				mrtc: &metricsCache{m: make(map[string][]*assessment.Metric)},
				pkg:  DefaultRegoPackage,
			},
			compliant: map[string]bool{
				"AtRestEncryptionAlgorithm":         true,
				"AtRestEncryptionEnabled":           true,
				"ObjectStoragePublicAccessDisabled": true,
			},
			args: args{
				resource: &ontology.ObjectStorage{
					Id:           mockObjStorage1ResourceID,
					CreationTime: timestamppb.New(time.Unix(1621086669, 0)),
					AtRestEncryption: &ontology.AtRestEncryption{
						Type: &ontology.AtRestEncryption_CustomerKeyEncryption{
							CustomerKeyEncryption: &ontology.CustomerKeyEncryption{
								Algorithm: "AES256",
								Enabled:   true,
								KeyUrl:    "SomeUrl",
							},
						},
					},
					PublicAccess: false,
				},
				evidenceID: mockObjStorage1EvidenceID,
				src:        &mockMetricsSource{t: t},
			},
			wantErr: assert.NoError,
		},
		{
			name: "ObjectStorage: Non-Compliant Case with no Encryption at rest",
			fields: fields{
				qc:   newQueryCache(),
				mrtc: &metricsCache{m: make(map[string][]*assessment.Metric)},
				pkg:  DefaultRegoPackage,
			},
			args: args{
				resource: &ontology.ObjectStorage{
					Id:           mockObjStorage1ResourceID,
					CreationTime: timestamppb.New(time.Unix(1621086669, 0)),
					AtRestEncryption: &ontology.AtRestEncryption{
						Type: &ontology.AtRestEncryption_CustomerKeyEncryption{
							CustomerKeyEncryption: &ontology.CustomerKeyEncryption{
								Algorithm: "NoGoodAlg",
								Enabled:   false,
							},
						},
					},
					PublicAccess: true,
				},
				evidenceID: mockObjStorage2EvidenceID,
				src:        &mockMetricsSource{t: t},
			},
			compliant: map[string]bool{
				"AtRestEncryptionAlgorithm":         false,
				"AtRestEncryptionEnabled":           false,
				"ObjectStoragePublicAccessDisabled": false,
			},
			wantErr: assert.NoError,
		},
		{
			name: "ObjectStorage: Non-Compliant Case 2 with no customer managed key",
			fields: fields{
				qc:   newQueryCache(),
				mrtc: &metricsCache{m: make(map[string][]*assessment.Metric)},
				pkg:  DefaultRegoPackage,
			},
			args: args{
				resource: &ontology.ObjectStorage{
					Id:           mockObjStorage1ResourceID,
					CreationTime: timestamppb.New(time.Unix(1621086669, 0)),
					AtRestEncryption: &ontology.AtRestEncryption{
						Type: &ontology.AtRestEncryption_CustomerKeyEncryption{
							CustomerKeyEncryption: &ontology.CustomerKeyEncryption{
								// Normally given but for test case purpose only check that no key URL is given
								Algorithm: "",
								Enabled:   false,
							},
						},
					},
					PublicAccess: true,
				},
				evidenceID: mockObjStorage2EvidenceID,
				src:        &mockMetricsSource{t: t},
			},
			compliant: map[string]bool{
				"AtRestEncryptionAlgorithm":         false,
				"AtRestEncryptionEnabled":           false,
				"ObjectStoragePublicAccessDisabled": false,
			},
			wantErr: assert.NoError,
		},
		{
			name: "VM: Compliant Case",
			fields: fields{
				qc:   newQueryCache(),
				mrtc: &metricsCache{m: make(map[string][]*assessment.Metric)},
				pkg:  DefaultRegoPackage,
			},
			args: args{
				src: &mockMetricsSource{t: t},
				resource: &ontology.VirtualMachine{
					Id: mockVM1ResourceID,
					AutomaticUpdates: &ontology.AutomaticUpdates{
						Enabled:      true,
						Interval:     durationpb.New(time.Hour * 24 * 30),
						SecurityOnly: true,
					},
					BootLogging: &ontology.BootLogging{
						LoggingServiceIds: []string{"SomeResourceId1", "SomeResourceId2"},
						Enabled:           true,
						RetentionPeriod:   durationpb.New(36 * time.Hour * 24),
					},
					OsLogging: &ontology.OSLogging{
						LoggingServiceIds: []string{"SomeResourceId2"},
						Enabled:           true,
						RetentionPeriod:   durationpb.New(36 * time.Hour * 24),
					},
					MalwareProtection: &ontology.MalwareProtection{
						Enabled:              true,
						DurationSinceActive:  durationpb.New(time.Hour * 24 * 5),
						NumberOfThreatsFound: 5,
						ApplicationLogging: &ontology.ApplicationLogging{
							Enabled:           true,
							RetentionPeriod:   durationpb.New(time.Hour * 24 * 36),
							LoggingServiceIds: []string{"SomeAnalyticsService?"},
						},
					},
				},
				evidenceID: mockVM1EvidenceID,
			},
			compliant: map[string]bool{
				"AutomaticUpdatesEnabled":  true,
				"AutomaticUpdatesInterval": true,
				"BootLoggingEnabled":       true,
				"BootLoggingOutput":        true,
				"BootLoggingRetention":     true,
				"MalwareProtectionEnabled": true,
				"MalwareProtectionOutput":  true,
				"OSLoggingRetention":       true,
				"OSLoggingOutput":          true,
				"OSLoggingEnabled":         true,
			},
			wantErr: assert.NoError,
		},
		{
			name: "VM: Non-Compliant Case",
			fields: fields{
				qc:   newQueryCache(),
				mrtc: &metricsCache{m: make(map[string][]*assessment.Metric)},
				pkg:  DefaultRegoPackage,
			},
			args: args{
				resource: &ontology.VirtualMachine{
					Id: mockVM2ResourceID,
					BootLogging: &ontology.BootLogging{
						LoggingServiceIds: nil,
						Enabled:           false,
						RetentionPeriod:   durationpb.New(1 * time.Hour * 24),
					},
					OsLogging: &ontology.OSLogging{
						LoggingServiceIds: []string{"SomeResourceId3"},
						Enabled:           false,
						RetentionPeriod:   durationpb.New(1 * time.Hour * 24),
					},
				},
				evidenceID: mockVM2EvidenceID,
				src:        &mockMetricsSource{t: t},
			},
			compliant: map[string]bool{
				"AutomaticUpdatesEnabled":  false,
				"AutomaticUpdatesInterval": false,
				"BootLoggingEnabled":       false,
				"BootLoggingOutput":        false,
				"BootLoggingRetention":     false,
				"MalwareProtectionEnabled": false,
				"OSLoggingEnabled":         false,
				"OSLoggingOutput":          true,
				"OSLoggingRetention":       false,
			},
			wantErr: assert.NoError,
		},
		{
			name: "VM: Related Evidence: non-compliant VMDiskEncryptionEnabled",
			fields: fields{
				qc:   newQueryCache(),
				mrtc: &metricsCache{m: make(map[string][]*assessment.Metric)},
				pkg:  DefaultRegoPackage,
			},
			args: args{
				resource: &ontology.VirtualMachine{
					Id:              mockVM2ResourceID,
					BlockStorageIds: []string{mockBlockStorage1ID},
				},
				evidenceID: mockVM1EvidenceID,
				src:        &mockMetricsSource{t: t},
				related: map[string]ontology.IsResource{
					mockBlockStorage1ID: &ontology.BlockStorage{
						Id: mockBlockStorage1ID,
						AtRestEncryption: &ontology.AtRestEncryption{
							Type: &ontology.AtRestEncryption_CustomerKeyEncryption{
								CustomerKeyEncryption: &ontology.CustomerKeyEncryption{
									Enabled:   false,
									Algorithm: "AES256",
								},
							},
						},
					},
				},
			},
			compliant: map[string]bool{
				"AutomaticUpdatesEnabled":             false,
				"AutomaticUpdatesInterval":            false,
				"BootLoggingEnabled":                  false,
				"BootLoggingOutput":                   false,
				"BootLoggingRetention":                false,
				"MalwareProtectionEnabled":            false,
				"OSLoggingEnabled":                    false,
				"OSLoggingOutput":                     false,
				"OSLoggingRetention":                  false,
				"VirtualMachineDiskEncryptionEnabled": false,
			},
			wantErr: assert.NoError,
		},
		{
			name: "VM: Related Evidence",
			fields: fields{
				qc:   newQueryCache(),
				mrtc: &metricsCache{m: make(map[string][]*assessment.Metric)},
				pkg:  DefaultRegoPackage,
			},
			args: args{
				resource: &ontology.VirtualMachine{
					Id:              mockVM2ResourceID,
					BlockStorageIds: []string{mockBlockStorage1ID},
				},
				evidenceID: mockVM1EvidenceID,
				src:        &mockMetricsSource{t: t},
				related: map[string]ontology.IsResource{
					mockBlockStorage1ID: &ontology.BlockStorage{
						Id: mockBlockStorage1ID,
						AtRestEncryption: &ontology.AtRestEncryption{
							Type: &ontology.AtRestEncryption_CustomerKeyEncryption{
								CustomerKeyEncryption: &ontology.CustomerKeyEncryption{
									Enabled:   true,
									Algorithm: "AES256",
								},
							},
						},
					},
				},
			},
			compliant: map[string]bool{
				"AutomaticUpdatesEnabled":             false,
				"AutomaticUpdatesInterval":            false,
				"BootLoggingEnabled":                  false,
				"BootLoggingOutput":                   false,
				"BootLoggingRetention":                false,
				"MalwareProtectionEnabled":            false,
				"OSLoggingEnabled":                    false,
				"OSLoggingOutput":                     false,
				"OSLoggingRetention":                  false,
				"VirtualMachineDiskEncryptionEnabled": true,
			},
			wantErr: assert.NoError,
		},
		{
			name: "Application: StrongCryptographicHash",
			fields: fields{
				qc:   newQueryCache(),
				mrtc: &metricsCache{m: make(map[string][]*assessment.Metric)},
				pkg:  DefaultRegoPackage,
			},
			args: args{
				resource: &ontology.Application{
					Id: "app",
					// TODO: why is the hash metric evaluated to true without any functionalities?
					// Functionalities: []*ontology.Functionality{
					// 	{
					// 		Type: &ontology.Functionality_CipherSuite{
					// 			CipherSuite: &ontology.CipherSuite{
					// 				SessionCipher: "AES",
					// 			},
					// 		},
					// 	},
					// },
				},
				evidenceID: mockVM1EvidenceID,
				src:        &mockMetricsSource{t: t},
			},
			compliant: map[string]bool{
				"AutomaticUpdatesEnabled":  false,
				"AutomaticUpdatesInterval": false,
				"StrongCryptographicHash":  true,
			},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := regoEval{
				qc:   tt.fields.qc,
				mrtc: tt.fields.mrtc,
				pkg:  tt.fields.pkg,
			}
			results, err := pe.Eval(&evidence.Evidence{
				Id:       tt.args.evidenceID,
				Resource: prototest.NewProtobufResource(t, tt.args.resource),
			}, tt.args.resource, tt.args.related, tt.args.src)

			tt.wantErr(t, err)

			assert.NotEmpty(t, results)

			var compliants = map[string]bool{}

			for _, result := range results {
				if result.Applicable {
					compliants[result.MetricName] = result.Compliant
				}
			}

			assert.Equal(t, tt.compliant, compliants)
		})
	}
}

func Test_regoEval_evalMap(t *testing.T) {
	type fields struct {
		qc   *queryCache
		mrtc *metricsCache
		pkg  string
	}
	type args struct {
		baseDir  string
		targetID string
		metric   *assessment.Metric
		m        map[string]interface{}
		src      MetricsSource
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    assert.Want[*CombinedResult]
		wantErr assert.WantErr
	}{
		{
			name: "default metric configuration",
			fields: fields{
				qc:   newQueryCache(),
				mrtc: &metricsCache{m: make(map[string][]*assessment.Metric)},
				pkg:  DefaultRegoPackage,
			},
			args: args{
				targetID: testdata.MockTargetOfEvaluationID1,
				metric: &assessment.Metric{
					Id:       "84eaed86-759d-4419-9954-f3d3ea1f5200",
					Name:     "AutomaticUpdatesEnabled",
					Category: "EndpointSecurity",
					Version:  "v1",
					Comments: "Test comments",
				},
				baseDir: ".",
				m: map[string]interface{}{
					"automaticUpdates": map[string]interface{}{
						"enabled": true,
					},
				},
				src: &mockMetricsSource{t: t},
			},
			want: func(t *testing.T, got *CombinedResult, args ...any) bool {
				want := &CombinedResult{
					Applicable: true,
					Compliant:  true,
					MetricID:   "84eaed86-759d-4419-9954-f3d3ea1f5200",
					MetricName: "AutomaticUpdatesEnabled",
					Config: &assessment.MetricConfiguration{
						Operator:             "==",
						TargetValue:          structpb.NewBoolValue(true),
						IsDefault:            true,
						UpdatedAt:            nil,
						MetricId:             "84eaed86-759d-4419-9954-f3d3ea1f5200",
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
					},
					Message: assessment.DefaultCompliantMessage,
				}

				return assert.Equal(t, want, got)
			},
			wantErr: assert.NoError,
		},
		{
			name: "updated metric configuration",
			fields: fields{
				qc:   newQueryCache(),
				mrtc: &metricsCache{m: make(map[string][]*assessment.Metric)},
				pkg:  DefaultRegoPackage,
			},
			args: args{
				targetID: testdata.MockTargetOfEvaluationID1,
				metric: &assessment.Metric{
					Id:       "84eaed86-759d-4419-9954-f3d3ea1f5200",
					Name:     "AutomaticUpdatesEnabled",
					Category: "EndpointSecurity",
				},
				baseDir: ".",
				m: map[string]interface{}{
					"automaticUpdates": map[string]interface{}{
						"enabled": true,
					},
				},
				src: &updatedMockMetricsSource{mockMetricsSource{t: t}},
			},
			want: func(t *testing.T, got *CombinedResult, args ...any) bool {
				want := &CombinedResult{
					Applicable: true,
					Compliant:  false,
					MetricID:   "84eaed86-759d-4419-9954-f3d3ea1f5200",
					MetricName: "AutomaticUpdatesEnabled",
					Config: &assessment.MetricConfiguration{
						Operator:             "==",
						TargetValue:          structpb.NewBoolValue(false),
						IsDefault:            false,
						UpdatedAt:            timestamppb.New(time.Date(2022, 12, 1, 0, 0, 0, 0, time.Local)),
						MetricId:             "84eaed86-759d-4419-9954-f3d3ea1f5200",
						TargetOfEvaluationId: testdata.MockTargetOfEvaluationID1,
					},
					Message: assessment.DefaultNonCompliantMessage,
				}

				return assert.Equal(t, want, got)
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			re := &regoEval{
				qc:   tt.fields.qc,
				mrtc: tt.fields.mrtc,
				pkg:  tt.fields.pkg,
			}
			gotResult, err := re.evalMap(tt.args.baseDir, tt.args.targetID, tt.args.metric, tt.args.m, tt.args.src)

			tt.wantErr(t, err)
			tt.want(t, gotResult)
		})
	}
}

func Test_reencode(t *testing.T) {
	type args struct {
		in  any
		out any
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.WantErr
	}{
		{
			name: "Happy path: Map to Map",
			args: args{
				in:  map[string]string{"key": "value"},
				out: new(map[string]string),
			},
			wantErr: assert.NoError,
		},
		{
			name: "Happy path: Slice to Slice",
			args: args{
				in:  []int{1, 2, 3},
				out: new([]int),
			},
			wantErr: assert.NoError,
		},
		{
			name: "Happy path: String to String",
			args: args{
				in:  "hello",
				out: new(string),
			},
			wantErr: assert.NoError,
		},
		{
			name: "Invalid Input (Marshal)",
			args: args{
				in:  make(chan int),
				out: new(map[string]string),
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "JSON marshal failed")
			},
		},
		{
			name: "Invalid Input (Unmarshal)",
			args: args{
				in:  42, // Invalid input for unmarshalling into a map[string]string
				out: new(map[string]string),
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "JSON unmarshal failed")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			switch out := tt.args.out.(type) {
			case *map[string]string:
				err = reencode(tt.args.in, out)
			case *([]int):
				err = reencode(tt.args.in, out)
			case *string:
				err = reencode(tt.args.in, out)
			default:
				t.Fatalf("unsupported type: %T", out)
			}

			tt.wantErr(t, err)
		})
	}
}

// Test_regoEval_HandleMetricEvent tests cache eviction on metric events
func Test_regoEval_HandleMetricEvent_EvictOnImplementationChange(t *testing.T) {
	re := &regoEval{
		qc:   newQueryCache(),
		mrtc: &metricsCache{m: make(map[string][]*assessment.Metric)},
		pkg:  DefaultRegoPackage,
	}

	// Add a query to cache (cast to interface{} then back)
	re.qc.cache["metric-123-target1"] = &rego.PreparedEvalQuery{}

	// Verify cache has entry
	assert.Equal(t, 1, len(re.qc.cache))

	// Handle metric implementation change event
	event := &orchestrator.ChangeEvent{
		Category: orchestrator.EventCategory_EVENT_CATEGORY_METRIC_IMPLEMENTATION,
		EntityId: "metric-123",
	}

	err := re.HandleMetricEvent(event)
	assert.NoError(t, err)

	// Verify cache entry is evicted
	assert.Equal(t, 0, len(re.qc.cache))
}

// Test_regoEval_HandleMetricEvent_EvictOnConfigurationChange tests eviction for configuration changes
func Test_regoEval_HandleMetricEvent_EvictOnConfigurationChange(t *testing.T) {
	re := &regoEval{
		qc:   newQueryCache(),
		mrtc: &metricsCache{m: make(map[string][]*assessment.Metric)},
		pkg:  DefaultRegoPackage,
	}

	// Add multiple queries to cache for the same metric
	re.qc.cache["metric-456-target1"] = &rego.PreparedEvalQuery{}
	re.qc.cache["metric-456-target2"] = &rego.PreparedEvalQuery{}
	re.qc.cache["metric-789-target1"] = &rego.PreparedEvalQuery{}

	// Verify cache has 3 entries
	assert.Equal(t, 3, len(re.qc.cache))

	// Handle metric configuration change event for metric-456
	event := &orchestrator.ChangeEvent{
		Category: orchestrator.EventCategory_EVENT_CATEGORY_METRIC_CONFIGURATION,
		EntityId: "metric-456",
	}

	err := re.HandleMetricEvent(event)
	assert.NoError(t, err)

	// Verify only metric-456 entries are evicted
	assert.Equal(t, 1, len(re.qc.cache))
	_, exists := re.qc.cache["metric-789-target1"]
	assert.True(t, exists, "metric-789-target1 should still exist")
}

// Test_regoEval_HandleMetricEvent_PartialKeyMatching verifies prefix-based eviction
func Test_regoEval_HandleMetricEvent_PartialKeyMatching(t *testing.T) {
	re := &regoEval{
		qc:   newQueryCache(),
		mrtc: &metricsCache{m: make(map[string][]*assessment.Metric)},
		pkg:  DefaultRegoPackage,
	}

	// Add cache entries with similar keys
	// Note: Evict uses HasPrefix, so "metric-123" will match "metric-123*" AND "metric-1234*"
	re.qc.cache["metric-123-config1"] = &rego.PreparedEvalQuery{}
	re.qc.cache["metric-123-config2"] = &rego.PreparedEvalQuery{}
	re.qc.cache["metric-456-config1"] = &rego.PreparedEvalQuery{}

	// Evict metric-123
	event := &orchestrator.ChangeEvent{
		Category: orchestrator.EventCategory_EVENT_CATEGORY_METRIC_IMPLEMENTATION,
		EntityId: "metric-123",
	}
	err := re.HandleMetricEvent(event)
	assert.NoError(t, err)

	// Verify only metric-123 entries are evicted
	assert.Equal(t, 1, len(re.qc.cache))
	_, exists := re.qc.cache["metric-456-config1"]
	assert.True(t, exists, "metric-456-config1 should still exist")
}

// Test_queryCache_GetExecutesOrElseOnMiss tests cache hit/miss behavior
func Test_queryCache_GetExecutesOrElseOnMiss(t *testing.T) {
	qc := newQueryCache()

	callCount := 0
	orElseFn := func(key string) (*rego.PreparedEvalQuery, error) {
		callCount++
		return &rego.PreparedEvalQuery{}, nil
	}

	// First call should execute orElseFn
	result1, err := qc.Get("key1", orElseFn)
	assert.NoError(t, err)
	assert.NotNil(t, result1)
	assert.Equal(t, 1, callCount)

	// Second call should NOT execute orElseFn (cache hit)
	orElseFn2 := func(key string) (*rego.PreparedEvalQuery, error) {
		t.Fatal("orElseFn should not be called on cache hit")
		return nil, nil
	}
	result2, err := qc.Get("key1", orElseFn2)
	assert.NoError(t, err)
	assert.NotNil(t, result2)
	assert.Equal(t, 1, callCount) // Still 1, not incremented
}

// Test_queryCache_EvictRemovesMatchingPrefixes tests prefix-based eviction
func Test_queryCache_EvictRemovesMatchingPrefixes(t *testing.T) {
	qc := newQueryCache()

	// Add cache entries
	qc.cache["metric-123-config1"] = &rego.PreparedEvalQuery{}
	qc.cache["metric-123-config2"] = &rego.PreparedEvalQuery{}
	qc.cache["metric-456-config1"] = &rego.PreparedEvalQuery{}

	// Evict entries for metric-123
	qc.Evict("metric-123")

	// Verify correct entries removed
	assert.Equal(t, 1, len(qc.cache))
	_, exists := qc.cache["metric-456-config1"]
	assert.True(t, exists, "metric-456-config1 should remain")
}

// Test_queryCache_EmptyClears verifies Empty() clears all cache
func Test_queryCache_EmptyClears(t *testing.T) {
	qc := newQueryCache()

	// Add multiple entries
	qc.cache["key1"] = &rego.PreparedEvalQuery{}
	qc.cache["key2"] = &rego.PreparedEvalQuery{}
	qc.cache["key3"] = &rego.PreparedEvalQuery{}
	assert.Equal(t, 3, len(qc.cache))

	// Clear cache
	qc.Empty()

	// Verify completely empty
	assert.Equal(t, 0, len(qc.cache))
}

// Test_queryCache_ConcurrentAccess tests thread-safety with concurrent operations
func Test_queryCache_ConcurrentAccess(t *testing.T) {
	qc := newQueryCache()

	// Add initial entries
	qc.cache["initial-1"] = &rego.PreparedEvalQuery{}
	qc.cache["initial-2"] = &rego.PreparedEvalQuery{}

	done := make(chan bool)
	errors := make(chan error, 10)

	// Concurrent Get operations
	for i := 0; i < 3; i++ {
		go func(id int) {
			key := fmt.Sprintf("key-%d", id)
			counter := 0
			_, err := qc.Get(key, func(k string) (*rego.PreparedEvalQuery, error) {
				counter++
				if counter > 1 {
					errors <- fmt.Errorf("orElseFn called multiple times for key %s", k)
				}
				return &rego.PreparedEvalQuery{}, nil
			})
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Concurrent Evict operations
	for i := 0; i < 2; i++ {
		go func(prefix string) {
			qc.Evict(prefix)
			done <- true
		}(fmt.Sprintf("prefix-%d", i))
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Check for errors
	select {
	case err := <-errors:
		t.Fatalf("Concurrent access error: %v", err)
	default:
		// No errors
	}

	// Verify cache is still in valid state
	assert.NotNil(t, qc.cache)
}

// Test_queryCache_GetHandlesOrElseError tests error handling
func Test_queryCache_GetHandlesOrElseError(t *testing.T) {
	qc := newQueryCache()

	testError := fmt.Errorf("test error from orElse")
	orElseFn := func(key string) (*rego.PreparedEvalQuery, error) {
		return nil, testError
	}

	// First call returns error
	result, err := qc.Get("key1", orElseFn)
	assert.Nil(t, result)
	assert.ErrorContains(t, err, "test error from orElse")

	// Verify key is NOT cached after error
	assert.Equal(t, 0, len(qc.cache))

	// Second call should re-execute orElseFn (not cached)
	callCount := 0
	orElseFn2 := func(key string) (*rego.PreparedEvalQuery, error) {
		callCount++
		return nil, testError
	}

	result2, err2 := qc.Get("key1", orElseFn2)
	assert.Nil(t, result2)
	assert.ErrorContains(t, err2, "test error from orElse")
	assert.Equal(t, 1, callCount)
}

// Test_NewRegoEval_WithoutEventPublisher tests regoEval creation without event subscription
func Test_NewRegoEval_WithoutEventSubscriber(t *testing.T) {
	re := NewRegoEval()
	assert.NotNil(t, re)

	// Should be able to use it without events
	regoEvalInstance := re.(*regoEval)
	assert.Nil(t, regoEvalInstance.eventSubscriber)
	assert.Equal(t, int64(-1), regoEvalInstance.subscriberID)
}

// Test_NewRegoEval_WithEventSubscriber tests regoEval creation with event subscription
func Test_NewRegoEval_WithEventSubscriber(t *testing.T) {
	mockSub := &mockEventSubscriber{
		subscribers: make(map[int64]*mockSubscriber),
	}

	re := NewRegoEval(WithEventSubscriber(mockSub))
	assert.NotNil(t, re)

	regoEvalInstance := re.(*regoEval)
	assert.NotNil(t, regoEvalInstance.eventSubscriber)

	// Give subscription goroutine time to start
	time.Sleep(50 * time.Millisecond)

	// Should have registered as a subscriber
	assert.True(t, mockSub.subscriberCount() > 0)

	// Cleanup
	regoEvalInstance.Close()
	time.Sleep(50 * time.Millisecond)
}

// Test_regoEval_EventSubscriptionReceivesMetricEvents tests event handling
func Test_regoEval_EventSubscriptionReceivesMetricEvents(t *testing.T) {
	mockSub := &mockEventSubscriber{
		subscribers: make(map[int64]*mockSubscriber),
	}

	re := NewRegoEval(WithEventSubscriber(mockSub))
	regoEvalInstance := re.(*regoEval)

	// Wait for subscription to start
	time.Sleep(50 * time.Millisecond)

	// Add some cache entries
	regoEvalInstance.qc.cache["metric-123-config1"] = &rego.PreparedEvalQuery{}
	regoEvalInstance.qc.cache["metric-456-config1"] = &rego.PreparedEvalQuery{}
	assert.Equal(t, 2, len(regoEvalInstance.qc.cache))

	// Publish an implementation change event
	event := &orchestrator.ChangeEvent{
		Category: orchestrator.EventCategory_EVENT_CATEGORY_METRIC_IMPLEMENTATION,
		EntityId: "metric-123",
	}
	mockSub.PublishEvent(event)

	// Give event processing time
	time.Sleep(100 * time.Millisecond)

	// Verify cache was evicted for metric-123
	assert.Equal(t, 1, len(regoEvalInstance.qc.cache))
	_, exists := regoEvalInstance.qc.cache["metric-456-config1"]
	assert.True(t, exists, "metric-456-config1 should still exist")

	// Cleanup
	regoEvalInstance.Close()
}

// Test_regoEval_EventSubscriptionHandlesMultipleEvents tests multiple event handling
func Test_regoEval_EventSubscriptionHandlesMultipleEvents(t *testing.T) {
	mockSub := &mockEventSubscriber{
		subscribers: make(map[int64]*mockSubscriber),
	}

	re := NewRegoEval(WithEventSubscriber(mockSub))
	regoEvalInstance := re.(*regoEval)

	time.Sleep(50 * time.Millisecond)

	// Add cache entries for multiple metrics
	regoEvalInstance.qc.cache["metric-111-config1"] = &rego.PreparedEvalQuery{}
	regoEvalInstance.qc.cache["metric-222-config1"] = &rego.PreparedEvalQuery{}
	regoEvalInstance.qc.cache["metric-333-config1"] = &rego.PreparedEvalQuery{}
	assert.Equal(t, 3, len(regoEvalInstance.qc.cache))

	// Publish configuration change for metric-222
	event1 := &orchestrator.ChangeEvent{
		Category: orchestrator.EventCategory_EVENT_CATEGORY_METRIC_CONFIGURATION,
		EntityId: "metric-222",
	}
	mockSub.PublishEvent(event1)

	time.Sleep(50 * time.Millisecond)

	// Publish implementation change for metric-333
	event2 := &orchestrator.ChangeEvent{
		Category: orchestrator.EventCategory_EVENT_CATEGORY_METRIC_IMPLEMENTATION,
		EntityId: "metric-333",
	}
	mockSub.PublishEvent(event2)

	time.Sleep(50 * time.Millisecond)

	// Verify correct entries were evicted
	assert.Equal(t, 1, len(regoEvalInstance.qc.cache))
	_, exists := regoEvalInstance.qc.cache["metric-111-config1"]
	assert.True(t, exists, "metric-111-config1 should remain")

	regoEvalInstance.Close()
}

// Test_regoEval_CloseUnsubscribesFromEvents tests cleanup and unsubscription
func Test_regoEval_CloseUnsubscribesFromEvents(t *testing.T) {
	mockSub := &mockEventSubscriber{
		subscribers: make(map[int64]*mockSubscriber),
	}

	re := NewRegoEval(WithEventSubscriber(mockSub))
	regoEvalInstance := re.(*regoEval)

	time.Sleep(50 * time.Millisecond)

	initialCount := mockSub.subscriberCount()
	assert.True(t, initialCount > 0)

	// Close the regoEval
	err := regoEvalInstance.Close()
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	// Verify subscriber was unregistered
	assert.Equal(t, 0, mockSub.subscriberCount())
}

// Test_regoEval_EventSubscriptionIgnoresNilEvents tests nil event handling
func Test_regoEval_EventSubscriptionIgnoresNilEvents(t *testing.T) {
	mockSub := &mockEventSubscriber{
		subscribers: make(map[int64]*mockSubscriber),
	}

	re := NewRegoEval(WithEventSubscriber(mockSub))
	regoEvalInstance := re.(*regoEval)

	time.Sleep(50 * time.Millisecond)

	regoEvalInstance.qc.cache["metric-123-config1"] = &rego.PreparedEvalQuery{}

	// Publish a nil event (should be ignored gracefully)
	mockSub.PublishNilEvent()

	time.Sleep(50 * time.Millisecond)

	// Cache should remain unchanged
	assert.Equal(t, 1, len(regoEvalInstance.qc.cache))

	regoEvalInstance.Close()
}

// Mock event subscriber for testing
type mockEventSubscriber struct {
	subscribers map[int64]*mockSubscriber
	nextID      int64
	mutex       sync.Mutex
}

type mockSubscriber struct {
	ch     chan *orchestrator.ChangeEvent
	filter *orchestrator.SubscribeRequest_Filter
}

func (m *mockEventSubscriber) RegisterSubscriber(filter *orchestrator.SubscribeRequest_Filter) (ch <-chan *orchestrator.ChangeEvent, id int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	channelBuf := make(chan *orchestrator.ChangeEvent, 100)
	id = m.nextID
	m.nextID++

	m.subscribers[id] = &mockSubscriber{
		ch:     channelBuf,
		filter: filter,
	}

	return channelBuf, id
}

func (m *mockEventSubscriber) UnregisterSubscriber(id int64) (err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if sub, ok := m.subscribers[id]; ok {
		delete(m.subscribers, id)
		close(sub.ch)
	}

	return nil
}

func (m *mockEventSubscriber) PublishEvent(event *orchestrator.ChangeEvent) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, sub := range m.subscribers {
		select {
		case sub.ch <- event:
		default:
			// Channel full, skip
		}
	}
}

func (m *mockEventSubscriber) PublishNilEvent() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, sub := range m.subscribers {
		select {
		case sub.ch <- nil:
		default:
			// Channel full, skip
		}
	}
}

func (m *mockEventSubscriber) subscriberCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return len(m.subscribers)
}
