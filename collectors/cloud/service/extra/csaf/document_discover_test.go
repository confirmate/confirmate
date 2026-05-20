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
	"strings"
	"testing"

	"confirmate.io/collectors/cloud/internal/config"
	"confirmate.io/collectors/cloud/internal/crypto/openpgp"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/assert"
	"github.com/gocsaf/csaf/v3/csaf"
)

func Test_csafCollector_handleAdvisory(t *testing.T) {
	type fields struct {
		domain string
		ctID   string
		client *http.Client
	}
	type args struct {
		label    csaf.TLPLabel
		file     csaf.AdvisoryFile
		keyring  openpgp.EntityList
		parentId string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantDoc assert.Want[*ontology.SecurityAdvisoryDocument]
		wantErr bool
	}{
		{
			name: "happy path",
			fields: fields{
				domain: goodProvider.Domain(),
				ctID:   config.DefaultTargetOfEvaluationID,
				client: goodProvider.Client(),
			},
			args: args{
				label: csaf.TLPLabelWhite,
				file: csaf.DirectoryAdvisoryFile{
					Path: goodProvider.URL + "/.well-known/csaf/white/2020/some-id.json",
				},
				keyring: goodProvider.Keyring,
			},
			wantDoc: func(t *testing.T, got *ontology.SecurityAdvisoryDocument, msgAndArgs ...any) bool {
				return assert.Equal(t, "some-id", got.Id) &&
					assert.NotEmpty(t, got.Raw) &&
					assert.False(t, strings.Contains(got.Raw, "security_advisory_document"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &csafCollector{
				domain: tt.fields.domain,
				ctID:   tt.fields.ctID,
				client: tt.fields.client,
			}
			gotDoc, err := d.handleAdvisory(tt.args.label, tt.args.file, tt.args.keyring, tt.args.parentId)
			if (err != nil) != tt.wantErr {
				t.Errorf("csafCollector.handleAdvisory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantDoc(t, gotDoc)
		})
	}
}

func Test_csafCollector_collectSecurityAdvisories(t *testing.T) {
	type fields struct {
		domain string
		ctID   string
		client *http.Client
	}
	type args struct {
		md       *csaf.LoadedProviderMetadata
		keyring  openpgp.EntityList
		parentId string
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		wantDocuments assert.Want[[]ontology.IsResource]
		wantErr       bool
	}{
		{
			name: "happy path",
			fields: fields{
				domain: goodProvider.Domain(),
				ctID:   config.DefaultTargetOfEvaluationID,
				client: goodProvider.Client(),
			},
			args: args{
				md: &csaf.LoadedProviderMetadata{
					URL:      goodProvider.WellKnownProviderURL(),
					Document: goodProvider.DocumentAny(),
				},
			},
			wantDocuments: func(t *testing.T, got []ontology.IsResource, msgAndArgs ...any) bool {
				return assert.NotEmpty(t, got) && assert.Equal(t, "some-id", got[0].GetId())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &csafCollector{
				domain: tt.fields.domain,
				ctID:   tt.fields.ctID,
				client: tt.fields.client,
			}
			gotDocuments, err := d.collectSecurityAdvisories(tt.args.md, tt.args.keyring, tt.args.parentId)
			if (err != nil) != tt.wantErr {
				t.Errorf("csafCollector.collectSecurityAdvisories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			tt.wantDocuments(t, gotDocuments)
		})
	}
}
