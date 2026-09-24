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

package ionos

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	collector "confirmate.io/collectors/cloud/internal/collector"
	"confirmate.io/collectors/cloud/internal/config"
	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/core/util/assert"

	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

type mockSender struct{}

// mockErrorSender is used to simulate errors in the RoundTrip method.
type mockErrorSender struct{}

func newMockSender() *mockSender {
	return &mockSender{}
}

func newMockErrorSender() *mockErrorSender {
	return &mockErrorSender{}
}

// RoundTrip implements http.RoundTripper.
func (mockSender) RoundTrip(req *http.Request) (res *http.Response, err error) {
	if hasEmptySegmentInURL(req.URL.Path) {
		return createResponse(req, map[string]any{}, 404)
	} else if strings.HasSuffix(req.URL.Path, "/labels") {
		return createResponse(req, map[string]any{
			"items": []map[string]any{
				{
					"id": "label-1",
					"properties": map[string]any{
						"key":   "label1",
						"value": "value1",
					},
				},
			},
		}, 200)
	}

	switch req.URL.Path {
	case "/datacenters":
		return createResponse(req, map[string]any{
			"items": []map[string]any{
				{
					"id": testdata.MockIonosDatacenterID1,
					"properties": map[string]any{
						"name":        testdata.MockIonosDatacenterName1,
						"description": testdata.MockIonosDatacenterDescription1,
						"location":    testdata.MockIonosDatacenterLocation1,
					},
					"metadata": map[string]any{},
				},
				{
					"id": testdata.MockIonosDatacenterID2,
					"properties": map[string]any{
						"name":        testdata.MockIonosDatacenterName2,
						"description": testdata.MockIonosDatacenterDescription2,
						"location":    testdata.MockIonosDatacenterLocation2,
					},
					"metadata": map[string]any{},
				},
			},
		}, 200)
	case "/datacenters/" + testdata.MockIonosDatacenterID1 + "/servers":
		return createResponse(req, map[string]any{
			"items": []map[string]any{
				{
					"id": testdata.MockIonosVMID1,
					"properties": map[string]any{
						"name": testdata.MockIonosVMName1,
					},
					"metadata": map[string]any{
						"createdDate": testdata.MockIonosCreationTime,
					},
					"entities": map[string]any{
						"volumes": map[string]any{
							"items": []map[string]any{
								{
									"id": testdata.MockIonosVolumeID1,
									"properties": map[string]any{
										"name": testdata.MockIonosVolumeName1,
									},
								},
							},
						},
						"nics": map[string]any{
							"items": []map[string]any{
								{
									"id": testdata.MockIonosNicID1,
									"properties": map[string]any{
										"name": testdata.MockIonosNicName1,
									},
								},
							},
						},
					},
				},
			},
		}, 200)
	case "/datacenters/" + testdata.MockIonosDatacenterID1 + "/volumes":
		return createResponse(req, map[string]any{
			"items": []map[string]any{
				{
					"id": testdata.MockIonosVolumeID1,
					"properties": map[string]any{
						"name": testdata.MockIonosVolumeName1,
					},
					"metadata": map[string]any{
						"createdDate": testdata.MockIonosCreationTime,
					},
				},
			},
		}, 200)
	case "/datacenters/" + testdata.MockIonosDatacenterID1 + "/loadbalancers":
		return createResponse(req, map[string]any{
			"items": []map[string]any{
				{
					"id": testdata.MockIonosLoadBalancerID1,
					"properties": map[string]any{
						"name": testdata.MockIonosLoadBalancerName1,
					},
					"metadata": map[string]any{
						"createdDate": testdata.MockIonosCreationTime,
					},
					"entities": map[string]any{
						"balancednics": map[string]any{
							"items": []map[string]any{
								{
									"id": testdata.MockIonosNicID1,
									"properties": map[string]any{
										"name": testdata.MockIonosNicName1,
									},
								},
							},
						},
					},
				},
			},
		}, 200)
	case "/datacenters/" + testdata.MockIonosDatacenterID2 + "/servers",
		"/datacenters/" + testdata.MockIonosDatacenterID2 + "/volumes",
		"/datacenters/" + testdata.MockIonosDatacenterID2 + "/loadbalancers":
		return createResponse(req, map[string]any{
			"items": []map[string]any{},
		}, 200)
	default:
		log.Error("not handling mock for path yet", "path", req.URL.Path)
		return createResponse(req, map[string]any{}, 404)
	}
}

func (mockErrorSender) RoundTrip(req *http.Request) (res *http.Response, err error) {
	return createResponse(req, map[string]any{}, 404)
}

func createResponse(req *http.Request, body any, status int) (*http.Response, error) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(buf),
		Request:    req,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}, nil
}

// newMockIonosCollector returns a collector whose client talks to a mocked transport instead of a real server.
func newMockIonosCollector(roundTrip http.RoundTripper) *ionosCollector {
	_, isErrClient := roundTrip.(*mockErrorSender)

	return &ionosCollector{
		ctID:       config.DefaultTargetOfEvaluationID,
		authConfig: &ionoscloud.Configuration{HTTPClient: &http.Client{Transport: roundTrip}},
		client:     mockAPIClient(isErrClient),
	}
}

func mockAPIClient(errClient bool) *ionoscloud.APIClient {
	transport := http.RoundTripper(newMockSender())
	if errClient {
		transport = newMockErrorSender()
	}

	return ionoscloud.NewAPIClient(&ionoscloud.Configuration{
		HTTPClient: &http.Client{Transport: transport},
		Servers: ionoscloud.ServerConfigurations{
			{URL: "https://mock"},
		},
	})
}

func Test_ionosCollector_Name(t *testing.T) {
	d := &ionosCollector{}

	assert.Equal(t, "IONOS Cloud", d.Name())
}

func TestNewIonosCollector(t *testing.T) {
	type args struct {
		opts []CollectorOption
	}
	tests := []struct {
		name string
		args args
		want assert.Want[collector.Collector]
	}{
		{
			name: "error: authorizer not set",
			args: args{},
			want: assert.Nil[collector.Collector],
		},
		{
			name: "Happy path",
			args: args{
				opts: []CollectorOption{
					WithAuthorizer(&ionoscloud.Configuration{HTTPClient: http.DefaultClient}),
					WithTargetOfEvaluationID(testdata.MockTargetOfEvaluationID1),
				},
			},
			want: assert.NotNil[collector.Collector],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewIonosCollector(tt.args.opts...)
			tt.want(t, got)
		})
	}
}

func Test_ionosCollector_TargetOfEvaluationID(t *testing.T) {
	d := &ionosCollector{ctID: testdata.MockTargetOfEvaluationID1}

	assert.Equal(t, testdata.MockTargetOfEvaluationID1, d.TargetOfEvaluationID())
}

func TestNewAuthorizer(t *testing.T) {
	cfg, err := NewAuthorizer()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}

func Test_ionosCollector_List(t *testing.T) {
	d := newMockIonosCollector(newMockSender())

	list, err := d.List()

	assert.NoError(t, err)
	// 2 datacenters, 1 block storage, 1 server, 1 network interface, 1 load balancer, 1 network interface
	assert.Equal(t, 7, len(list))
}

func Test_ionosCollector_Collect(t *testing.T) {
	d := newMockIonosCollector(newMockSender())

	list, err := d.Collect()

	assert.NoError(t, err)
	assert.Equal(t, 7, len(list))
}
