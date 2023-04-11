package slack

import (
	"context"
	"strings"

	retry "github.com/avast/retry-go/v4"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

// GetUser gets information about a slack organization user by id
func (c *Client) GetUser(ctx context.Context, id string) (*slack.User, error) {
	if id == "" {
		return nil, ErrBadParameter
	}

	c.logger.Debug("getting slack user info", zap.String("user.id", id))

	var user *slack.User

	err := retry.Do(
		func() error {
			var err error
			user, err = c.slackService.GetUserInfoContext(ctx, id)
			return err
		},
		retry.Attempts(retryAttempts),
		retry.Delay(retryDelay),
		retry.DelayType(retry.BackOffDelay),
		retry.RetryIf(func(err error) bool {
			return strings.Contains(err.Error(), "503 Service Unavailable")
		}),
	)
	if err != nil {
		if err.Error() == SlackErrorUsersNotFound {
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

	var user *slack.User

	err := retry.Do(
		func() error {
			var err error
			user, err = c.slackService.GetUserByEmailContext(ctx, email)
			return err
		},
		retry.Attempts(retryAttempts),
		retry.Delay(retryDelay),
		retry.DelayType(retry.BackOffDelay),
		retry.RetryIf(func(err error) bool {
			return strings.Contains(err.Error(), "503 Service Unavailable")
		}),
	)
	if err != nil {
		if err.Error() == SlackErrorUserNotFound {
			return nil, ErrSlackUserNotFound
		}

		return nil, err
	}

	c.logger.Debug("returning slack user info", zap.Any("slack.user", user))

	return user, nil
}
