module confirmate.io/collectors/cloud

go 1.25.2

require confirmate.io/core v0.0.0

replace confirmate.io/core => ../../core

require (
	connectrpc.com/connect v1.19.1
	github.com/go-co-op/gocron v1.37.0
	github.com/google/go-cmp v0.7.0
	github.com/lmittmann/tint v1.1.3
	github.com/sirupsen/logrus v1.9.4
	github.com/urfave/cli/v3 v3.6.2
	google.golang.org/protobuf v1.36.11
	k8s.io/apimachinery v0.34.1
)

require (
	github.com/cloudflare/circl v1.6.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/tools v0.41.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/utils v0.0.0-20250604170112-4c0f3b243397 // indirect
	sigs.k8s.io/json v0.0.0-20241014173422-cfa47c3a1cc8 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
	sigs.k8s.io/structured-merge-diff/v6 v6.3.0 // indirect
)

// runtime dependencies (AWS)
require (
	github.com/aws/aws-sdk-go-v2 v1.39.6
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.7.3 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.31.17
	github.com/aws/aws-sdk-go-v2/credentials v1.18.21 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.13 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.4.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.264.0
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.9.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.19.13 // indirect
	github.com/aws/aws-sdk-go-v2/service/lambda v1.81.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.90.0
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.39.1
	github.com/aws/smithy-go v1.23.2
)

// runtime dependencies (Azure)
require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.20.0
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.10.1
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2 v2.3.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v3 v3.0.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos v1.0.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dataprotection/armdataprotection v1.0.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/machinelearning/armmachinelearning v1.0.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor v0.11.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork v1.1.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources v1.2.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/security/armsecurity v0.14.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/sql/armsql v1.2.0
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage v1.8.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/subscription/armsubscription v1.2.0
	github.com/AzureAD/microsoft-authentication-library-for-go v1.4.2 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	golang.org/x/net v0.49.0 // indirect
)

// runtime dependencies (CSAF)
require (
	github.com/Intevation/gval v1.3.0 // indirect
	github.com/Intevation/jsonpath v0.2.1 // indirect
	github.com/ProtonMail/go-crypto v1.3.0
	github.com/gocsaf/csaf/v3 v3.4.0
	github.com/shopspring/decimal v1.4.0 // indirect
	go.etcd.io/bbolt v1.4.3 // indirect
	golang.org/x/time v0.14.0 // indirect
)

// runtime dependencies (k8s)
require (
	github.com/emicklei/go-restful/v3 v3.12.2 // indirect
	github.com/go-openapi/jsonpointer v0.21.0 // indirect
	github.com/go-openapi/jsonreference v0.20.2 // indirect
	github.com/go-openapi/swag v0.23.0 // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	golang.org/x/term v0.39.0 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.12.0 // indirect
	k8s.io/api v0.34.1
	k8s.io/client-go v0.34.1
	k8s.io/kube-openapi v0.0.0-20250710124328-f3f2b991d03b // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

// runtime dependencies (OpenStack)
require github.com/gophercloud/gophercloud/v2 v2.7.0

// runtime dependencies (security)
require (
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/oauth2 v0.32.0 // indirect
)

// other runtime dependencies
require (
	github.com/google/uuid v1.6.0
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/santhosh-tekuri/jsonschema/v6 v6.0.2 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	google.golang.org/grpc v1.78.0
)
