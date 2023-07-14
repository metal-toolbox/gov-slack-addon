# gov-slack-addon

`gov-slack-addon` is an addon to integrate Slack with Governor.

## Usage

This addon handles the create/delete/update of Slack user groups in Slack Enterprise Grid.

`gov-slack-addon` subscribes to the Governor event stream where change events are published. The events published by Governor contain the group id that changed and the type of action. Events are published on NATS subjects dedicated to the resource type ie. `governor.events.groups` for group events. When `gov-slack-addon` receives an event, it first checks that it's associated with a `slack` application in Governor, and then requests additional information from Governor about the included resource IDs and tries to match them to corresponding groups in Slack.

Slack Enterprise Grid acts as a parent organization for multiple workspaces (also called teams in Slack). For this reason `gov-slack-addon` needs a Slack token with organization-level permissions, and it also needs to be explicitly allowed in any workspaces that should be managed by the addon. In Governor, for each Slack workspace where you want to manage groups you need to create an application with type `slack` and a name that exactly matches the name of the Slack workspace, then associate that app with any Governor groups which should exist in Slack. You can associate one group with multiple slack applications and it will be created in all of the corresponding workspaces (with a `[Governor]` prefix).

As a side-note, users in Slack Enterprise Grid exist at the organization level but need to be invited to each workspace before they can be assigned to user groups there. The addon will silently fail to add group users if they are not already in the workspace. User matching between Governor and Slack is based on email address. Also note that we are only managing "User groups" which are used for mentions in Slack and exist at the workspace level (these are the traditional groups in Slack). Grid also has "IDP groups" which are at the organization level and are used for authorization (e.g. giving a group of users access to specific channels).

## Development

### Pre-requisites for running locally

Follow the directions [here](https://github.com/metal-toolbox/governor-api/blob/main/README.md#running-governor-locally) for starting the governor-api devcontainer.

The **first time** you'll need to create a local hydra client for `gov-slack-addon-governor` and copy the nats creds file. After that you can just export the env variables.

First create a local audit log for testing in the `gov-slack-addon` directory:

```sh
touch audit.log
```

#### NATS Creds

Run in the governor-api devcontainer:

```sh
cat /tmp/user.creds
```

Then create and copy into `gov-slack-addon/user.local.creds`

#### Hydra

```sh
export GSA_GOVERNOR_CLIENT_ID="gov-slack-addon-governor"
# Copy for env variable later
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

#### Env

Export the required env variables to point to our local Governor and Hydra:

```sh
export GSA_GOVERNOR_URL="http://127.0.0.1:3001"
export GSA_GOVERNOR_AUDIENCE="http://api:3001/"
export GSA_GOVERNOR_TOKEN_URL="http://127.0.0.1:4444/oauth2/token"
export GSA_GOVERNOR_CLIENT_ID="gov-slack-addon-governor"
```

Also ensure you have the following secrets exported:

```sh
# Retrieve from pw vault, look for governor tag
export GSA_SLACK_TOKEN="REPLACE"
export GSA_GOVERNOR_CLIENT_SECRET="REPLACE"
```

#### Troubleshooting

**"error": "Unable to insert or update resource because a resource with that value exists already"**

Run `hydra clients delete gov-slack-addon-governor` in the governor-api devcontainer. Then rerun the steps for hydra.

**"error": "error",**
**"error_description": "The error is unrecognizable"**

Same as above.

### Testing addon locally

Start the addon (adjust the flags as needed):

```sh
go run . serve --audit-log-path=audit.log --nats-creds-file user.local.creds --pretty --debug --dry-run
```

## License

[Apache License, Version 2.0](LICENSE)
