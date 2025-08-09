package platforms

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"
)

type GitPlatform string

var (
	GitHubPlatform GitPlatform = "github"
	GitLabPlatform GitPlatform = "gitlab"
)

var successCodes = []int{200, 201, 202, 203, 204}

var (
	//ValidPlatforms     = []string{"github", "gitlab"}
	ErrBadResponse     = errors.New("request was unsuccessful")
	ErrInvalidPlatform = errors.New("invalid platform")
)

var ValidPlatforms = map[GitPlatform]func(PullRequest) (PRCreator, error){
	GitHubPlatform: NewGitHub,
	GitLabPlatform: NewGitLab,
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

// Entrypoint to create PR. It's simple but allows us flexibility later.
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
