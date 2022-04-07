module chainguard.dev/demos/github-issue-opener

go 1.17

replace chainguard.dev/api => ../../api

require (
	chainguard.dev/api v0.0.0-00010101000000-000000000000
	github.com/cloudevents/sdk-go/v2 v2.8.0
	github.com/coreos/go-oidc/v3 v3.1.0
	github.com/google/go-github/v43 v43.0.0
	github.com/kelseyhightower/envconfig v1.4.0
	golang.org/x/oauth2 v0.0.0-20220309155454-6242fa91716a
	knative.dev/pkg v0.0.0-20220329144915-0a1ec2e0d46c
)

require (
	chainguard.dev/go-grpc-kit v0.3.0 // indirect
	github.com/blendle/zapdriver v1.3.1 // indirect
	github.com/go-logr/logr v1.2.2 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.7.3 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.7.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220106191415-9b9b3d81d5e3 // indirect
	golang.org/x/net v0.0.0-20220325170049-de3da57026de // indirect
	golang.org/x/sys v0.0.0-20220328115105-d36c6a25d886 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220324131243-acbaeb5b85eb // indirect
	google.golang.org/grpc v1.45.0 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/square/go-jose.v2 v2.6.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/api v0.23.5 // indirect
	k8s.io/apimachinery v0.23.5 // indirect
	k8s.io/klog/v2 v2.60.1-0.20220317184644-43cc75f9ae89 // indirect
	k8s.io/utils v0.0.0-20220210201930-3a6ce19ff2f9 // indirect
	sigs.k8s.io/json v0.0.0-20211208200746-9f7c6b3444d2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
)

replace k8s.io/apimachinery => k8s.io/apimachinery v0.23.5

replace k8s.io/client-go => k8s.io/client-go v0.23.5

replace k8s.io/code-generator => k8s.io/code-generator v0.23.5

replace knative.dev/pkg => knative.dev/pkg v0.0.0-20220407170445-9ae44fe1fb6d
