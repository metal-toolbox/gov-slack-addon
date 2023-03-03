package slack

import (
	"context"

	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

// ListWorkspaces returns a list of all workspaces (teams) the token has access to. In the
// case of Enterprise Grid, this will return a list of all the workspaces in the Slack orgnization.
// Keep in mind that Slack uses the terms "team" and "workspace" interchangeably in this context.
// Let's use workspace wherever possible to avoid confusion.
func (c *Client) ListWorkspaces(ctx context.Context) ([]slack.Team, error) {
	c.logger.Debug("getting slack workspaces")

	opts := slack.ListTeamsParameters{
		Limit: 100, //nolint:gomnd
	}

	teams, _, err := c.slackService.ListTeamsContext(ctx, opts)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("returning slack workspaces", zap.Any("slack.workspaces", teams))

	return teams, nil
}
