package slack

import (
	"context"
	"time"

	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

// Client is a client that can talk to the Slack API
type Client struct {
	logger       *zap.Logger
	token        string
	slackService slackService
}

const (
	// default number of retry requests on GitHub requests
	retryAttempts = 3
	// default delay before retry on GitHub requests
	retryDelay = 5 * time.Second
)

type slackService interface {
	CreateUserGroupContext(context.Context, slack.UserGroup) (slack.UserGroup, error)
	DisableUserGroupContext(context.Context, string, ...slack.DisableUserGroupOption) (slack.UserGroup, error)
	EnableUserGroupContext(context.Context, string, ...slack.DisableUserGroupOption) (slack.UserGroup, error)
	GetUserGroupMembersContext(context.Context, string, ...slack.GetUserGroupMembersOption) ([]string, error)
	GetUserGroupsContext(context.Context, ...slack.GetUserGroupsOption) ([]slack.UserGroup, error)
	GetUserInfoContext(context.Context, string) (*slack.User, error)
	GetUserByEmailContext(context.Context, string) (*slack.User, error)
	ListTeamsContext(ctx context.Context, params slack.ListTeamsParameters) ([]slack.Team, string, error)
	UpdateUserGroupContext(context.Context, string, ...slack.UpdateUserGroupsOption) (slack.UserGroup, error)
	UpdateUserGroupMembersContext(context.Context, string, string, ...slack.UpdateUserGroupMembersOption) (slack.UserGroup, error)
}

// Option is a functional configuration option
type Option func(c *Client)

// WithToken sets the token
func WithToken(t string) Option {
	return func(c *Client) {
		c.token = t
	}
}

// WithLogger sets logger
func WithLogger(l *zap.Logger) Option {
	return func(c *Client) {
		c.logger = l
	}
}

// NewClient returns a new Slack client
func NewClient(opts ...Option) *Client {
	client := Client{
		logger: zap.NewNop(),
	}

	for _, opt := range opts {
		opt(&client)
	}

	client.slackService = slack.New(client.token)

	return &client
}

func stringPtr(s string) *string {
	return &s
}
