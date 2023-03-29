package slack

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

func (m *mockSlackService) CreateUserGroupContext(_ context.Context, ug slack.UserGroup) (slack.UserGroup, error) {
	if m.Error != nil {
		return slack.UserGroup{}, m.Error
	}

	if ug.Name == "Already Exists" {
		return slack.UserGroup{}, errors.New("name_already_exists") //nolint:goerr113
	}

	return slack.UserGroup{
		ID:          "G0001",
		TeamID:      ug.TeamID,
		Name:        ug.Name,
		Handle:      ug.Handle,
		Description: ug.Description,
	}, nil
}

func (m *mockSlackService) DisableUserGroupContext(_ context.Context, groupID string, _ ...slack.DisableUserGroupOption) (slack.UserGroup, error) {
	if m.Error != nil {
		return slack.UserGroup{}, m.Error
	}

	if groupID == "notfound" {
		return slack.UserGroup{}, errors.New("subteam_not_found") //nolint:goerr113
	}

	return *m.userGroupResp, nil
}

func (m *mockSlackService) EnableUserGroupContext(_ context.Context, groupID string, _ ...slack.DisableUserGroupOption) (slack.UserGroup, error) {
	if m.Error != nil {
		return slack.UserGroup{}, m.Error
	}

	if groupID == "notfound" {
		return slack.UserGroup{}, errors.New("subteam_not_found") //nolint:goerr113
	}

	return *m.userGroupResp, nil
}

func (m *mockSlackService) GetUserGroupMembersContext(_ context.Context, groupID string, _ ...slack.GetUserGroupMembersOption) ([]string, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	if groupID == "notfound" {
		return nil, errors.New("subteam_not_found") //nolint:goerr113
	}

	return []string{"U0001", "U0002", "U0003"}, nil
}

func (m *mockSlackService) GetUserGroupsContext(_ context.Context, opts ...slack.GetUserGroupsOption) ([]slack.UserGroup, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	params := &slack.GetUserGroupsParams{}
	for _, opt := range opts {
		opt(params)
	}

	if params.TeamID == "notfound" {
		return nil, errors.New("team_not_found") //nolint:goerr113
	}

	return []slack.UserGroup{*m.userGroupResp}, nil
}

func (m *mockSlackService) UpdateUserGroupContext(_ context.Context, groupID string, opts ...slack.UpdateUserGroupsOption) (slack.UserGroup, error) {
	if m.Error != nil {
		return slack.UserGroup{}, m.Error
	}

	params := &slack.UpdateUserGroupsParams{}
	for _, opt := range opts {
		opt(params)
	}

	return slack.UserGroup{
		ID:          groupID,
		TeamID:      *params.TeamID,
		Name:        params.Name,
		Handle:      params.Handle,
		Description: *params.Description,
	}, nil
}

func (m *mockSlackService) UpdateUserGroupMembersContext(_ context.Context, groupID, members string, opts ...slack.UpdateUserGroupMembersOption) (slack.UserGroup, error) {
	if m.Error != nil {
		return slack.UserGroup{}, m.Error
	}

	if groupID == "notfound" {
		return slack.UserGroup{}, errors.New("subteam_not_found") //nolint:goerr113
	}

	params := &slack.UpdateUserGroupMembersParams{}
	for _, opt := range opts {
		opt(params)
	}

	return slack.UserGroup{
		ID:     groupID,
		TeamID: params.TeamID,
		Users:  strings.Split(members, ","),
	}, nil
}

func TestClient_CreateUserGroup(t *testing.T) {
	type args struct {
		teamID    string
		userGroup *UserGroupReq
	}

	testResp := &slack.UserGroup{
		ID:          "G0001",
		TeamID:      "T0123",
		Name:        "Group 1",
		Handle:      "group1",
		Description: "Test Group 1",
	}

	tests := []struct {
		name    string
		args    args
		err     error
		want    *slack.UserGroup
		wantErr bool
	}{
		{
			name: "successful create user group",
			args: args{
				teamID: "T0123",
				userGroup: &UserGroupReq{
					Name:        stringPtr("Group 1"),
					Handle:      stringPtr("group1"),
					Description: stringPtr("Test Group 1"),
				},
			},
			want: testResp,
		},
		{
			name:    "empty team id",
			args:    args{teamID: ""},
			wantErr: true,
		},
		{
			name: "missing name",
			args: args{
				teamID: "T0123",
				userGroup: &UserGroupReq{
					Handle:      stringPtr("group1"),
					Description: stringPtr("Test Group 1"),
				},
			},
			wantErr: true,
		},
		{
			name: "missing handle",
			args: args{
				teamID: "T0123",
				userGroup: &UserGroupReq{
					Name:        stringPtr("Group 1"),
					Description: stringPtr("Test Group 1"),
				},
			},
			wantErr: true,
		},
		{
			name: "user group already exists",
			args: args{
				teamID: "T0123",
				userGroup: &UserGroupReq{
					Name:        stringPtr("Already Exists"),
					Handle:      stringPtr("group-exists"),
					Description: stringPtr("Test group that already exists"),
				},
			},
			wantErr: true,
		},
		{
			name:    "slack error",
			err:     errors.New("boom"), //nolint:goerr113
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				logger: zap.NewNop(),
				slackService: &mockSlackService{
					Error: tt.err,
				},
			}

			got, err := c.CreateUserGroup(context.TODO(), tt.args.teamID, tt.args.userGroup)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.CreateUserGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.CreateUserGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_DisableUserGroup(t *testing.T) {
	type args struct {
		groupID string
		teamID  string
	}

	testResp := &slack.UserGroup{
		ID:          "G0001",
		TeamID:      "T0123",
		Name:        "Group 1",
		Handle:      "group1",
		Description: "Test Group 1",
	}

	tests := []struct {
		name    string
		args    args
		err     error
		resp    *slack.UserGroup
		want    *slack.UserGroup
		wantErr bool
	}{
		{
			name: "successful disable user group",
			args: args{
				groupID: "G0001",
				teamID:  "T0123",
			},
			resp: testResp,
			want: testResp,
		},
		{
			name:    "empty group id",
			args:    args{groupID: "", teamID: "T0123"},
			wantErr: true,
		},
		{
			name:    "empty team id",
			args:    args{groupID: "G0001", teamID: ""},
			wantErr: true,
		},
		{
			name: "user group not found",
			args: args{
				groupID: "notfound",
				teamID:  "T0123",
			},
			wantErr: true,
		},
		{
			name:    "slack error",
			err:     errors.New("boom"), //nolint:goerr113
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				logger: zap.NewNop(),
				slackService: &mockSlackService{
					Error:         tt.err,
					userGroupResp: tt.resp,
				},
			}

			got, err := c.DisableUserGroup(context.TODO(), tt.args.groupID, tt.args.teamID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.DisableUserGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.DisableUserGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_EnableUserGroup(t *testing.T) {
	type args struct {
		groupID string
		teamID  string
	}

	testResp := &slack.UserGroup{
		ID:          "G0001",
		TeamID:      "T0123",
		Name:        "Group 1",
		Handle:      "group1",
		Description: "Test Group 1",
	}

	tests := []struct {
		name    string
		args    args
		err     error
		resp    *slack.UserGroup
		want    *slack.UserGroup
		wantErr bool
	}{
		{
			name: "successful enable user group",
			args: args{
				groupID: "G0001",
				teamID:  "T0123",
			},
			resp: testResp,
			want: testResp,
		},
		{
			name:    "empty group id",
			args:    args{groupID: "", teamID: "T0123"},
			wantErr: true,
		},
		{
			name:    "empty team id",
			args:    args{groupID: "G0001", teamID: ""},
			wantErr: true,
		},
		{
			name: "user group not found",
			args: args{
				groupID: "notfound",
				teamID:  "T0123",
			},
			wantErr: true,
		},
		{
			name:    "slack error",
			err:     errors.New("boom"), //nolint:goerr113
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				logger: zap.NewNop(),
				slackService: &mockSlackService{
					Error:         tt.err,
					userGroupResp: tt.resp,
				},
			}

			got, err := c.EnableUserGroup(context.TODO(), tt.args.groupID, tt.args.teamID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.EnableUserGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.EnableUserGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetUserGroups(t *testing.T) {
	type args struct {
		teamID          string
		includeDisabled bool
	}

	testResp := []slack.UserGroup{
		{
			ID:          "G0001",
			TeamID:      "T0123",
			Name:        "Group 1",
			Handle:      "group1",
			Description: "Test Group 1",
		},
	}

	tests := []struct {
		name    string
		args    args
		err     error
		resp    []slack.UserGroup
		want    []slack.UserGroup
		wantErr bool
	}{
		{
			name: "successful get user groups",
			args: args{
				teamID:          "T0123",
				includeDisabled: true,
			},
			resp: testResp,
			want: testResp,
		},
		{
			name:    "empty team id",
			args:    args{teamID: ""},
			resp:    testResp,
			wantErr: true,
		},
		{
			name:    "team not found",
			args:    args{teamID: "notfound"},
			resp:    testResp,
			wantErr: true,
		},
		{
			name:    "slack error",
			resp:    testResp,
			err:     errors.New("boom"), //nolint:goerr113
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				logger: zap.NewNop(),
				slackService: &mockSlackService{
					Error:         tt.err,
					userGroupResp: &tt.resp[0],
				},
			}

			got, err := c.GetUserGroups(context.TODO(), tt.args.teamID, tt.args.includeDisabled)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetUserGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetUserGroups() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetUserGroupMembers(t *testing.T) {
	type args struct {
		groupID         string
		teamID          string
		includeDisabled bool
	}

	testResp := []string{"U0001", "U0002", "U0003"}

	tests := []struct {
		name    string
		args    args
		err     error
		want    []string
		wantErr bool
	}{
		{
			name: "successful get user group members",
			args: args{groupID: "G0001", teamID: "T0123"},
			want: testResp,
		},
		{
			name:    "empty group id",
			args:    args{groupID: "", teamID: "T0123"},
			wantErr: true,
		},
		{
			name:    "empty team id",
			args:    args{groupID: "G0001", teamID: ""},
			wantErr: true,
		},
		{
			name:    "team not found",
			args:    args{teamID: "notfound"},
			wantErr: true,
		},
		{
			name:    "slack error",
			err:     errors.New("boom"), //nolint:goerr113
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				logger: zap.NewNop(),
				slackService: &mockSlackService{
					Error: tt.err,
				},
			}

			got, err := c.GetUserGroupMembers(context.TODO(), tt.args.groupID, tt.args.teamID, tt.args.includeDisabled)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetUserGroupMembers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetUserGroupMembers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_UpdateUserGroup(t *testing.T) {
	type args struct {
		groupID   string
		teamID    string
		userGroup UserGroupReq
	}

	testResp := &slack.UserGroup{
		ID:          "G0001",
		TeamID:      "T0123",
		Name:        "Group 1",
		Handle:      "group1",
		Description: "Test Group 1",
	}

	tests := []struct {
		name    string
		args    args
		err     error
		want    *slack.UserGroup
		wantErr bool
	}{
		{
			name: "successful update user group",
			args: args{
				groupID: "G0001",
				teamID:  "T0123",
				userGroup: UserGroupReq{
					Name:        stringPtr("Group 1"),
					Handle:      stringPtr("group1"),
					Description: stringPtr("Test Group 1"),
				},
			},
			want: testResp,
		},
		{
			name:    "empty group id",
			args:    args{groupID: "", teamID: "T0123"},
			wantErr: true,
		},
		{
			name:    "empty team id",
			args:    args{groupID: "G0001", teamID: ""},
			wantErr: true,
		},
		{
			name: "missing group params",
			args: args{
				groupID:   "G0001",
				teamID:    "T0123",
				userGroup: UserGroupReq{},
			},
			wantErr: true,
		},
		{
			name:    "slack error",
			err:     errors.New("boom"), //nolint:goerr113
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				logger: zap.NewNop(),
				slackService: &mockSlackService{
					Error: tt.err,
				},
			}

			got, err := c.UpdateUserGroup(context.TODO(), tt.args.groupID, tt.args.teamID, tt.args.userGroup)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.UpdateUserGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.UpdateUserGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_UpdateUserGroupMembers(t *testing.T) {
	type args struct {
		groupID string
		teamID  string
		members []string
	}

	testResp := &slack.UserGroup{
		ID:     "G0001",
		TeamID: "T0123",
		Users:  []string{"U0001", "U0002", "U0003"},
	}

	tests := []struct {
		name    string
		args    args
		err     error
		want    *slack.UserGroup
		wantErr bool
	}{
		{
			name: "successful update user group members",
			args: args{
				groupID: "G0001",
				teamID:  "T0123",
				members: []string{"U0001", "U0002", "U0003"},
			},
			want: testResp,
		},
		{
			name: "empty group id",
			args: args{
				teamID:  "T0123",
				members: []string{"U0001", "U0002", "U0003"},
			},
			wantErr: true,
		},
		{
			name: "empty team id",
			args: args{
				groupID: "G0001",
				members: []string{"U0001", "U0002", "U0003"},
			},
			wantErr: true,
		},
		{
			name: "empty members",
			args: args{
				groupID: "G0001",
				teamID:  "T0123",
			},
			wantErr: true,
		},
		{
			name: "user group not found",
			args: args{
				groupID: "notfound",
				teamID:  "T0123",
			},
			wantErr: true,
		},
		{
			name:    "slack error",
			err:     errors.New("boom"), //nolint:goerr113
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				logger: zap.NewNop(),
				slackService: &mockSlackService{
					Error: tt.err,
				},
			}

			got, err := c.UpdateUserGroupMembers(context.TODO(), tt.args.groupID, tt.args.teamID, tt.args.members)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.UpdateUserGroupMembers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.UpdateUserGroupMembers() = %v, want %v", got, tt.want)
			}
		})
	}
}
