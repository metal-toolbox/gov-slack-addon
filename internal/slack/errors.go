package slack

import (
	"errors"
	"fmt"
)

// list of error messages returned by the slack api
const (
	SlackErrorNameAlreadyExists = "name_already_exists"
	SlackErrorNoSuchSubteam     = "no_such_subteam"
	SlackErrorSubteamNotFound   = "subteam_not_found"
	SlackErrorTeamNotFound      = "team_not_found"
	SlackErrorUserNotFound      = "user_not_found"
	SlackErrorUsersNotFound     = "users_not_found"
)

var (
	// ErrBadParameter is returned when bad parameters are passed to a slack request
	ErrBadParameter = errors.New("bad parameters in request")

	// ErrEmptyUserGroupMembers is returned when we try to set an empty user group member list
	ErrEmptyUserGroupMembers = errors.New("user group members cannot be empty list")

	// ErrMissingUserGroupParameter is returned when there are missing user group request parameters
	ErrMissingUserGroupParameter = errors.New("missing required user group parameters in request")

	// ErrSlackGroupAlreadyExists is returned when the slack user group already exists
	ErrSlackGroupAlreadyExists = errors.New("slack user group already exists")

	// ErrSlackGroupNotFound is returned when the slack user group is not found
	ErrSlackGroupNotFound = errors.New("slack user group not found")

	// ErrSlackUserNotFound is returned when the slack user is not found
	ErrSlackUserNotFound = errors.New("slack user not found")

	// ErrSlackWorkspaceNotFound is returned when the slack workspace (team) is not found
	ErrSlackWorkspaceNotFound = errors.New("slack workspace not found")

	// ErrSlackAPI is returned when a request to the Slack API fails. It wraps the
	// underlying slack-go error so callers can tell the failure originated from Slack.
	ErrSlackAPI = errors.New("slack api request failed")
)

// apiError wraps an error returned by the Slack API with the operation that
// failed, making it clear the error came from Slack rather than governor or
// another dependency (e.g. "slack api request failed: list workspaces: invalid_auth").
func apiError(op string, err error) error {
	return fmt.Errorf("%w: %s: %w", ErrSlackAPI, op, err)
}
