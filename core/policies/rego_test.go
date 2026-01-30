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
	"testing"
	"time"

	"confirmate.io/core/api/assessment"
	"confirmate.io/core/api/evidence"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/assert"
	"confirmate.io/core/util/prototest"
	"confirmate.io/core/util/testdata"
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
					Applicable:  true,
					Compliant:   true,
					TargetValue: true,
					Operator:    "==",
					MetricID:    "84eaed86-759d-4419-9954-f3d3ea1f5200",
					MetricName:  "AutomaticUpdatesEnabled",
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
					Applicable:  true,
					Compliant:   false,
					TargetValue: false,
					Operator:    "==",
					MetricID:    "84eaed86-759d-4419-9954-f3d3ea1f5200",
					MetricName:  "AutomaticUpdatesEnabled",
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
