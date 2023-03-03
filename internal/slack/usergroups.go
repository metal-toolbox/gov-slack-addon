package slack

import (
	"context"
	"strings"

	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

// UserGroupReq is a request to create or update a user group
type UserGroupReq struct {
	Name        *string
	Handle      *string
	Description *string
}

// CreateUserGroup creates a new user group in a workspace
func (c *Client) CreateUserGroup(ctx context.Context, teamID string, userGroup *UserGroupReq) (*slack.UserGroup, error) {
	if teamID == "" {
		return nil, ErrBadParameter
	}

	if userGroup.Name == nil || userGroup.Handle == nil {
		return nil, ErrMissingUserGroupParameter
	}

	c.logger.Debug("creating slack user group",
		zap.String("slack.workspace.id", teamID),
		zap.Any("slack.usergroup.req", userGroup),
	)

	ugReq := slack.UserGroup{
		Name:   *userGroup.Name,
		Handle: *userGroup.Handle,
		TeamID: teamID,
	}

	if userGroup.Description != nil {
		ugReq.Description = *userGroup.Description
	}

	ug, err := c.slackService.CreateUserGroupContext(ctx, ugReq)
	if err != nil {
		if err.Error() == "name_already_exists" {
			return nil, ErrSlackGroupAlreadyExists
		}

		return nil, err
	}

	c.logger.Debug("created slack user group", zap.Any("slack.usergroup", ug))

	return &ug, nil
}

// DisableUserGroup disables a user group. Slack does not support deleting a group.
func (c *Client) DisableUserGroup(ctx context.Context, groupID, teamID string) (*slack.UserGroup, error) {
	if groupID == "" || teamID == "" {
		return nil, ErrBadParameter
	}

	c.logger.Debug("disabling slack user group", zap.String("slack.usergroup.id", groupID), zap.String("workspace", teamID))

	opts := []slack.DisableUserGroupOption{
		slack.DisableUserGroupOptionIncludeCount(true),
		slack.DisableUserGroupOptionTeamID(teamID),
	}

	ug, err := c.slackService.DisableUserGroupContext(ctx, groupID, opts...)
	if err != nil {
		if err.Error() == "no_such_subteam" || err.Error() == "subteam_not_found" {
			return nil, ErrSlackGroupNotFound
		}

		return nil, err
	}

	c.logger.Debug("disabled slack user group", zap.Any("slack.usergroup", ug))

	return &ug, nil
}

// EnableUserGroup enables a disabled user group
func (c *Client) EnableUserGroup(ctx context.Context, groupID, teamID string) (*slack.UserGroup, error) {
	if groupID == "" || teamID == "" {
		return nil, ErrBadParameter
	}

	c.logger.Debug("enabling slack user group", zap.String("slack.usergroup.id", groupID), zap.String("workspace", teamID))

	opts := []slack.DisableUserGroupOption{
		slack.DisableUserGroupOptionIncludeCount(true),
		slack.DisableUserGroupOptionTeamID(teamID),
	}

	ug, err := c.slackService.EnableUserGroupContext(ctx, groupID, opts...)
	if err != nil {
		if err.Error() == "no_such_subteam" || err.Error() == "subteam_not_found" {
			return nil, ErrSlackGroupNotFound
		}

		return nil, err
	}

	c.logger.Debug("enabled slack user group", zap.Any("slack.usergroup", ug))

	return &ug, nil
}

// GetUserGroups gets all user groups in a workspace (given the workspace/team id)
func (c *Client) GetUserGroups(ctx context.Context, teamID string, includeDisabled bool) ([]slack.UserGroup, error) {
	if teamID == "" {
		return nil, ErrBadParameter
	}

	c.logger.Debug("getting slack user groups", zap.String("workspace", teamID), zap.Bool("includeDisabled", includeDisabled))

	opts := []slack.GetUserGroupsOption{
		slack.GetUserGroupsOptionIncludeDisabled(includeDisabled),
		slack.GetUserGroupsOptionIncludeCount(true),
		slack.GetUserGroupsOptionIncludeUsers(true),
		slack.GetUserGroupsOptionTeamID(teamID),
	}

	groups, err := c.slackService.GetUserGroupsContext(ctx, opts...)
	if err != nil {
		if err.Error() == "team_not_found" {
			return nil, ErrSlackWorkspaceNotFound
		}

		return nil, err
	}

	c.logger.Debug("returning slack user groups", zap.Any("slack.usergroups", groups))

	return groups, nil
}

// GetUserGroupMembers returns the members of a user group
func (c *Client) GetUserGroupMembers(ctx context.Context, groupID, teamID string, includeDisabled bool) ([]string, error) {
	if groupID == "" || teamID == "" {
		return nil, ErrBadParameter
	}

	c.logger.Debug("getting slack user group members",
		zap.String("slack.usergroup.id", groupID),
		zap.String("workspace", teamID),
		zap.Bool("includeDisabled", includeDisabled),
	)

	opts := []slack.GetUserGroupMembersOption{
		slack.GetUserGroupMembersOptionIncludeDisabled(includeDisabled),
		slack.GetUserGroupMembersOptionTeamID(teamID),
	}

	members, err := c.slackService.GetUserGroupMembersContext(ctx, groupID, opts...)
	if err != nil {
		if err.Error() == "no_such_subteam" || err.Error() == "subteam_not_found" {
			return nil, ErrSlackGroupNotFound
		}

		return nil, err
	}

	c.logger.Debug("returning slack user group members", zap.Any("slack.usergroup.members", members))

	return members, nil
}

// UpdateUserGroup updates one or more parameters of a user group
func (c *Client) UpdateUserGroup(ctx context.Context, groupID, teamID string, userGroup UserGroupReq) (*slack.UserGroup, error) {
	if groupID == "" || teamID == "" {
		return nil, ErrBadParameter
	}

	if userGroup.Name == nil && userGroup.Handle == nil && userGroup.Description == nil {
		return nil, ErrMissingUserGroupParameter
	}

	c.logger.Debug("updating slack user group",
		zap.String("slack.usergroup.id", groupID),
		zap.String("slack.workspace.id", teamID),
		zap.Any("slack.usergroup.req", userGroup),
	)

	opts := []slack.UpdateUserGroupsOption{
		slack.UpdateUserGroupsOptionTeamID(&teamID),
	}

	if userGroup.Name != nil {
		opts = append(opts, slack.UpdateUserGroupsOptionName(*userGroup.Name))
	}

	if userGroup.Handle != nil {
		opts = append(opts, slack.UpdateUserGroupsOptionHandle(*userGroup.Handle))
	}

	if userGroup.Description != nil {
		opts = append(opts, slack.UpdateUserGroupsOptionDescription(userGroup.Description))
	}

	ug, err := c.slackService.UpdateUserGroupContext(ctx, groupID, opts...)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("updated slack user group", zap.Any("slack.usergroup", ug))

	return &ug, nil
}

// UpdateUserGroupMembers updates the members of a user group. You cannot pass an empty members list as the
// Slack API doesn't allow removing all members of a group.
func (c *Client) UpdateUserGroupMembers(ctx context.Context, groupID, teamID string, members []string) (*slack.UserGroup, error) {
	if groupID == "" || teamID == "" {
		return nil, ErrBadParameter
	}

	if len(members) == 0 {
		return nil, ErrEmptyUserGroupMembers
	}

	c.logger.Debug("updating slack user group members",
		zap.String("slack.usergroup.id", groupID),
		zap.String("workspace", teamID),
		zap.Any("slack.usergroup.members", members),
	)

	opts := []slack.UpdateUserGroupMembersOption{
		slack.UpdateUserGroupMembersOptionIncludeCount(true),
		slack.UpdateUserGroupMembersOptionTeamID(teamID),
	}

	// we need a comma separated list of members
	m := strings.Join(members, ",")

	ug, err := c.slackService.UpdateUserGroupMembersContext(ctx, groupID, m, opts...)
	if err != nil {
		if err.Error() == "no_such_subteam" || err.Error() == "subteam_not_found" {
			return nil, ErrSlackGroupNotFound
		}

		return nil, err
	}

	c.logger.Debug("updated slack user group members", zap.Any("slack.usergroup", ug))

	return &ug, nil
}
