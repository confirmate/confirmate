package collectors

//go:generate buf generate
//go:generate buf generate --template buf.openapi.gen.yaml --path api/collectors/cloud -o openapi/cloudcollector
//go:generate buf generate --template buf.gen.ontology.yaml --path cloud/api/cloud.proto -o cloud/api/ontology
