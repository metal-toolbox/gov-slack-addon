package reconciler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/equinixmetal/gov-slack-addon/internal/auctx"
	"github.com/equinixmetal/gov-slack-addon/internal/slack"
	"go.equinixmetal.net/governor-api/pkg/api/v1alpha1"
	"go.uber.org/zap"
)

// UserGroup has the basic details of a slack user group
type UserGroup struct {
	ID          string
	Name        string
	Handle      string
	Description string
	Users       []string
}

// AddUserGroupMember adds a user to a user group if they are not already a member
func (r *Reconciler) AddUserGroupMember(ctx context.Context, groupID, userID string) error {
	if groupID == "" || userID == "" {
		return ErrBadParameter
	}

	group, err := r.GovernorClient.Group(ctx, groupID, false)
	if err != nil {
		r.Logger.Error("error getting governor group", zap.String("governor.group.id", groupID), zap.Error(err))
		return err
	}

	if len(group.Applications) == 0 {
		r.Logger.Debug("no applications linked to group", zap.String("governor.group.id", groupID))
		return nil
	}

	user, err := r.GovernorClient.User(ctx, userID, false)
	if err != nil {
		r.Logger.Error("error getting governor user", zap.Error(err))
		return err
	}

	if !contains(group.Members, user.ID) {
		r.Logger.Error("governor group does not contain requested membership",
			zap.String("governor.group.id", group.ID),
			zap.String("governor.user.id", user.ID),
		)

		return nil
	}

	if user.Status.String == v1alpha1.UserStatusPending {
		r.Logger.Debug("skipping pending user", zap.String("governor.user.id", userID), zap.String("governor.user.email", user.Email))
		return nil
	}

	// check all applications linked to the group and add group member if a slack application
	for _, appID := range group.Applications {
		isSlack, workspace, err := r.isSlackApplication(ctx, appID)
		if err != nil {
			r.Logger.Warn("failed to get application from governor", zap.String("governor.app.id", appID), zap.Error(err))
			continue
		}

		if !isSlack {
			r.Logger.Debug("not a slack application, skipping", zap.String("governor.app.id", appID), zap.String("governor.app.name", workspace))
			continue
		}

		logger := r.Logger.With(
			zap.String("slack.workspace.name", workspace),
			zap.String("governor.app.id", appID),
			zap.String("governor.group.id", group.ID),
			zap.String("governor.group.slug", group.Slug),
			zap.String("governor.user.id", userID),
			zap.String("governor.user.email", user.Email),
		)

		teamID, err := r.teamIDFromName(ctx, workspace)
		if err != nil {
			logger.Error("failed to get workspace id", zap.Error(err))
			continue
		}

		ug, err := r.userGroupFromName(ctx, r.userGroupName(group.Name), teamID, false)
		if err != nil {
			logger.Error("failed to get slack user group", zap.Error(err))
			continue
		}

		u, err := r.Client.GetUserByEmail(ctx, user.Email)
		if err != nil {
			logger.Error("failed to get slack user", zap.Error(err))
			continue
		}

		if contains(ug.Users, u.ID) {
			logger.Info("user already in group, skipping")
			continue
		}

		newUsers := ug.Users
		newUsers = append(newUsers, u.ID)

		logger.Debug("updating user group members", zap.Any("slack.usergroup.existing", ug.Users), zap.Any("slack.usergroup.new", newUsers))

		if r.dryrun {
			logger.Info("SKIP adding user to group", zap.Any("slack.usergroup", *ug))
			continue
		}

		_, err = r.Client.UpdateUserGroupMembers(ctx, ug.ID, teamID, newUsers)
		if err != nil {
			logger.Error("failed to create user group", zap.String("slack.usergroup.name", r.userGroupName(group.Name)), zap.Error(err))
			return err
		}

		logger.Info("added user to group")

		if err := auctx.WriteAuditEvent(ctx, r.auditEventWriter, "UserGroupAddMember", map[string]string{
			"slack.workspace.name": workspace,
			"slack.usergroup.name": ug.Name,
			"slack.usergroup.id":   ug.ID,
			"slack.user.id":        u.ID,
			"governor.app.id":      appID,
			"governor.group.id":    group.ID,
			"governor.group.slug":  group.Slug,
			"governor.user.id":     userID,
		}); err != nil {
			logger.Error("error writing audit event", zap.Error(err))
		}
	}

	return nil
}

// CreateUserGroup creates a slack user group for the given governor group if it's
// linked to a slack application
func (r *Reconciler) CreateUserGroup(ctx context.Context, groupID, appID string) error {
	if groupID == "" || appID == "" {
		return ErrBadParameter
	}

	isSlack, workspace, err := r.isSlackApplication(ctx, appID)
	if err != nil {
		r.Logger.Error("failed to get application from governor", zap.String("governor.app.id", appID), zap.Error(err))
		return err
	}

	if !isSlack {
		r.Logger.Debug("not a slack application, skipping", zap.String("governor.app.id", appID), zap.String("governor.app.name", workspace))
		return nil
	}

	logger := r.Logger.With(zap.String("slack.workspace.name", workspace), zap.String("governor.app.id", appID))

	teamID, err := r.teamIDFromName(ctx, workspace)
	if err != nil {
		return err
	}

	group, err := r.GovernorClient.Group(ctx, groupID, false)
	if err != nil {
		logger.Error("error getting governor group", zap.String("governor.group.id", groupID), zap.Error(err))
		return err
	}

	if r.dryrun {
		logger.Info("SKIP creating slack user group", zap.String("slack.usergroup.name", r.userGroupName(group.Name)))
		return nil
	}

	ug, err := r.Client.CreateUserGroup(ctx, teamID, r.userGroupReq(group))
	if err != nil {
		if !errors.Is(err, slack.ErrSlackGroupAlreadyExists) {
			logger.Error("failed to create user group", zap.String("slack.usergroup.name", r.userGroupName(group.Name)), zap.Error(err))
		}

		return err
	}

	logger.Info("created user group", zap.Any("slack.usergroup", ug))

	if err := auctx.WriteAuditEvent(ctx, r.auditEventWriter, "UserGroupCreate", map[string]string{
		"slack.workspace.name": workspace,
		"slack.usergroup.name": ug.Name,
		"slack.usergroup.id":   ug.ID,
		"governor.app.id":      appID,
		"governor.group.id":    groupID,
	}); err != nil {
		logger.Error("error writing audit event", zap.Error(err))
	}

	return nil
}

// DeleteUserGroup deletes a corresponding slack user group for the given governor group
// if it's linked to a slack application. Since Slack doesn't support deleting user groups
// we simulate deletion by disabling and renaming the user group to a timestamped name.
func (r *Reconciler) DeleteUserGroup(ctx context.Context, groupID, appID string) error {
	if groupID == "" || appID == "" {
		return ErrBadParameter
	}

	isSlack, workspace, err := r.isSlackApplication(ctx, appID)
	if err != nil {
		r.Logger.Error("failed to get application from governor", zap.String("governor.app.id", appID), zap.Error(err))
		return err
	}

	if !isSlack {
		r.Logger.Debug("not a slack application, skipping", zap.String("governor.app.id", appID), zap.String("governor.app.name", workspace))
		return nil
	}

	logger := r.Logger.With(zap.String("slack.workspace.name", workspace), zap.String("governor.app.id", appID))

	group, err := r.GovernorClient.Group(ctx, groupID, true)
	if err != nil {
		logger.Error("error getting governor group", zap.String("governor.group.id", groupID), zap.Error(err))
		return err
	}

	teamID, err := r.teamIDFromName(ctx, workspace)
	if err != nil {
		return err
	}

	ug, err := r.userGroupFromName(ctx, r.userGroupName(group.Name), teamID, false)
	if err != nil {
		return err
	}

	if r.dryrun {
		logger.Info("SKIP deleting slack user group", zap.Any("slack.usergroup", *ug))
		return nil
	}

	// we'll rename the group first to avoid future conflicts, and then disable it,
	// since slack doesn't support deleting groups

	ts := timestamp()
	nameR := fmt.Sprintf("%s (deleted %s)", ug.Name, ts)
	handleR := fmt.Sprintf("%s-deleted-%s", ug.Handle, ts)
	descriptionR := fmt.Sprintf("%s (deleted by gov-slack-addon %s)", ug.Description, ts)

	if _, err := r.Client.UpdateUserGroup(ctx, ug.ID, teamID, slack.UserGroupReq{
		Name:        &nameR,
		Handle:      &handleR,
		Description: &descriptionR,
	}); err != nil {
		logger.Error("failed to rename user group", zap.String("slack.usergroup.name", nameR), zap.Error(err))
		return err
	}

	if _, err := r.Client.DisableUserGroup(ctx, ug.ID, teamID); err != nil {
		logger.Error("failed to disable user group", zap.Any("slack.usergroup", *ug), zap.Error(err))

		// try to restore the original user group details
		if _, err := r.Client.UpdateUserGroup(ctx, ug.ID, teamID, slack.UserGroupReq{
			Name:        &ug.Name,
			Handle:      &ug.Handle,
			Description: &ug.Description,
		}); err != nil {
			logger.Error("failed to restore user group name", zap.String("slack.usergroup.name", ug.Name), zap.Error(err))
		}

		return err
	}

	logger.Info("deleted user group", zap.Any("slack.usergroup", ug))

	if err := auctx.WriteAuditEvent(ctx, r.auditEventWriter, "UserGroupDelete", map[string]string{
		"slack.workspace.name": workspace,
		"slack.usergroup.name": ug.Name,
		"slack.usergroup.id":   ug.ID,
		"governor.app.id":      appID,
		"governor.group.id":    groupID,
	}); err != nil {
		logger.Error("error writing audit event", zap.Error(err))
	}

	return nil
}

// RemoveUserGroupMember removes a user from a user group. Slack doesn't allow removing the last user in a group,
// so in that case we'll return an error.
func (r *Reconciler) RemoveUserGroupMember(ctx context.Context, groupID, userID string) error {
	if groupID == "" || userID == "" {
		return ErrBadParameter
	}

	group, err := r.GovernorClient.Group(ctx, groupID, false)
	if err != nil {
		r.Logger.Error("error getting governor group", zap.String("governor.group.id", groupID), zap.Error(err))
		return err
	}

	if len(group.Applications) == 0 {
		r.Logger.Debug("no applications linked to group", zap.String("governor.group.id", groupID))
		return nil
	}

	user, err := r.GovernorClient.User(ctx, userID, false)
	if err != nil {
		r.Logger.Error("error getting governor user", zap.Error(err))
		return err
	}

	if contains(group.Members, user.ID) {
		r.Logger.Error("governor group contains requested membership",
			zap.String("governor.group.id", group.ID),
			zap.String("governor.user.id", user.ID),
		)

		return nil
	}

	if user.Status.String == v1alpha1.UserStatusPending {
		r.Logger.Debug("skipping pending user", zap.String("governor.user.id", userID), zap.String("governor.user.email", user.Email))
		return nil
	}

	// check all applications linked to the group and remove group member if a slack application
	for _, appID := range group.Applications {
		isSlack, workspace, err := r.isSlackApplication(ctx, appID)
		if err != nil {
			r.Logger.Warn("failed to get application from governor", zap.String("governor.app.id", appID), zap.Error(err))
			continue
		}

		if !isSlack {
			r.Logger.Debug("not a slack application, skipping", zap.String("governor.app.id", appID), zap.String("governor.app.name", workspace))
			continue
		}

		logger := r.Logger.With(
			zap.String("slack.workspace.name", workspace),
			zap.String("governor.app.id", appID),
			zap.String("governor.group.id", group.ID),
			zap.String("governor.group.slug", group.Slug),
			zap.String("governor.user.id", userID),
			zap.String("governor.user.email", user.Email),
		)

		teamID, err := r.teamIDFromName(ctx, workspace)
		if err != nil {
			logger.Error("failed to get workspace id", zap.Error(err))
			continue
		}

		ug, err := r.userGroupFromName(ctx, r.userGroupName(group.Name), teamID, false)
		if err != nil {
			logger.Error("failed to get slack user group", zap.Error(err))
			continue
		}

		u, err := r.Client.GetUserByEmail(ctx, user.Email)
		if err != nil {
			logger.Error("failed to get slack user", zap.Error(err))
			continue
		}

		if !contains(ug.Users, u.ID) {
			logger.Info("user not in group, skipping")
			continue
		}

		newUsers := remove(ug.Users, u.ID)

		logger.Debug("updating user group members", zap.Any("slack.usergroup.existing", ug.Users), zap.Any("slack.usergroup.new", newUsers))

		if r.dryrun {
			logger.Info("SKIP removing user to group", zap.Any("slack.usergroup", *ug))
			continue
		}

		_, err = r.Client.UpdateUserGroupMembers(ctx, ug.ID, teamID, newUsers)
		if err != nil {
			logger.Error("failed to remove user group", zap.String("slack.usergroup.name", r.userGroupName(group.Name)), zap.Error(err))
			return err
		}

		logger.Info("removed user from group")

		if err := auctx.WriteAuditEvent(ctx, r.auditEventWriter, "UserGroupRemoveMember", map[string]string{
			"slack.workspace.name": workspace,
			"slack.usergroup.name": ug.Name,
			"slack.usergroup.id":   ug.ID,
			"slack.user.id":        u.ID,
			"governor.app.id":      appID,
			"governor.group.id":    group.ID,
			"governor.group.slug":  group.Slug,
			"governor.user.id":     userID,
		}); err != nil {
			logger.Error("error writing audit event", zap.Error(err))
		}
	}

	return nil
}

// UpdateUserGroupMembers updates the members of a slack user group to match the members of the governor group
func (r *Reconciler) UpdateUserGroupMembers(ctx context.Context, groupID, appID string) error {
	if groupID == "" || appID == "" {
		return ErrBadParameter
	}

	isSlack, workspace, err := r.isSlackApplication(ctx, appID)
	if err != nil {
		r.Logger.Error("failed to get application from governor", zap.String("governor.app.id", appID), zap.Error(err))
		return err
	}

	if !isSlack {
		r.Logger.Debug("not a slack application, skipping", zap.String("governor.app.id", appID), zap.String("governor.app.name", workspace))
		return nil
	}

	logger := r.Logger.With(zap.String("slack.workspace.name", workspace), zap.String("governor.app.id", appID))

	group, err := r.GovernorClient.Group(ctx, groupID, false)
	if err != nil {
		logger.Error("error getting governor group", zap.String("governor.group.id", groupID), zap.Error(err))
		return err
	}

	// get the current members of the governor group
	members, err := r.GovernorClient.GroupMembers(ctx, groupID)
	if err != nil {
		logger.Error("error getting governor group members", zap.Any("governor.group.id", groupID), zap.Error(err))
		return err
	}

	var memberEmails []string
	for _, m := range members {
		memberEmails = append(memberEmails, strings.ToLower(m.Email))
	}

	logger.Debug("got governor group members", zap.Any("governor.members", memberEmails))

	teamID, err := r.teamIDFromName(ctx, workspace)
	if err != nil {
		return err
	}

	ug, err := r.userGroupFromName(ctx, r.userGroupName(group.Name), teamID, false)
	if err != nil {
		return err
	}

	var newUsers []string

	for _, m := range memberEmails {
		u, err := r.Client.GetUserByEmail(ctx, m)
		if err != nil {
			logger.Info("didn't find slack user", zap.String("user.email", m), zap.Error(err))

			// exit out to prevent deleting valid users
			// only continue if user is not found (404 error)
			if !errors.Is(err, slack.ErrSlackUserNotFound) {
				return err
			}

			continue
		}

		newUsers = append(newUsers, u.ID)
	}

	if equal(ug.Users, newUsers) {
		logger.Debug("no need to update members", zap.Any("slack.usergroup.existing", ug.Users), zap.Any("slack.usergroup.new", newUsers))
		return nil
	}

	logger.Debug("updating user group members", zap.Any("slack.usergroup.existing", ug.Users), zap.Any("slack.usergroup.new", newUsers))

	if r.dryrun {
		logger.Info("SKIP updating slack user group members", zap.Any("slack.usergroup", *ug))
		return nil
	}

	ugUpdated, err := r.Client.UpdateUserGroupMembers(ctx, ug.ID, teamID, newUsers)
	if err != nil {
		logger.Error("failed to update user group", zap.String("slack.usergroup.name", r.userGroupName(group.Name)), zap.Error(err))
		return err
	}

	logger.Info("updated user group members", zap.Any("slack.group.name", ugUpdated.Name), zap.Any("slack.usergroup.users", ugUpdated.Users))

	if err := auctx.WriteAuditEvent(ctx, r.auditEventWriter, "UserGroupUpdateMembers", map[string]string{
		"slack.workspace.name": workspace,
		"slack.usergroup.name": ug.Name,
		"slack.usergroup.id":   ug.ID,
		"slack.user.old":       strings.Join(ug.Users, ","),
		"slack.user.new":       strings.Join(ugUpdated.Users, ","),
		"governor.app.id":      appID,
		"governor.group.id":    group.ID,
		"governor.group.slug":  group.Slug,
	}); err != nil {
		logger.Error("error writing audit event", zap.Error(err))
	}

	return nil
}

// teamID searches all the workspaces (teams) for the given name and returns the ID
// or an error if the team is not found.
func (r *Reconciler) teamIDFromName(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", ErrBadParameter
	}

	workspaces, err := r.Client.ListWorkspaces(ctx)
	if err != nil {
		return "", err
	}

	for _, ws := range workspaces {
		if ws.Name == name {
			r.Logger.Debug("found slack workspace", zap.String("slack.workspace.name", name), zap.String("slack.workspace.id", ws.ID))

			return ws.ID, nil
		}
	}

	r.Logger.Debug("slack workspace not found", zap.String("slack.workspace.name", name))

	return "", ErrSlackWorkspaceNotFound
}

// userGroupFromName searches all the user groups in the workspace for the given name and
// returns the group details or an error if the group is not found.
func (r *Reconciler) userGroupFromName(ctx context.Context, name, teamID string, includeDisabled bool) (*UserGroup, error) {
	if name == "" || teamID == "" {
		return nil, ErrBadParameter
	}

	usergroups, err := r.Client.GetUserGroups(ctx, teamID, includeDisabled)
	if err != nil {
		return nil, err
	}

	for _, ug := range usergroups {
		if ug.Name == name {
			r.Logger.Debug("found slack user group",
				zap.String("slack.usergroup.name", name),
				zap.String("slack.usergroup.id", ug.ID),
				zap.String("slack.workspace.id", teamID),
			)

			return &UserGroup{
				ID:          ug.ID,
				Name:        ug.Name,
				Handle:      ug.Handle,
				Description: ug.Description,
				Users:       ug.Users,
			}, nil
		}
	}

	r.Logger.Debug("slack user group not found", zap.String("slack.usergroup.name", name), zap.String("slack.workspace.id", teamID))

	return nil, ErrSlackUserGroupNotFound
}

// userGroupName applies prefix to the governor group name to create a slack user group name
func (r *Reconciler) userGroupName(groupName string) string {
	return r.userGroupPrefix + groupName
}

// userGroupHandle returns the handle for the slack user group based on the governor group slug
func (r *Reconciler) userGroupHandle(groupSlug string) string {
	// we can just use the group slug as the handle
	return groupSlug
}

func (r *Reconciler) userGroupReq(group *v1alpha1.Group) *slack.UserGroupReq {
	name := r.userGroupName(group.Name)
	handle := r.userGroupHandle(group.Slug)

	return &slack.UserGroupReq{
		Name:        &name,
		Handle:      &handle,
		Description: &group.Description,
	}
}

// contains returns true if the item is in the list
func contains(list []string, item string) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}

	return false
}

// equal returns true if the two slices have the same elements, regardless of order.
// We assume the elements in both slices are unique, i.e. no duplicate values.
func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	exists := make(map[string]bool)
	for _, value := range a {
		exists[value] = true
	}

	for _, value := range b {
		if !exists[value] {
			return false
		}
	}

	return true
}

// remove removes the item from the list
func remove(list []string, item string) []string {
	for i, v := range list {
		if v == item {
			return append(list[:i], list[i+1:]...)
		}
	}

	return list
}

// timestamp returns the current timestamp in Unix format
func timestamp() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}
