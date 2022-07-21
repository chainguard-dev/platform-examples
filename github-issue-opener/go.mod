module chainguard.dev/demos/github-issue-opener

go 1.17

replace chainguard.dev/api => ../../api

require (
	chainguard.dev/api v0.0.0-00010101000000-000000000000
	github.com/cloudevents/sdk-go/v2 v2.8.0
	github.com/coreos/go-oidc/v3 v3.2.0
	github.com/google/go-github/v43 v43.0.0
	github.com/kelseyhightower/envconfig v1.4.0
	golang.org/x/oauth2 v0.0.0-20220718184931-c8730f7fcb92
	knative.dev/pkg v0.0.0-20220715183228-f1f36a2c977e
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/stretchr/testify v1.7.2 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e // indirect
	golang.org/x/net v0.0.0-20220708220712-1185a9018129 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
)

replace k8s.io/apimachinery => k8s.io/apimachinery v0.23.5

replace k8s.io/client-go => k8s.io/client-go v0.23.5

replace k8s.io/code-generator => k8s.io/code-generator v0.23.5

replace knative.dev/pkg => knative.dev/pkg v0.0.0-20220407170445-9ae44fe1fb6d
