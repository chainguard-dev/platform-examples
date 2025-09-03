# digestabotctl

Updates image digests in files. 

## GitHub

```
jobs:
  digestabot:
    name: Digestabot
    runs-on: ubuntu-latest

    permissions:
      contents: write
      pull-requests: write
      id-token: write

    steps:
    - uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683 # v4.2.2
    - uses: chainguard-dev/setup-chainctl@v0.3.2
      with:
        identity: '<your-assumable-id>'

    - name: digestabot
      env:
        DIGESTABOT_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        DIGESTABOT_BRANCH: digestabot-update # branch to push commits to 
        DIGESTABOT_CREATE_PR: true
        DIGESTABOT_PLATFORM: github
        DIGESTABOT_OWNER: org-owner
        DIGESTABOT_REPO: repo-name
        DIGESTABOT_SIGN: true # set to true if you want to sign commits with sigstore
        DIGESTABOT_EMAIL: committer email
        DIGESTABOT_NAME: committer username
      run: |
        ./digestabotctl update files
```

## GitLab

```
stages:
  - update
workflow:
  rules:
    - if: $CI_PIPELINE_SOURCE == "web" || $CI_PIPELINE_SOURCE == "schedule"
variables:
  DIGESTABOT_TOKEN: ${PUSH_TOKEN}
  DIGESTABOT_BRANCH: digestabot-update # branch to push commits to
  DIGESTABOT_CREATE_PR: true
  DIGESTABOT_PLATFORM: gitlab
  DIGESTABOT_OWNER: $CI_PROJECT_NAMESPACE
  DIGESTABOT_REPO: $CI_PROJECT_ID
  DIGESTABOT_SIGN: true
  DIGESTABOT_SIGNING_TOKEN: $SIGSTORE_TOKEN # needed for GitLab since it's not an API exchange
  DIGESTABOT_EMAIL: $GITLAB_USER_EMAIL
  DIGESTABOT_NAME: $GITLAB_USER_NAME

digestabot:
  stage: update
  id_tokens:
    ID_TOKEN_1:
      aud: https://gitlab.com
    SIGSTORE_TOKEN:
      aud: sigstore # get token with audience for commit signing
  script:
    - wget -O /bin/chainctl "https://dl.enforce.dev/chainctl/latest/chainctl_linux_$(uname -m)"
    - chmod 755 /bin/chainctl
    - chainctl auth login --identity-token $ID_TOKEN_1 --identity $CGR_IDENTITY --audience apk.cgr.dev
    - chainctl auth configure-docker --identity-token $ID_TOKEN_1 --identity $CGR_IDENTITY
    - digestabotctl update files

```


## CLI Reference is [here](./docs)
