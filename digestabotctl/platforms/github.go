package platforms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"text/template"
)

type GitHub struct {
	Owner string
	Repo  string
	Token string
	GitHubPR
}

type GitHubPR struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Head  string `json:"head"`
	Base  string `json:"base"`
}

func (g GitHub) CreatePR(logger *slog.Logger) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", g.Owner, g.Repo)
	body, err := json.Marshal(g.GitHubPR)
	if err != nil {
		return err
	}
	logger.Debug(string(body))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.Token))

	resp, err := sendRequest(req)
	if err != nil {
		return err
	}
	logger.Debug(string(resp))

	return nil
}

func NewGitHub(pr PullRequest) (PRCreator, error) {
	temp := template.Must(template.New("file").Parse(prTemplate))
	var buf = bytes.Buffer{}
	if err := temp.Execute(&buf, pr); err != nil {
		return GitHub{}, err
	}

	return GitHub{
		Owner: pr.Owner,
		Repo:  pr.Repo,
		Token: pr.Token,
		GitHubPR: GitHubPR{
			Title: pr.Title,
			Body:  buf.String(),
			Head:  pr.Head,
			Base:  pr.Base,
		},
	}, nil
}
