package platforms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"text/template"
)

type GitLab struct {
	Owner string
	Repo  string
	Token string
	GitLabMR
}

type GitLabMR struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Source      string `json:"source_branch"`
	Target      string `json:"target_branch"`
	Labels      string `json:"labels"`
}

func (g GitLab) CreatePR(logger *slog.Logger) error {
	url := fmt.Sprintf("https://gitlab.com/api/v4/projects/%s/merge_requests", g.Repo)
	body, err := json.Marshal(g.GitLabMR)
	if err != nil {
		return err
	}
	logger.Debug(string(body))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.Token))
	req.Header.Add("Content-Type", "application/json")

	resp, err := sendRequest(req)
	if err != nil {
		return err
	}
	logger.Debug(string(resp))

	return nil
}

func NewGitLab(pr PullRequest) (PRCreator, error) {
	temp := template.Must(template.New("file").Parse(prTemplate))
	var buf = bytes.Buffer{}
	if err := temp.Execute(&buf, pr); err != nil {
		return GitLab{}, err
	}

	return GitLab{
		Owner: pr.Owner,
		Repo:  pr.Repo,
		Token: pr.Token,
		GitLabMR: GitLabMR{
			Title:       pr.Title,
			Description: buf.String(),
			Source:      pr.Head,
			Target:      pr.Base,
			Labels:      strings.Join(pr.Labels, ","),
		},
	}, nil
}
