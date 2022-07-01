## Slack Webhook

This demo application shows how users can write a very simple application that
authenticates Chainguard webhook requests for continuous verification policy
violations, and turns them into Slack notifications.

### Setup

Create a Slack Webhook URL as detailed in the
[Incoming webhooks for Slack](https://slack.com/help/articles/115005265063-Incoming-webhooks-for-Slack)
help document.

Then, update `config/100-slack-secret.yaml` to uncomment the secret, and
add the Slack webhook url:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: slack
stringData:
  # Change this to a Slack Webhook URL:
  webhook: FILL ME IN!
```

Next, update `config/200-slack.yaml` to replace the following values as
documented:

```yaml
- name: ISSUER_URL
  # Change this to the issuer of the Chainguard environment
  # we want to test against.
  value: http://issuer.oidc-system.svc
- name: GROUP
  # Change this to the IAM group with which the chainguard subscription
  # webhook is associated.
  value: e97622b20fb99bd935ad7f5d7b9fd4cc6d46b9b7

```

This application can then be applied via:
```shell
ko apply -Bf config/
```

When the Knative Service becomes ready, you should be able to get the resulting
URL with:

```shell
# We will refer to the url from here below as ${URL}
kubectl get ksvc slack -ojson | jq -r .status.url
```

> _Note:_ `${GROUP}` refers to the group Chainguard Enforce was installed into
> for your cluster.

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
apiVersion: policy.sigstore.dev/v1alpha1
kind: ClusterImagePolicy
metadata:
  name: slack-test
spec:
  images:
  - glob: 'docker.io/ubuntu*'
  authorities:
  - keyless:
      url: https://fulcio.sigstore.dev
```

At this point, you should have a Slack notification in the configured channel
with the offending image!
