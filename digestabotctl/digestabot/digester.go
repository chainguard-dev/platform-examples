package digestabot

import (
	"github.com/google/go-containerregistry/pkg/crane"
)

type Digester interface {
	Digest(string) (string, error)
}

type TestDigester struct {
	Old string
	New string
}

func (t TestDigester) Digest(string) (string, error) {
	return t.New, nil
}

type Crane struct{}

func (c Crane) Digest(image string) (string, error) {
	return crane.Digest(image)
}
