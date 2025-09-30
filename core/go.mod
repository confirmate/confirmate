module confirmate.io/core

go 1.24.5

// runtime dependencies
require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.9-20250912141014-52f32327d4b0.1
	connectrpc.com/connect v1.18.1
	connectrpc.com/vanguard v0.3.1-0.20250909182909-a5d6122b29b4
	github.com/lib/pq v1.10.9
	github.com/lmittmann/tint v1.1.2
	github.com/mfridman/cli v0.2.1
	github.com/mfridman/xflag v0.1.0 // indirect
	github.com/proullon/ramsql v0.1.4
	golang.org/x/exp v0.0.0-20230713183714-613f0c0eb8a1 // indirect
	golang.org/x/net v0.38.0
	google.golang.org/genproto/googleapis/api v0.0.0-20250922171735-9219d122eba9
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250908214217-97024824d090 // indirect
	google.golang.org/protobuf v1.36.9
)

// build dependencies
require (
	github.com/bmatcuk/doublestar/v4 v4.0.2 // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/google/addlicense v1.2.0
	github.com/lyft/protoc-gen-star/v2 v2.0.3 // indirect
	github.com/spf13/afero v1.5.1 // indirect
	github.com/srikrsna/protoc-gen-gotag v1.0.2
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/sync v0.12.0 // indirect
	golang.org/x/text v0.23.0 // indirect
	golang.org/x/tools v0.21.1-0.20240508182429-e35e4ccd0d2d // indirect
)
