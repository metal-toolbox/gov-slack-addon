# gov-slack-addon

`gov-slack-addon` is an addon to integrate Slack with Governor.

## Usage

This addon handles the create/delete/update of Slack user groups in Enterprise Grid.

`gov-slack-addon` subscribes to the Governor event stream where change events are published. The events published by Governor contain the group id that changed and the type of action. Events are published on NATS subjects dedicated to the resource type ie. `equinixmetal.governor.events.groups` for group events. When `gov-slack-addon` receives an event, it first checks that it's associted with a `slack` application in Governor, and then requests additional information from Governor about the included resource IDs and tries to match them to corresponding groups in Slack.

Slack Enterprise Grid acts as a parent organization for multiple workspaces (also called teams in Slack). For this reason `gov-slack-addon` needs a Slack token with organization-level permissions, and it also needs to be explicitly allowed in any workspaces that should be managed by the addon. In Governor, for each Slack workspace where you want to manage groups you need to create an application with type `slack` and a name that exactly matches the name of the Slack workspace, then associate that app with any Governor groups which should exist in Slack. You can associate one group with multiple slack applications and it will be created in all of the corresponding workspaces (with a `[Governor]` prefix).

As a side-note, users in Slack Enterprise Grid exist at the organization level but need to be invited to each workspace before they can be assigned to user groups there. The addon will silently fail to add group users if they are not already in the workspace. User matching between Governor and Slack is based on email address. Also note that we are only managing "User groups" which are used for mentions in Slack and exist at the workspace level (these are the traditional groups in Slack). Grid also has "IDP groups" which are at the organization level and are used for authorization (e.g. giving a group of users access to specific channels).

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
