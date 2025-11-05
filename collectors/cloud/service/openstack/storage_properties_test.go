package openstack

import (
	"testing"

	"confirmate.io/collectors/cloud/internal/testdata"
	"confirmate.io/collectors/cloud/internal/testutil/assert"
	"github.com/gophercloud/gophercloud/v2/openstack/blockstorage/v3/volumes"
)

func Test_getParentID(t *testing.T) {
	type args struct {
		volume *volumes.Volume
	}
	tests := []struct {
		name string
		args args
		want assert.Want[string]
	}{
		{
			name: "Happy path: no attached server available",
			args: args{
				&volumes.Volume{
					TenantID: testdata.MockVolumeTenantID,
				},
			},
			want: func(t *testing.T, got string) bool {
				return assert.Equal(t, testdata.MockVolumeTenantID, got)
			},
		},
		{
			name: "Happy path: attached serverID",
			args: args{
				&volumes.Volume{
					TenantID: testdata.MockVolumeTenantID,
					Attachments: []volumes.Attachment{
						{
							ServerID: testdata.MockServerID1,
						},
					},
				},
			},
			want: func(t *testing.T, got string) bool {
				return assert.Equal(t, testdata.MockServerID1, got)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getParentID(tt.args.volume)

			tt.want(t, got)
		})
	}
}
