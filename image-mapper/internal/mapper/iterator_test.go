package mapper

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReaderIterator(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "empty input",
			input:    "",
			expected: nil,
		},
		{
			name:     "single image",
			input:    "nginx",
			expected: []string{"nginx"},
		},
		{
			name:     "multiple images",
			input:    "nginx\nredis\npostgres",
			expected: []string{"nginx", "redis", "postgres"},
		},
		{
			name:     "images with whitespace",
			input:    "nginx\n  redis  \npostgres\n",
			expected: []string{"nginx", "  redis  ", "postgres"},
		},
		{
			name:     "images with empty lines",
			input:    "nginx\n\nredis\n\npostgres",
			expected: []string{"nginx", "redis", "postgres"},
		},
		{
			name:     "images with registry names",
			input:    "gcr.io/project/nginx\nregistry.example.com/redis:latest\npostgres:13",
			expected: []string{"gcr.io/project/nginx", "registry.example.com/redis:latest", "postgres:13"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reader := strings.NewReader(tc.input)
			iterator := NewReaderIterator(reader)

			var results []string
			for {
				image, err := iterator.Next()
				if err == ErrIteratorDone {
					break
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				results = append(results, image)
			}

			if diff := cmp.Diff(tc.expected, results); diff != "" {
				t.Errorf("results mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestReaderIteratorError(t *testing.T) {
	// Test with a reader that returns an error
	errorReader := &errorReader{err: errors.New("test error")}
	iterator := NewReaderIterator(errorReader)

	_, err := iterator.Next()
	if err == nil {
		t.Error("expected error, got nil")
	}
	if err.Error() != "test error" {
		t.Errorf("expected 'test error', got %s", err.Error())
	}
}

func TestArgsIterator(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "empty args",
			args: nil,
		},
		{
			name: "single arg",
			args: []string{"nginx"},
		},
		{
			name: "multiple args",
			args: []string{"nginx", "redis", "postgres"},
		},
		{
			name: "args with empty strings",
			args: []string{"nginx", "", "redis"},
		},
		{
			name: "args with registry names",
			args: []string{"gcr.io/project/nginx", "registry.example.com/redis:latest", "postgres:13"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			iterator := NewArgsIterator(tc.args)

			var results []string
			for {
				image, err := iterator.Next()
				if err == ErrIteratorDone {
					break
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				results = append(results, image)
			}

			if diff := cmp.Diff(tc.args, results); diff != "" {
				t.Errorf("results mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestArgsIteratorReuse(t *testing.T) {
	args := []string{"nginx", "redis", "postgres"}
	iterator := NewArgsIterator(args)

	// First pass
	var firstPass []string
	for {
		image, err := iterator.Next()
		if err == ErrIteratorDone {
			break
		}
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		firstPass = append(firstPass, image)
	}

	// Second pass should return ErrIteratorDone immediately
	image, err := iterator.Next()
	if err != ErrIteratorDone {
		t.Errorf("expected ErrIteratorDone, got %v (image: %s)", err, image)
	}

	// Verify first pass results
	if diff := cmp.Diff(args, firstPass); diff != "" {
		t.Errorf("first pass results mismatch (-want +got):\n%s", diff)
	}
}

// errorReader is a helper type that always returns an error when Read is called
type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}
