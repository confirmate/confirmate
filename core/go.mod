module confirmate.io/core

go 1.24.6

// runtime dependencies - CLI
require (
	github.com/fatih/color v1.18.0 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20211117102719-0474bc63780f
	github.com/lmittmann/tint v1.1.3
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20
	github.com/urfave/cli/v3 v3.6.2
	golang.org/x/sys v0.40.0 // indirect
)

// runtime dependencies - protobuf/Connect
require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.11-20251209175733-2a1774d88802.1
	buf.build/go/protovalidate v1.1.0
	cel.dev/expr v0.25.1 // indirect
	connectrpc.com/connect v1.19.1
	connectrpc.com/vanguard v0.3.1-0.20250909182909-a5d6122b29b4
	github.com/antlr4-go/antlr/v4 v4.13.1 // indirect
	github.com/google/cel-go v0.27.0 // indirect
	github.com/google/uuid v1.6.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260203192932-546029d2fa20
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260203192932-546029d2fa20 // indirect
	google.golang.org/protobuf v1.36.11
)

// runtime dependencies - database
require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.8.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/proullon/ramsql v0.1.4
	golang.org/x/crypto v0.47.0 // indirect
	golang.org/x/exp v0.0.0-20260112195511-716be5621a96 // indirect
	golang.org/x/text v0.33.0 // indirect
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.31.1
)

// runtime dependencies - assessment
require (
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/lestrrat-go/blackmagic v1.0.4 // indirect
	github.com/lestrrat-go/dsig v1.0.0 // indirect
	github.com/lestrrat-go/dsig-secp256k1 v1.0.0 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc/v3 v3.0.2 // indirect
	github.com/lestrrat-go/jwx/v3 v3.0.13 // indirect
	github.com/lestrrat-go/option/v2 v2.0.0 // indirect
	github.com/open-policy-agent/opa v1.13.1-0.20260206180421-90d65b74abc9
	github.com/rcrowley/go-metrics v0.0.0-20250401214520-65e299d6c5c9 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/sirupsen/logrus v1.9.4 // indirect
	github.com/tchap/go-patricia/v2 v2.3.3 // indirect
	github.com/valyala/fastjson v1.6.7 // indirect
	github.com/vektah/gqlparser/v2 v2.5.31 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/yashtewari/glob-intersection v0.2.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4
	sigs.k8s.io/yaml v1.6.0 // indirect
)

// test dependencies
require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/google/go-cmp v0.7.0
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.11.1
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// build dependencies
require (
	github.com/bmatcuk/doublestar/v4 v4.10.0 // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/google/addlicense v1.2.0
	github.com/lyft/protoc-gen-star/v2 v2.0.4 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/srikrsna/protoc-gen-gotag v1.0.2
	golang.org/x/mod v0.32.0 // indirect
	golang.org/x/sync v0.19.0
	golang.org/x/tools v0.41.0 // indirect
)

require google.golang.org/grpc v1.78.0

require (
	github.com/robfig/cron/v3 v3.0.1 // indirect
	go.uber.org/atomic v1.9.0 // indirect
)

require github.com/go-co-op/gocron v1.37.0

/// Use confirmate/ramsql fork instead of proullon/ramsql due to required bugfixes and compatibility
/// improvements not present in upstream.
replace github.com/proullon/ramsql => github.com/confirmate/ramsql v0.0.0-20260129104154-5b108a47b09b
