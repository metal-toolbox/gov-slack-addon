package slack

import (
	"context"

	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

// GetUser gets information about a slack organization user by id
func (c *Client) GetUser(ctx context.Context, id string) (*slack.User, error) {
	if id == "" {
		return nil, ErrBadParameter
	}

	c.logger.Debug("getting slack user info", zap.String("user.id", id))

	user, err := c.slackService.GetUserInfoContext(ctx, id)
	if err != nil {
		if err.Error() == "user_not_found" {
			return nil, ErrSlackUserNotFound
		}

		return nil, err
	}

	c.logger.Debug("returning slack user info", zap.Any("slack.user", user))

	return user, nil
}

// GetUserByEmail gets information about a slack organization user by email
func (c *Client) GetUserByEmail(ctx context.Context, email string) (*slack.User, error) {
	if email == "" {
		return nil, ErrBadParameter
	}

	c.logger.Debug("getting slack user info", zap.String("user.email", email))

	user, err := c.slackService.GetUserByEmailContext(ctx, email)
	if err != nil {
		if err.Error() == "users_not_found" {
			return nil, ErrSlackUserNotFound
		}

		return nil, err
	}

	c.logger.Debug("returning slack user info", zap.Any("slack.user", user))

	return user, nil
}
