// main.go
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	ecrsvc "github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	lambdasvc "github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// ---------- Event & config ----------

type JobEvent struct {
	Index    int    `json:"index,omitempty"`     // which repo in the list
	Repo     string `json:"repo,omitempty"`      // single explicit repo (optional)
	TagStart int    `json:"tag_start,omitempty"` // resume offset within a repo's tag list
}

func main() { lambda.Start(handler) }

func handler(ctx context.Context, raw json.RawMessage) error {
	start := time.Now()

	evt, err := parseJobEvent(raw)
	if err != nil {
		log.Printf("WARN: event parse failed (%v); defaulting to index=0", err)
		evt = JobEvent{Index: 0}
	}

	repos, err := loadRepoList(ctx)
	if err != nil {
		return fmt.Errorf("load repo list: %w", err)
	}
	if len(repos) == 0 {
		log.Printf("No repositories to process; exiting.")
		return nil
	}

	// Single-repo mode (no index advancement), but may paginate by tag_start.
	if r := strings.TrimSpace(evt.Repo); r != "" {
		more, nextStart, total, processed, err := mirrorSingleRepoBatch(ctx, r, evt.TagStart)
		if err != nil {
			return fmt.Errorf("mirror explicit repo %q: %w", r, err)
		}
		if more {
			log.Printf("Repo %s: processed %d/%d tags, chaining to tag_start=%d", r, processed, total, nextStart)
			return invokeSelfAsync(ctx, JobEvent{Repo: r, TagStart: nextStart})
		}
		log.Printf("Repo %s: completed all %d tags (elapsed %s)", r, total, time.Since(start))
		return nil
	}

	// Index-based multi-repo chaining.
	if evt.Index < 0 {
		evt.Index = 0
	}
	if evt.Index >= len(repos) {
		log.Printf("Index %d >= repo count %d; nothing to do.", evt.Index, len(repos))
		return nil
	}
	current := repos[evt.Index]
	log.Printf("Processing repo %d/%d: %s (tag_start=%d)", evt.Index+1, len(repos), current, evt.TagStart)

	more, nextStart, total, processed, err := mirrorSingleRepoBatch(ctx, current, evt.TagStart)
	if err != nil {
		return fmt.Errorf("mirror %q: %w", current, err)
	}

	if more {
		log.Printf("Repo %s: processed %d/%d tags so far, chaining to tag_start=%d", current, processed, total, nextStart)
		return invokeSelfAsync(ctx, JobEvent{Index: evt.Index, TagStart: nextStart})
	}

	// Move to next repo
	next := evt.Index + 1
	if next < len(repos) {
		if err := invokeSelfAsync(ctx, JobEvent{Index: next, TagStart: 0}); err != nil {
			return fmt.Errorf("invoke self for next index=%d: %w", next, err)
		}
		log.Printf("Completed repo %s; queued next repo index=%d (elapsed %s)", current, next, time.Since(start))
		return nil
	}
	log.Printf("Completed all %d repos 🎉 (elapsed %s)", len(repos), time.Since(start))
	return nil
}

func parseJobEvent(raw json.RawMessage) (JobEvent, error) {
	var evt JobEvent
	if len(raw) == 0 {
		idx := 0
		if s := strings.TrimSpace(os.Getenv("START_INDEX")); s != "" {
			if n, err := strconv.Atoi(s); err == nil {
				idx = n
			}
		}
		return JobEvent{Index: idx}, nil
	}
	if err := json.Unmarshal(raw, &evt); err != nil {
		return JobEvent{}, err
	}
	return evt, nil
}

// ---------- Repo list discovery ----------

func loadRepoList(ctx context.Context) ([]string, error) {
	// Highest precedence: JSON env
	if s := strings.TrimSpace(os.Getenv("REPO_LIST_JSON")); s != "" {
		var repos []string
		if err := json.Unmarshal([]byte(s), &repos); err == nil {
			return normalizeRepoList(repos), nil
		}
		log.Printf("WARN: REPO_LIST_JSON invalid; falling back...")
	}

	// Next: CSV env
	if s := strings.TrimSpace(os.Getenv("REPO_LIST_CSV")); s != "" {
		return normalizeRepoList(strings.Split(s, ",")), nil
	}

	// Next: SSM parameter (JSON array or CSV)
	if name := strings.TrimSpace(os.Getenv("REPO_LIST_SSM_PARAM")); name != "" {
		repos, err := loadReposFromSSM(ctx, name)
		if err != nil {
			return nil, fmt.Errorf("load from SSM %q: %w", name, err)
		}
		return repos, nil
	}

	// Fallback: minimal discovery (replace with real implementation if desired)
	group := os.Getenv("GROUP_NAME")
	if group == "" {
		group = "chainguard"
	}
	srcRegistry := os.Getenv("SRC_REGISTRY")
	if srcRegistry == "" {
		srcRegistry = "cgr.dev"
	}

	return []string{
		fmt.Sprintf("%s/%s/foo", srcRegistry, group),
		fmt.Sprintf("%s/%s/bar", srcRegistry, group),
		fmt.Sprintf("%s/%s/baz", srcRegistry, group),
	}, nil
}

func normalizeRepoList(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if t := strings.TrimSpace(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func loadReposFromSSM(ctx context.Context, paramName string) ([]string, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	client := ssm.NewFromConfig(cfg)
	withDecryption := true
	resp, err := client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           aws.String(paramName),
		WithDecryption: &withDecryption,
	})
	if err != nil {
		return nil, err
	}
	if resp.Parameter == nil || resp.Parameter.Value == nil {
		return nil, errors.New("empty SSM parameter")
	}
	val := strings.TrimSpace(*resp.Parameter.Value)

	// Try JSON first.
	var repos []string
	if json.Unmarshal([]byte(val), &repos) == nil {
		return normalizeRepoList(repos), nil
	}
	// Else treat as CSV.
	return normalizeRepoList(strings.Split(val, ",")), nil
}

// ---------- Mirroring with per-repo tags, batching & skip-existing ----------

func mirrorSingleRepoBatch(ctx context.Context, srcRepo string, tagStart int) (more bool, nextStart int, total int, processed int, err error) {
	if strings.EqualFold(os.Getenv("MIRROR_DRY_RUN"), "true") {
		log.Printf("[DRY-RUN] Would mirror: %s (from tag_start=%d)", srcRepo, tagStart)
		return false, 0, 0, 0, nil
	}

	// Source auth (cgr.dev)
	srcUser := os.Getenv("CGR_USERNAME")
	srcPass := os.Getenv("CGR_PASSWORD")
	if srcUser == "" || srcPass == "" {
		return false, 0, 0, 0, fmt.Errorf("CGR_USERNAME/CGR_PASSWORD not set")
	}
	srcAuth := &authn.Basic{Username: srcUser, Password: srcPass}

	// ECR client & auth
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return false, 0, 0, 0, err
	}
	ecr := ecrsvc.NewFromConfig(cfg)

	ao, err := ecr.GetAuthorizationToken(ctx, &ecrsvc.GetAuthorizationTokenInput{})
	if err != nil {
		return false, 0, 0, 0, fmt.Errorf("ecr:GetAuthorizationToken: %w", err)
	}
	if len(ao.AuthorizationData) == 0 {
		return false, 0, 0, 0, fmt.Errorf("no ECR authorization data")
	}
	ad := ao.AuthorizationData[0]
	dec, err := base64.StdEncoding.DecodeString(aws.ToString(ad.AuthorizationToken))
	if err != nil {
		return false, 0, 0, 0, fmt.Errorf("decode ecr token: %w", err)
	}
	parts := strings.SplitN(string(dec), ":", 2)
	if len(parts) != 2 {
		return false, 0, 0, 0, fmt.Errorf("unexpected ecr token format")
	}
	ecrPass := parts[1]
	ecrEndpoint := strings.TrimPrefix(aws.ToString(ad.ProxyEndpoint), "https://")
	dstAuth := &authn.Basic{Username: "AWS", Password: ecrPass}

	// Compute dest repo name; ensure it exists
	srcRegistry := os.Getenv("SRC_REGISTRY")
	if srcRegistry == "" {
		srcRegistry = "cgr.dev"
	}
	srcNoHost := strings.TrimPrefix(srcRepo, srcRegistry+"/")

	dstPrefix := strings.Trim(strings.TrimPrefix(os.Getenv("DST_PREFIX"), "/"), " ")
	dstRepoName := srcNoHost
	if dstPrefix != "" {
		dstRepoName = path.Join(dstPrefix, srcNoHost)
	}

	_, err = ecr.DescribeRepositories(ctx, &ecrsvc.DescribeRepositoriesInput{
		RepositoryNames: []string{dstRepoName},
	})
	if err != nil {
		var rnfe *ecrtypes.RepositoryNotFoundException
		if errors.As(err, &rnfe) {
			_, cerr := ecr.CreateRepository(ctx, &ecrsvc.CreateRepositoryInput{
				RepositoryName: aws.String(dstRepoName),
				ImageScanningConfiguration: &ecrtypes.ImageScanningConfiguration{
					// bool (not *bool) in your SDK
					ScanOnPush: true,
				},
			})
			if cerr != nil {
				return false, 0, 0, 0, fmt.Errorf("create ECR repo %s: %w", dstRepoName, cerr)
			}
			log.Printf("Created ECR repo: %s", dstRepoName)
		} else {
			return false, 0, 0, 0, fmt.Errorf("describe ECR repo %s: %w", dstRepoName, err)
		}
	}

	// Decide which tags to mirror (REPO_TAGS_JSON > COPY_ALL_TAGS > latest)
	var tags []string
	if rt := strings.TrimSpace(os.Getenv("REPO_TAGS_JSON")); rt != "" {
		var repoTags map[string][]string
		if err := json.Unmarshal([]byte(rt), &repoTags); err == nil {
			if custom, ok := repoTags[srcRepo]; ok && len(custom) > 0 {
				tags = custom
			}
		} else {
			log.Printf("WARN: REPO_TAGS_JSON invalid JSON; ignoring: %v", err)
		}
	}
	if len(tags) == 0 {
		copyAll := strings.EqualFold(os.Getenv("COPY_ALL_TAGS"), "true")
		if copyAll {
			repoRef, err := name.NewRepository(srcRepo)
			if err != nil {
				return false, 0, 0, 0, fmt.Errorf("parse src repository: %w", err)
			}
			tags, err = remote.List(repoRef, remote.WithAuth(srcAuth), remote.WithContext(ctx))
			if err != nil {
				return false, 0, 0, 0, fmt.Errorf("list tags for %s: %w", srcRepo, err)
			}
			if len(tags) == 0 {
				log.Printf("No tags found for %s", srcRepo)
				return false, 0, 0, 0, nil
			}
		} else {
			tags = []string{"latest"} // default smoke test
		}
	}
	total = len(tags)

	// Determine batch window
	maxPer := 10
	if s := strings.TrimSpace(os.Getenv("MAX_TAGS_PER_INVOKE")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			maxPer = n
		}
	}
	if tagStart < 0 {
		tagStart = 0
	}
	if tagStart >= total {
		return false, tagStart, total, 0, nil
	}
	end := tagStart + maxPer
	if end > total {
		end = total
	}

	// Time guard: avoid hitting 900s timeout mid-upload
	deadline, hasDeadline := ctx.Deadline()
	const buffer = 25 * time.Second

	// Process batch
	for i := tagStart; i < end; i++ {
		if hasDeadline && time.Until(deadline) < buffer {
			more = true
			nextStart = i
			processed = i - tagStart
			return more, nextStart, total, processed, nil
		}

		tag := tags[i]
		src := fmt.Sprintf("%s:%s", srcRepo, tag)
		dst := fmt.Sprintf("%s/%s:%s", ecrEndpoint, dstRepoName, tag)

		srcRef, err := name.ParseReference(src)
		if err != nil {
			return false, i, total, i - tagStart, fmt.Errorf("parse src ref %s: %w", src, err)
		}
		dstRef, err := name.ParseReference(dst)
		if err != nil {
			return false, i, total, i - tagStart, fmt.Errorf("parse dst ref %s: %w", dst, err)
		}

		// Source descriptor/digest
		desc, err := remote.Get(srcRef, remote.WithAuth(srcAuth), remote.WithContext(ctx))
		if err != nil {
			return false, i, total, i - tagStart, fmt.Errorf("get %s: %w", src, err)
		}
		srcDigest := desc.Descriptor.Digest.String()

		// Skip if same digest already tagged in ECR
		ebr, err := ecr.BatchGetImage(ctx, &ecrsvc.BatchGetImageInput{
			RepositoryName: aws.String(dstRepoName),
			ImageIds:       []ecrtypes.ImageIdentifier{{ImageTag: aws.String(tag)}},
		})
		if err == nil && len(ebr.Images) > 0 && ebr.Images[0].ImageId != nil && ebr.Images[0].ImageId.ImageDigest != nil {
			if aws.ToString(ebr.Images[0].ImageId.ImageDigest) == srcDigest {
				log.Printf("Skipping %s (tag %q) — already present with digest %s", srcRepo, tag, srcDigest)
				continue
			}
		}

		// Write image or index
		if desc.MediaType.IsIndex() {
			idx, err := desc.ImageIndex()
			if err != nil {
				return false, i, total, i - tagStart, fmt.Errorf("read index %s: %w", src, err)
			}
			log.Printf("Copying index %s -> %s", src, dst)
			if err := remote.WriteIndex(dstRef, idx, remote.WithAuth(dstAuth), remote.WithContext(ctx)); err != nil {
				return false, i, total, i - tagStart, fmt.Errorf("write index to %s: %w", dst, err)
			}
		} else {
			img, err := desc.Image()
			if err != nil {
				return false, i, total, i - tagStart, fmt.Errorf("read image %s: %w", src, err)
			}
			log.Printf("Copying image %s -> %s", src, dst)
			if err := remote.Write(dstRef, img, remote.WithAuth(dstAuth), remote.WithContext(ctx)); err != nil {
				return false, i, total, i - tagStart, fmt.Errorf("write image to %s: %w", dst, err)
			}
		}
	}

	processed = end - tagStart
	if end < total {
		return true, end, total, processed, nil // more tags remain
	}
	log.Printf("Mirrored %s (processed %d tag(s) this run, total %d)", srcRepo, processed, total)
	return false, end, total, processed, nil
}

// ---------- Self-invocation ----------

func invokeSelfAsync(ctx context.Context, ev JobEvent) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}
	client := lambdasvc.NewFromConfig(cfg)

	fn := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	if fn == "" {
		return errors.New("AWS_LAMBDA_FUNCTION_NAME not set")
	}

	payload, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	_, err = client.Invoke(ctx, &lambdasvc.InvokeInput{
		FunctionName:   aws.String(fn),
		InvocationType: lambdatypes.InvocationTypeEvent,
		Payload:        payload,
	})
	return err
}