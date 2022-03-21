## Github Issue Opener

This demo application shows how users can write a very simple application that
authenticates Chainguard webhook requests for continuous verification policy
violations (TODO), and turns them into Github issues.

### Setup

First, determine the Chainguard IAM Group with which you would like to associate
the events with:
```shell
# We will refer to the group chosen here below as ${GROUP}
chainctl iam groups ls
```

Next, update `config/200-opener.yaml` to replace the following values as
documented:

```yaml
- name: ISSUER_URL
  # Change this to the issuer of the Chainguard environment
  # we want to test against.
  value: http://issuer.oidc-system.svc
- name: GROUP
  # Change this to the IAM group with which this webhook is
  # associated.
  value: e97622b20fb99bd935ad7f5d7b9fd4cc6d46b9b7
- name: GITHUB_ORG
  # Change this to the Github org (or user) hosting the
  # repository in which to file issues.
  value: mattmoor
- name: GITHUB_REPO
  # Change this to the Github repository in which to file issues.
  value: eks-demo
```

Next, update `config/100-github-secret.yaml` to uncomment the secret, and
fill in a "personal access token" with the `repo` permission:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: github-token
stringData:
  # Change this to a personal access token with "repo" permissions.
  pat: FILL ME IN!
```

This application can then be applied via:
```shell
ko apply -Bf config/
```

When the Knative Service becomes ready, you should be able to get the resulting
URL with:

```shell
# We will refer to the url from here below as ${URL}
kubectl get ksvc opener -ojson | jq -r .status.url
```

You can then associate the webhook with:
```shell
chainctl events subscriptions create --group="${GROUP}" "${URL}"
```

### Testing things (STILL ASPIRATIONAL)

> _Note:_ We assume here that your cluster has Chainguard Enforce installed!

First, in a namespace *without* the `cosigned` admission control enabled run:

```shell
kubectl run -n "${NO_COSIGNED_NAMESPACE}" unsigned-image \
   --image="docker.io/ubuntu" -- sleep 3600
```

Next, apply a policy to your cluster:

```yaml
# kubectl apply this!
apiVersion: cosigned.sigstore.dev/v1alpha1
kind: ClusterImagePolicy
metadata:
  name: github-test
spec:
  images:
  - glob: 'docker.io/ubuntu*'
  authorities:
  - keyless:
      url: https://fulcio.sigstore.dev
```

At this point, you should have an issue opened in the configured repository
with the offending image!
