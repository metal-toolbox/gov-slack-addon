package slack

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

func (m *mockSlackService) GetUserInfoContext(_ context.Context, id string) (*slack.User, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	if id == "notfound" {
		return nil, errors.New("user_not_found") //nolint:goerr113
	}

	return m.userResp, nil
}

func (m *mockSlackService) GetUserByEmailContext(_ context.Context, id string) (*slack.User, error) {
	if m.Error != nil {
		return nil, m.Error
	}

	if id == "notfound" {
		return nil, errors.New("users_not_found") //nolint:goerr113
	}

	return m.userResp, nil
}

func TestClient_GetUser(t *testing.T) {
	type args struct {
		id string
	}

	testResp := &slack.User{
		ID:   "U0001",
		Name: "User 1",
	}

	tests := []struct {
		name    string
		args    args
		err     error
		resp    *slack.User
		want    *slack.User
		wantErr bool
	}{
		{
			name: "successful user by id",
			args: args{id: "U0001"},
			resp: testResp,
			want: testResp,
		},
		{
			name:    "empty id",
			args:    args{id: ""},
			wantErr: true,
		},
		{
			name:    "user not found",
			args:    args{id: "notfound"},
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
					Error:    tt.err,
					userResp: tt.resp,
				},
			}

			got, err := c.GetUser(context.TODO(), tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetUserByEmail(t *testing.T) {
	type args struct {
		email string
	}

	testResp := &slack.User{
		ID:   "U0001",
		Name: "User 1",
		Profile: slack.UserProfile{
			Email: "user1@example.com",
		},
	}

	tests := []struct {
		name    string
		args    args
		err     error
		resp    *slack.User
		want    *slack.User
		wantErr bool
	}{
		{
			name: "successful user by email",
			args: args{email: "user1@example.com"},
			resp: testResp,
			want: testResp,
		},
		{
			name:    "empty email",
			args:    args{email: ""},
			wantErr: true,
		},
		{
			name:    "user not found",
			args:    args{email: "notfound"},
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
					Error:    tt.err,
					userResp: tt.resp,
				},
			}

			got, err := c.GetUserByEmail(context.TODO(), tt.args.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.GetUserByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.GetUserByEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}
