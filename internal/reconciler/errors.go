package reconciler

import "errors"

var (
	// ErrBadParameter is returned when bad parameters are passed to a request
	ErrBadParameter = errors.New("bad parameters in request")

	// ErrRequestNonSuccess is returned when a call to the slack API returns a non-success status
	ErrRequestNonSuccess = errors.New("got a non-success response from slack")

	// ErrAppNameEmpty is returned when the governor application name is empty
	ErrAppNameEmpty = errors.New("governor application name is empty")

	// ErrGovernorUserPendingStatus is returned when an event it received for a user with pending status
	ErrGovernorUserPendingStatus = errors.New("governor user has pending status")

	// ErrGroupMembershipNotFound is returned when a group membership action
	// is requested and the user is not found in the group
	ErrGroupMembershipNotFound = errors.New("user not found in group")

	// ErrGroupMembershipFound is returned when a group membership delete request finds the
	// user in the governor group
	ErrGroupMembershipFound = errors.New("delete request user found in group")

	// ErrSlackUserGroupNotFound is returned when the slack user group is not found
	ErrSlackUserGroupNotFound = errors.New("slack user group not found")

	// ErrSlackWorkspaceNotFound is returned when the slack workspace (team) is not found
	ErrSlackWorkspaceNotFound = errors.New("slack workspace not found")
)
