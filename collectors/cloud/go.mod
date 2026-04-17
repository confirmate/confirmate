module confirmate.io/collectors/cloud

go 1.24.5

require (
	confirmate.io/core v0.0.0
	connectrpc.com/connect v1.19.1
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.18.1
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.10.1
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v7 v7.0.0
	github.com/google/uuid v1.6.0
	google.golang.org/protobuf v1.36.11
)

replace confirmate.io/core => ../../core
