package digestabot

import (
	"bufio"
	"bytes"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestUpdateHashes(t *testing.T) {
	tt := []struct {
		name        string
		inFile      io.Reader
		outFile     *bufio.Writer
		buf         *bytes.Buffer
		currentHash string
		newHash     string
		expected    string
	}{
		{
			name:        "Dockerfile single line replacement",
			currentHash: "sha256:123456",
			newHash:     "sha256:456789",
			inFile:      strings.NewReader("FROM cgr.dev/chainguard/python:latest@sha256:123456"),
			expected:    "FROM cgr.dev/chainguard/python:latest@sha256:456789\n",
			buf:         new(bytes.Buffer),
		},
		{
			name:        "Dockerfile multi-line replacement",
			currentHash: "sha256:123456",
			newHash:     "sha256:456789",
			inFile:      strings.NewReader("FROM cgr.dev/chainguard/python:latest@sha256:123456\n\nRUN test\n\nFROM cgr.dev/chainguard/python:latest@sha256:123456"),
			expected:    "FROM cgr.dev/chainguard/python:latest@sha256:456789\n\nRUN test\n\nFROM cgr.dev/chainguard/python:latest@sha256:456789\n",
			buf:         new(bytes.Buffer),
		},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			digester := TestDigester{
				Old: v.currentHash,
				New: v.newHash,
			}
			opts := UpdateOptions{
				Name:     "testing",
				Digester: digester,
				InFile:   v.inFile,
				OutFile:  bufio.NewWriter(v.buf),
				Logger:   slog.New(slog.NewTextHandler(os.Stdout, nil)),
			}
			if err := UpdateHashes(opts); err != nil {
				t.Fatal(err)
			}

			out := v.buf.String()
			if out != v.expected {
				t.Errorf("expected \n%s \n but got \n%s", v.expected, out)
			}

		})
	}
}

func TestFindFiles(t *testing.T) {
	tt := []struct {
		name      string
		directory string
		fileTypes []string
		expected  []string
	}{
		{name: "dockerfiles", directory: "../examples", fileTypes: []string{"Dockerfile*"}, expected: []string{"../examples/Dockerfile", "../examples/Dockerfile2", "../examples/nested_dir/Dockerfile"}},
		{name: "all", directory: "../examples", fileTypes: DefaultFileTypes, expected: []string{"../examples/Dockerfile", "../examples/Dockerfile2", "../examples/Makefile", "../examples/nested_dir/Dockerfile", "../examples/test.yaml"}},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			files, err := FindFiles(v.fileTypes, v.directory)
			if err != nil {
				t.Fatal(err)
			}

			if !slices.Equal(v.expected, files) {
				t.Errorf("expected \n%v \n but got \n%v", v.expected, files)
			}
		})
	}
}
