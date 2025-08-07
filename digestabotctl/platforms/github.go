package platforms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
)

type GitHub struct {
	Owner string
	Repo  string
	Token string
}

type GitHubPR struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}

func (g GitHub) CreatePR(pr GitHubPR) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", g.Owner, g.Repo)
	body, err := json.Marshal(pr)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.Token))

	_, err = sendRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func NewGithubPR(g GitHub, p PullRequest) (GitHubPR, error) {
	temp := template.Must(template.New("file").Parse(prTemplate))
	var buf = bytes.Buffer{}
	if err := temp.Execute(&buf, p); err != nil {
		return GitHubPR{}, err
	}

	return GitHubPR{
		Title: p.Title,
		Body:  buf.String(),
		Head:  p.Head,
		Base:  p.Base,
	}, nil
}
