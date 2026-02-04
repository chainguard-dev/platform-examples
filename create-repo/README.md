# Create Repository

An example of adding an image to your Chainguard organization with the
Chainguard Go SDK.

## Usage

1. Build the example.

```
go build -o create-repo main.go
```

2. Login with `chainctl`. The example will reuse the token this creates for
   authentication.

```
chainctl auth login
```

3. Create the repository. This will add the `nginx` image to your organization.

```
./create-repo -parent your.org -repo nginx
```

## Assumable Identity

Rather than relying on `chainctl` for authentication, the example can assume an
identity.

Refer to [the
documentation](https://edu.chainguard.dev/chainguard/administration/assumable-ids/)
for examples of how to create an assumable identity.

```
./create-repo \
    -parent your.org \
    -repo nginx \
    -identity <identity-id> \
    -identity-token <identity-token>
```
