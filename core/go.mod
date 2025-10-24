module confirmate.io/core

go 1.24.5

// runtime dependencies
require (
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.10-20250912141014-52f32327d4b0.1
	connectrpc.com/connect v1.19.1
	connectrpc.com/vanguard v0.3.1-0.20250909182909-a5d6122b29b4
	github.com/lmittmann/tint v1.1.2
	github.com/mfridman/cli v0.2.1
	github.com/mfridman/xflag v0.1.0 // indirect
	golang.org/x/net v0.46.0
	google.golang.org/genproto/googleapis/api v0.0.0-20251020155222-88f65dc88635
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251020155222-88f65dc88635 // indirect
	google.golang.org/protobuf v1.36.10
)

// build dependencies
require (
	github.com/bmatcuk/doublestar/v4 v4.9.1 // indirect
	github.com/fatih/structtag v1.2.0 // indirect
	github.com/google/addlicense v1.2.0
	github.com/lyft/protoc-gen-star/v2 v2.0.4 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/srikrsna/protoc-gen-gotag v1.0.2
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/sync v0.17.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	golang.org/x/tools v0.38.0 // indirect
)

// gorm
require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	gorm.io/gorm v1.31.0
)

replace github.com/proullon/ramsql v0.1.4 => github.com/confirmate/ramsql v0.0.0-20251001100412-1283bfcd79dc

require (
	github.com/google/go-cmp v0.7.0
	github.com/proullon/ramsql v0.1.4
	github.com/stretchr/testify v1.11.1
	gorm.io/driver/postgres v1.6.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.7.6 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	golang.org/x/crypto v0.43.0 // indirect
	golang.org/x/exp v0.0.0-20251017212417-90e834f514db // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
