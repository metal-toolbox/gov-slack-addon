package reconciler

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
)

func TestReconciler_isSlackApplication(t *testing.T) {
	type args struct {
		appID string
	}

	tests := []struct {
		name    string
		args    args
		resp    []byte
		err     error
		want1   bool
		want2   string
		wantErr bool
	}{
		{
			name:  "slack application lowercase",
			args:  args{appID: "101-slack"},
			resp:  []byte(`{"id": "101-slack", "name": "101-slack", "slug": "101-slack", "type": {"slug": "slack"}}`),
			want1: true,
			want2: "101-slack",
		},
		{
			name:  "slack application with uppercase",
			args:  args{appID: "102-slack"},
			resp:  []byte(`{"id": "102-slack", "name": "Test Slack workspace", "slug": "test-slack-workspace", "type": {"slug": "slack"}}`),
			want1: true,
			want2: "Test Slack workspace",
		},
		{
			name:  "not a slack application",
			args:  args{appID: "103-github"},
			resp:  []byte(`{"id": "103-github", "name": "github-org", "slug": "github-org", "type": {"slug": "github"}}`),
			want1: false,
			want2: "github-org",
		},
		{
			name:    "empty application name",
			args:    args{appID: "104-empty"},
			resp:    []byte(`{"id": "104-empty", "name": "", "slug": "", "type": {"slug": "slack"}}`),
			wantErr: true,
		},
		{
			name:    "governor error",
			args:    args{appID: "101-slack"},
			err:     errors.New("boom"), //nolint:goerr113
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reconciler{
				Logger: zap.NewNop(),
				GovernorClient: mockGovernorClient{
					err:  tt.err,
					resp: tt.resp,
				},
				applicationType: "slack",
			}

			got1, got2, err := r.isSlackApplication(context.TODO(), tt.args.appID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reconciler.isSlackApplication() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got1 != tt.want1 {
				t.Errorf("Reconciler.isSlackApplication() got1 = %v, want1 %v", got1, tt.want1)
			}

			if got2 != tt.want2 {
				t.Errorf("Reconciler.isSlackApplication() got2 = %v, want2 %v", got2, tt.want2)
			}
		})
	}
}
