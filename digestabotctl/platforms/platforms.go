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
	ValidPlatforms = []string{"github"}
	ErrBadResponse = fmt.Errorf("request was unsuccessful")
)

type PRCreator interface {
	CreatePR()
}

type PullRequest struct {
	Title       string
	Description string
	Diff        string
	Base        string
	Head        string
}

var prTemplate = `{{ $tick := "` + "```" + `" -}}
{{.Description}}

## Changes

<details>

{{ $tick }}diff 

{{ .Diff }}
{{ $tick }}

</details>
`

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
