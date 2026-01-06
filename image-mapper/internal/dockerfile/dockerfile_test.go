package dockerfile

import (
	"fmt"
	"os"
	"testing"

	"github.com/chainguard-dev/customer-success/scripts/image-mapper/internal/mapper"
	"github.com/google/go-cmp/cmp"
)

type mockMapper struct {
	mappings map[string][]string
}

func (m *mockMapper) Map(img string) (*mapper.Mapping, error) {
	return &mapper.Mapping{
		Image:   img,
		Results: m.mappings[img],
	}, nil
}

func TestMapDockerfile(t *testing.T) {
	m := &mockMapper{
		mappings: map[string][]string{
			"docker.io/python": {
				"cgr.dev/chainguard/python:latest-dev",
			},
			"python": {
				"cgr.dev/chainguard/python:latest-dev",
			},
			"python:3.13": {
				"cgr.dev/chainguard/python:3.13-dev",
			},
		},
	}

	testCases := map[string]struct{}{
		"singlestage": {},
		"multistage":  {},
		"args":        {},
		"copyfrom":    {},
		"runmount":    {},
	}

	for name := range testCases {
		t.Run(name, func(t *testing.T) {
			before, err := os.ReadFile(fmt.Sprintf("testdata/%s.before.Dockerfile", name))
			if err != nil {
				t.Fatalf("unexpected error reading before file: %s", err)
			}

			after, err := os.ReadFile(fmt.Sprintf("testdata/%s.after.Dockerfile", name))
			if err != nil {
				t.Fatalf("unexpected error reading before file: %s", err)
			}

			result, err := mapDockerfile(m, before)
			if err != nil {
				t.Fatalf("unexpected error mapping dockerfile: %s", err)
			}

			if diff := cmp.Diff(after, result); diff != "" {
				t.Errorf("unexpected result:\n%s", diff)
			}
		})
	}
}
