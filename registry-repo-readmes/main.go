package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"

	"chainguard.dev/sdk/auth/token"
	registry "chainguard.dev/sdk/proto/platform/registry/v1"
	"chainguard.dev/sdk/sts"
	"github.com/gomarkdown/markdown"
	"github.com/microcosm-cc/bluemonday"
)

const (
	group    = "chainguard"
	audience = "https://console-api.enforce.dev"
	issuer   = "https://issuer.enforce.dev"
	identity = "720909c9f5279097d847ad02a2f24ba8f59de36a/a033a6fabe0bfa0d"
)

func main() {
	log.Println("Authenticating to Chainguard API...")
	ctx := context.Background()
	var tok string
	if token.RemainingLife(token.KindAccess, audience, time.Minute) < 0 {
		log.Fatalf("token has expired, please run `chainctl auth login`")
	}
	tokb, err := token.Load(token.KindAccess, audience)
	if err != nil {
		log.Fatalf("loading token: %v", err)
	}
	tok = string(tokb)
	if group == "chainguard" {
		tok, err = sts.New(issuer, audience, sts.WithIdentity(identity)).Exchange(ctx, tok)
		if err != nil {
			log.Fatalf("exchanging token: %v", err)
		}
	}
	regclients, err := registry.NewClients(ctx, audience, tok)
	if err != nil {
		log.Fatalf("registry.NewClients(): %v", err)
	}

	log.Println("Fetching list of repos...")
	repos, err := regclients.Registry().ListRepos(ctx, &registry.RepoFilter{})
	if err != nil {
		log.Fatalf("regclients.Registry().ListRepos(): %v", err)
	}

	log.Println("Converting READMEs to HTML...")
	repoReadmeMap := map[string]string{}
	for _, repo := range repos.Items {
		maybeUnsafeHTML := markdown.ToHTML([]byte(repo.Readme), nil, nil)
		html := bluemonday.UGCPolicy().SanitizeBytes(maybeUnsafeHTML)
		repoReadmeMap[repo.Name] = string(html)
	}

	log.Println("Creating HTTP handlers...")
	sortedReposNames := []string{}
	for k := range repoReadmeMap {
		sortedReposNames = append(sortedReposNames, k)
	}
	sort.Strings(sortedReposNames)
	html := "<html><head><title>Registry Repo READMEs</title>"
	html += "<style>html {font-family: monospace;} table, th, td {border: 1px solid black;}</style></head>"
	html += "<body><h1>All Repos</h1><table><tr><th>name</th></tr>"
	for _, k := range sortedReposNames {
		html += fmt.Sprintf("<tr><td><a href='/%s'>%s</a></td></tr>", k, k)
	}
	html += "</table>"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, html)
	})
	for _, k := range sortedReposNames {
		func(name string) {
			http.HandleFunc(fmt.Sprintf("/%s", name), func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, repoReadmeMap[name])
			})
		}(k)
	}

	log.Println("Starting server on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
