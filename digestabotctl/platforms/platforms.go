package platforms

import (
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"
)

var successCodes = []int{200, 201, 202, 203, 204}

var (
	ValidPlatforms = []string{"github", "gitlab"}
	ErrBadResponse = fmt.Errorf("request was unsuccessful")
)

var prTemplate = `{{ $tick := "` + "```" + `" -}}
{{.Description}}

## Changes

<details>

{{ $tick }}diff 

{{ .Diff }}
{{ $tick }}

</details>
`

type PRCreator interface {
	CreatePR() error
}

type PullRequest struct {
	Title       string
	Description string
	Diff        string
	Base        string
	Head        string
	RepoData
}

type RepoData struct {
	Repo  string
	Owner string
	Token string
}

func CreatePR(p PRCreator) error {
	return p.CreatePR()
}

func sendRequest(req *http.Request) ([]byte, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if !slices.Contains(successCodes, resp.StatusCode) {
		return nil, fmt.Errorf("%w status: %d body: %s", ErrBadResponse, resp.StatusCode, string(body))
	}

	return body, nil
}
