module github.com/fercho/school-tracking/services/notification

go 1.24.0

require (
	github.com/caarlos0/env/v10 v10.0.0
	github.com/fercho/school-tracking/proto v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.11.2
	github.com/nats-io/nats.go v1.49.0
	github.com/stretchr/testify v1.11.1
	github.com/swaggo/swag v1.16.6
	github.com/twilio/twilio-go v1.23.5
	go.uber.org/fx v1.24.0
	go.uber.org/zap v1.27.1
	google.golang.org/grpc v1.79.2
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.19.6 // indirect
	github.com/go-openapi/spec v0.20.4 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/golang/mock v1.6.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.18.2 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/nats-io/nkeys v0.4.12 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	go.uber.org/dig v1.19.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.46.0 // indirect
	golang.org/x/mod v0.30.0 // indirect
	golang.org/x/net v0.48.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.32.0 // indirect
	golang.org/x/tools v0.39.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251202230838-ff82c1b0f217 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/fercho/school-tracking/proto => ../../proto
