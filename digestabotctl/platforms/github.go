package platforms

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/template"
)

type GitHub struct {
	Owner  string
	Repo   string
	Branch string
	Token  string
	Body   string
}

func (g GitHub) CreatePR() error {
	url := fmt.Sprintf("https://api.github.com/repos/[1]%s/[2]%s/branches/[3]%s", g.Owner, g.Repo, g.Branch)

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(g.Body))
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-Github-Api-Version", "2022-11-28")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", g.Token))

	resp, err := sendRequest(req)
	if err != nil {
		return err
	}

	fmt.Println(string(resp))

	return nil
}

func NewGithubPR(p PullRequest) error {
	temp := template.Must(template.New("file").Parse(prTemplate))
	return temp.Execute(os.Stdout, p)

}
