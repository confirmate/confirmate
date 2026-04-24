package azure

import (
	"testing"

	"confirmate.io/core/api/ontology"
	"confirmate.io/core/util"
	"confirmate.io/core/util/assert"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
)

func Test_tlsCipherSuites(t *testing.T) {
	type args struct {
		cs string
	}
	tests := []struct {
		name string
		args args
		want []*ontology.CipherSuite
	}{
		{
			name: "TLSCipherSuitesTLSAES128GCMSHA256",
			args: args{
				cs: string(armappservice.TLSCipherSuitesTLSAES128GCMSHA256),
			},
			want: []*ontology.CipherSuite{
				{
					SessionCipher: "AES-128-GCM",
					MacAlgorithm:  "SHA-256",
				},
			},
		},
		{
			name: "TLSCipherSuitesTLSECDHERSAWITHAES256GCMSHA384",
			args: args{
				cs: string(armappservice.TLSCipherSuitesTLSECDHERSAWITHAES256GCMSHA384),
			},
			want: []*ontology.CipherSuite{
				{
					AuthenticationMechanism: "RSA",
					KeyExchangeAlgorithm:    "ECDHE",
					SessionCipher:           "AES-256-GCM",
					MacAlgorithm:            "SHA-384",
				},
			},
		},
		{
			name: "not a TLS cipher",
			args: args{
				cs: "NOTTLS_AES_256",
			},
			want: nil,
		},
		{
			name: "invalid authentication",
			args: args{
				cs: "TLS_ECDHE_RSB_WITH_AES_256_GCM_SHA384",
			},
			want: nil,
		},
		{
			name: "invalid authentication",
			args: args{
				cs: "TLS_ECDHE_RSA_WITHOUT_AES_256_GCM_SHA384",
			},
			want: nil,
		},
		{
			name: "invalid session cipher algorithm",
			args: args{
				cs: "TLS_ECDHE_RSA_WITH_AIS_256_GCM_SHA384",
			},
			want: nil,
		},
		{
			name: "invalid session cipher key length",
			args: args{
				cs: "TLS_ECDHE_RSA_WITH_AES_257_GCM_SHA384",
			},
			want: nil,
		},
		{
			name: "invalid session cipher mode",
			args: args{
				cs: "TLS_ECDHE_RSA_WITH_AES_256_FCM_SHA384",
			},
			want: nil,
		},
		{
			name: "invalid mac algorithm",
			args: args{
				cs: "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHO384",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tlsCipherSuites(tt.args.cs)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_tlsVersion(t *testing.T) {
	type args struct {
		version *string
	}
	tests := []struct {
		name string
		args args
		want float32
	}{
		{
			name: "1_3",
			args: args{
				version: util.Ref("1_3"),
			},
			want: 1.3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tlsVersion(tt.args.version); got != tt.want {
				t.Errorf("tlsVersion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_azureStorageCollector_collectDiagnosticSettings(t *testing.T) {
	type fields struct {
		azureCollector *azureCollector
	}
	type args struct {
		resourceURI string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ontology.ActivityLogging
		wantErr assert.WantErr
	}{
		{
			name: "No Diagnostic Setting available",
			fields: fields{
				azureCollector: NewMockAzureCollector(newMockSender()),
			},
			args: args{
				resourceURI: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/res1/providers/Microsoft.Storage/storageAccounts/account3",
			},
			want: nil,
			wantErr: func(t *testing.T, err error, msgAndArgs ...any) bool {
				return assert.ErrorContains(t, err, ErrGettingNextPage.Error())
			},
		},
		{
			name: "Happy path: no workspace available",
			fields: fields{
				azureCollector: NewMockAzureCollector(newMockSender()),
			},
			args: args{
				resourceURI: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/res1/providers/Microsoft.Storage/storageAccounts/account2",
			},
			want:    nil,
			wantErr: assert.NoError,
		},
		{
			name: "Happy path: data logged",
			fields: fields{
				azureCollector: NewMockAzureCollector(newMockSender()),
			},
			args: args{
				resourceURI: "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/res1/providers/Microsoft.Storage/storageAccounts/account1",
			},
			want: &ontology.ActivityLogging{
				Enabled:           true,
				LoggingServiceIds: []string{"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/insights-integration/providers/Microsoft.OperationalInsights/workspaces/workspace1"},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := tt.fields.azureCollector

			// Init Diagnostic Settings Client
			_ = d.initDiagnosticsSettingsClient()

			got, raw, err := d.collectDiagnosticSettings(tt.args.resourceURI)

			tt.wantErr(t, err)
			if tt.wantErr != nil {
				assert.NotNil(t, raw)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
