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

package csaf

import (
	"net/http"
	"os"
	"testing"

	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/collectors/cloud/internal/collectortest/csaf/providertest"
	"confirmate.io/collectors/cloud/internal/config"
	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/assert"
	"github.com/gocsaf/csaf/v3/csaf"
	"github.com/google/uuid"
)

// validAdvisory contains the structure of a valid CSAF Advisory that validates against the JSON schema
var validAdvisory = &csaf.Advisory{
	Document: &csaf.Document{
		Category:    new(csaf.DocumentCategory("csaf_security_advisory")),
		CSAFVersion: new(csaf.CSAFVersion20),
		Title:       new("Buffer overflow in Test Product"),
		Publisher: &csaf.DocumentPublisher{
			Name:      new("Test Vendor"),
			Category:  new(csaf.CSAFCategoryVendor),
			Namespace: new("http://localhost"),
		},
		Tracking: &csaf.Tracking{
			ID:                 new(csaf.TrackingID("some-id")),
			CurrentReleaseDate: new("2020-07-01T10:09:07Z"),
			InitialReleaseDate: new("2020-07-01T10:09:07Z"),
			Generator: &csaf.Generator{
				Date: new("2020-07-01T10:09:07Z"),
				Engine: &csaf.Engine{
					Name:    new("test"),
					Version: new("1.0"),
				},
			},
			Status:  new(csaf.CSAFTrackingStatusFinal),
			Version: new(csaf.RevisionNumber("1")),
			RevisionHistory: csaf.Revisions{
				&csaf.Revision{
					Date:    new("2020-07-01T10:09:07Z"),
					Number:  new(csaf.RevisionNumber("1")),
					Summary: new("First and final version"),
				},
			},
		},
	},
	ProductTree: &csaf.ProductTree{
		Branches: csaf.Branches{
			&csaf.Branch{
				Category: new(csaf.CSAFBranchCategoryVendor),
				Name:     new("Test Vendor"),
				Product: &csaf.FullProductName{
					Name:      new("Test Product"),
					ProductID: new(csaf.ProductID("CSAFPID-0001")),
				},
			},
		},
	},
}

var goodProvider *providertest.TrustedProvider

func TestMain(m *testing.M) {
	var advisories = map[csaf.TLPLabel][]*csaf.Advisory{
		csaf.TLPLabelWhite: {
			validAdvisory,
		},
	}

	goodProvider = providertest.NewTrustedProvider(
		advisories,
		providertest.NewGoodIndexTxtWriter(),
		func(pmd *csaf.ProviderMetadata) {
			pmd.Publisher = &csaf.Publisher{
				Name:      new("Test Vendor"),
				Category:  new(csaf.CSAFCategoryVendor),
				Namespace: new("http://localhost"),
			}
		})
	defer goodProvider.Close()

	code := m.Run()
	os.Exit(code)
}

func TestNewTrustedProviderCollector(t *testing.T) {
	type args struct {
		opts []CollectorOption
	}
	tests := []struct {
		name string
		args args
		want collector.Collector
	}{
		{
			name: "Happy path",
			args: args{},
			want: &csafCollector{
				ctID:   config.DefaultTargetOfEvaluationID,
				domain: "confirmate.io",
				id:     uuid.NewSHA1(uuid.NameSpaceOID, []byte("csaf::"+config.DefaultTargetOfEvaluationID+"::confirmate.io")).String(),
				client: http.DefaultClient,
			},
		},
		{
			name: "Happy path: with target of evaluation id",
			args: args{
				opts: []CollectorOption{WithTargetOfEvaluationID(testdata.MockTargetOfEvaluationID1)},
			},
			want: &csafCollector{
				ctID:   testdata.MockTargetOfEvaluationID1,
				domain: "confirmate.io",
				id:     uuid.NewSHA1(uuid.NameSpaceOID, []byte("csaf::"+testdata.MockTargetOfEvaluationID1+"::confirmate.io")).String(),
				client: http.DefaultClient,
			},
		},
		{
			name: "Happy path: with domain",
			args: args{
				opts: []CollectorOption{WithProviderDomain("mock")},
			},
			want: &csafCollector{
				ctID:   config.DefaultTargetOfEvaluationID,
				client: http.DefaultClient,
				domain: "mock",
				id:     uuid.NewSHA1(uuid.NameSpaceOID, []byte("csaf::"+config.DefaultTargetOfEvaluationID+"::mock")).String(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTrustedProviderCollector(tt.args.opts...)
			assert.Equal(t, tt.want, got, assert.CompareAllUnexported())
		})
	}
}

func Test_csafCollector_List(t *testing.T) {
	type fields struct {
		domain string
		ctID   string
		client *http.Client
	}
	tests := []struct {
		name     string
		fields   fields
		wantList assert.Want[[]ontology.IsResource]
		wantErr  assert.WantErr
	}{
		{
			name: "fail",
			fields: fields{
				domain: "localhost:1234",
				client: http.DefaultClient,
				ctID:   config.DefaultTargetOfEvaluationID,
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "could not load provider-metadata.json")
			},
			wantList: assert.Empty[[]ontology.IsResource],
		},
		{
			name: "happy path",
			fields: fields{
				domain: goodProvider.Domain(),
				client: goodProvider.Client(),
				ctID:   config.DefaultTargetOfEvaluationID,
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.NoError(t, err)
			},
			wantList: func(t *testing.T, got []ontology.IsResource, msgAndArgs ...any) bool {
				return assert.NotEmpty(t, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &csafCollector{
				domain: tt.fields.domain,
				client: tt.fields.client,
				ctID:   tt.fields.ctID,
			}
			gotList, err := d.List()
			tt.wantErr(t, err)
			tt.wantList(t, gotList)
		})
	}
}

func Test_csafCollector_TargetOfEvaluationID(t *testing.T) {
	type fields struct {
		domain string
		ctID   string
		client *http.Client
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Happy path",
			fields: fields{
				ctID: testdata.MockTargetOfEvaluationID1,
			},
			want: testdata.MockTargetOfEvaluationID1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &csafCollector{
				domain: tt.fields.domain,
				ctID:   tt.fields.ctID,
				client: tt.fields.client,
			}
			if got := d.TargetOfEvaluationID(); got != tt.want {
				t.Errorf("csafCollector.TargetOfEvaluationID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_csafCollector_ID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "Happy path",
			id:   "csaf-collector-id",
			want: "csaf-collector-id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &csafCollector{id: tt.id}
			assert.Equal(t, tt.want, d.ID())
		})
	}
}

func Test_csafCollector_Collect(t *testing.T) {
	tests := []struct {
		name      string
		collector *csafCollector
		wantErr   assert.WantErr
		wantList  assert.Want[[]ontology.IsResource]
	}{
		{
			name: "fail",
			collector: &csafCollector{
				domain: "localhost:1234",
				client: http.DefaultClient,
				ctID:   config.DefaultTargetOfEvaluationID,
			},
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "could not load provider-metadata.json")
			},
			wantList: assert.Empty[[]ontology.IsResource],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotList, err := tt.collector.Collect()
			tt.wantErr(t, err)
			tt.wantList(t, gotList)
		})
	}
}
