package digestabot

import (
	"bufio"
	"bytes"
	"io"
	"log/slog"
	"os"
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
