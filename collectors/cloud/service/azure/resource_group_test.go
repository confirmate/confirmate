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

package azure

import (
	"testing"

	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util/assert"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription"
)

func Test_azureResourceGroupCollector_handleSubscription(t *testing.T) {
	type fields struct {
		azureCollector *azureCollector
	}
	type args struct {
		s *armsubscription.Subscription
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *ontology.Account
	}{
		{
			name: "Happy path",
			fields: fields{
				azureCollector: NewMockAzureCollector(newMockSender()),
			},
			args: args{
				s: &armsubscription.Subscription{
					SubscriptionID: new(testdata.MockSubscriptionID),
					DisplayName:    new("Wonderful Subscription"),
					ID:             new(testdata.MockSubscriptionResourceID),
				},
			},
			want: &ontology.Account{
				Id:   testdata.MockSubscriptionResourceID,
				Name: "Wonderful Subscription",
				Raw:  string(`{"*armsubscription.Subscription":[{"displayName":"Wonderful Subscription","id":"/subscriptions/00000000-0000-0000-0000-000000000000","subscriptionId":"00000000-0000-0000-0000-000000000000"}]}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.fields.azureCollector

			got := d.handleSubscription(tt.args.s)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_azureResourceGroupCollector_handleResourceGroup(t *testing.T) {
	type fields struct {
		azureCollector *azureCollector
	}
	type args struct {
		rg *armresources.ResourceGroup
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   ontology.IsResource
	}{
		{
			name: "Happy path",
			fields: fields{
				azureCollector: NewMockAzureCollector(newMockSender()),
			},
			args: args{
				rg: &armresources.ResourceGroup{
					ID:       new(testdata.MockResourceGroupID),
					Name:     new("res1"),
					Location: new("westus"),
					Tags: map[string]*string{
						"tag1Key": new("tag1"),
						"tag2Key": new("tag2"),
					},
				},
			},
			want: &ontology.ResourceGroup{
				Id:   testdata.MockResourceGroupID,
				Name: "res1",
				GeoLocation: &ontology.GeoLocation{
					Region: "westus",
				},
				Labels: map[string]string{
					"tag2Key": "tag2",
					"tag1Key": "tag1",
				},
				ParentId: new(testdata.MockSubscriptionResourceID),
				Raw:      string(`{"*armresources.ResourceGroup":[{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/res1","location":"westus","name":"res1","tags":{"tag1Key":"tag1","tag2Key":"tag2"}}]}`),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.fields.azureCollector

			got := d.handleResourceGroup(tt.args.rg)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_azureResourceGroupCollector_collectResourceGroups(t *testing.T) {
	type fields struct {
		azureCollector *azureCollector
	}
	tests := []struct {
		name     string
		fields   fields
		wantList []ontology.IsResource
		wantErr  assert.WantErr
	}{
		{
			name: "Collector error",
			fields: fields{
				// Intentionally use wrong sender
				azureCollector: NewMockAzureCollector(nil),
			},
			wantList: nil,
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, "could not get next page: GET ")
			},
		},
		{
			name: "Happy path",
			fields: fields{
				azureCollector: NewMockAzureCollector(newMockSender(),
					WithSubscription(&armsubscription.Subscription{
						DisplayName:    new("displayName"),
						ID:             new("/subscriptions/00000000-0000-0000-0000-000000000000"),
						SubscriptionID: new("00000000-0000-0000-0000-000000000000"),
					})),
			},
			wantList: []ontology.IsResource{
				&ontology.Account{
					Id:   "/subscriptions/00000000-0000-0000-0000-000000000000",
					Name: "displayName",
					Raw:  string(`{"*armsubscription.Subscription":[{"displayName":"displayName","id":"/subscriptions/00000000-0000-0000-0000-000000000000","subscriptionId":"00000000-0000-0000-0000-000000000000"}]}`),
				},
				&ontology.ResourceGroup{
					Id:   "/subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/res1",
					Name: "res1",
					GeoLocation: &ontology.GeoLocation{
						Region: "westus",
					},
					Labels: map[string]string{
						"testKey1": "testTag1",
						"testKey2": "testTag2",
					},
					ParentId: new("/subscriptions/00000000-0000-0000-0000-000000000000"),
					Raw:      string(`{"*armresources.ResourceGroup":[{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/res1","location":"westus","name":"res1","tags":{"testKey1":"testTag1","testKey2":"testTag2"}}]}`),
				},
				&ontology.ResourceGroup{
					Id:   "/subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/res2",
					Name: "res2",
					GeoLocation: &ontology.GeoLocation{
						Region: "eastus",
					},
					Labels: map[string]string{
						"testKey1": "testTag1",
						"testKey2": "testTag2",
					},
					ParentId: new("/subscriptions/00000000-0000-0000-0000-000000000000"),
					Raw:      string(`{"*armresources.ResourceGroup":[{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/res2","location":"eastus","name":"res2","tags":{"testKey1":"testTag1","testKey2":"testTag2"}}]}`),
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Happy path: with given resource group",
			fields: fields{
				azureCollector: NewMockAzureCollector(newMockSender(),
					WithResourceGroup("res1"),
					WithSubscription(&armsubscription.Subscription{
						DisplayName:    new("displayName"),
						ID:             new("/subscriptions/00000000-0000-0000-0000-000000000000"),
						SubscriptionID: new("00000000-0000-0000-0000-000000000000"),
					})),
			},
			wantList: []ontology.IsResource{
				&ontology.Account{
					Id:   "/subscriptions/00000000-0000-0000-0000-000000000000",
					Name: "displayName",
					Raw:  string(`{"*armsubscription.Subscription":[{"displayName":"displayName","id":"/subscriptions/00000000-0000-0000-0000-000000000000","subscriptionId":"00000000-0000-0000-0000-000000000000"}]}`),
				},
				&ontology.ResourceGroup{
					Id:   "/subscriptions/00000000-0000-0000-0000-000000000000/resourcegroups/res1",
					Name: "res1",
					GeoLocation: &ontology.GeoLocation{
						Region: "westus",
					},
					Labels: map[string]string{
						"testKey1": "testTag1",
						"testKey2": "testTag2",
					},
					ParentId: new("/subscriptions/00000000-0000-0000-0000-000000000000"),
					Raw:      string(`{"*armresources.ResourceGroup":[{"id":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/res1","location":"westus","name":"res1","tags":{"testKey1":"testTag1","testKey2":"testTag2"}}]}`),
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.fields.azureCollector

			gotList, err := d.collectResourceGroups()

			assert.Equal(t, tt.wantList, gotList)
			tt.wantErr(t, err)
		})
	}
}
