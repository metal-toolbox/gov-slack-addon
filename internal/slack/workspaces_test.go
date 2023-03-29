package slack

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

func (m *mockSlackService) ListTeamsContext(_ context.Context, _ slack.ListTeamsParameters) ([]slack.Team, string, error) {
	if m.Error != nil {
		return nil, "", m.Error
	}

	resp := []slack.Team{
		{
			ID:     "T0001",
			Name:   "Team 1",
			Domain: "",
		},
		{
			ID:     "T0002",
			Name:   "Team 2",
			Domain: "",
		},
	}

	return resp, "", nil
}

func TestClient_ListWorkspaces(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		want    []slack.Team
		wantErr bool
	}{
		{
			name: "success list workspaces",
			want: []slack.Team{
				{
					ID:     "T0001",
					Name:   "Team 1",
					Domain: "",
				},
				{
					ID:     "T0002",
					Name:   "Team 2",
					Domain: "",
				},
			},
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
				logger:       zap.NewNop(),
				slackService: &mockSlackService{Error: tt.err},
			}

			got, err := c.ListWorkspaces(context.TODO())
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.ListWorkspaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client.ListWorkspaces() = %v, want %v", got, tt.want)
			}
		})
	}
}
