# gov-slack-addon

`gov-slack-addon` is an addon to integrate Slack with Governor.

## Usage

Search this repo for `TODO` to find the places where code would need to be updated which should be limited do:

```bash
.buildkite/pipeline.yml

cmd/serve.go
cmd/root.go
cmd/errors.go

srv/doc.go
srv/errors.go
```

## Development

### Pre-requisites for running locally

You can run `gov-slack-addon` against your local Governor instance (if you don't already have one, follow the directions [here](https://github.com/equinixmetal/governor/blob/main/README.md#running-governor-locally).

The first time you'll need to create a local hydra client for `gov-slack-addon-governor` (you do this where hydra is running, so most likely the governor-api):

```
export GSA_GOVERNOR_CLIENT_ID="gov-slack-addon-governor"
export GSA_GOVERNOR_CLIENT_SECRET="$(openssl rand -hex 16)"
echo "${GSA_GOVERNOR_CLIENT_SECRET}"

hydra clients create \
    --id ${GSA_GOVERNOR_CLIENT_ID} \
    --secret ${GSA_GOVERNOR_CLIENT_SECRET} \
    --endpoint http://hydra:4445/ \
    --audience http://api:3001/ \
    --grant-types client_credentials \
    --response-types token,code \
    --token-endpoint-auth-method client_secret_post \
    --scope read:governor:users,read:governor:groups,read:governor:applications
```

Export the required env variables to point to our local Governor and Hydra:

```
export GSA_GOVERNOR_URL="http://127.0.0.1:3001"
export GSA_GOVERNOR_AUDIENCE="http://api:3001/"
export GSA_GOVERNOR_TOKEN_URL="http://127.0.0.1:4444/oauth2/token"
export GSA_GOVERNOR_CLIENT_ID="gov-slack-addon-governor"
export GSA_NATS_TOKEN="notused"
```

Also ensure you have the following secrets exported:

```
export GSA_SLACK_TOKEN="REPLACE"
export GSA_GOVERNOR_CLIENT_SECRET="REPLACE"
```

Create a local audit log for testing in the `gov-slack-addon` directory:

```
touch audit.log
```

### Testing addon locally

Start the addon (adjust the flags as needed):

```
go run . serve --audit-log-path=audit.log --pretty --development --debug --dry-run
```
