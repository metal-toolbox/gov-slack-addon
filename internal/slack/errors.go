package slack

import (
	"errors"
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
)
