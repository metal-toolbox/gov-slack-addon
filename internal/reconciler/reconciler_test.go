package reconciler

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/metal-toolbox/governor-api/pkg/api/v1alpha1"
	"go.uber.org/zap"
)

// ErrMissingMockedResponse is returned when a mocked response is missing
var ErrMissingMockedResponse = errors.New("missing mocked response")

type mockGovernorClient struct {
	err  error
	resp []byte
}

func (m mockGovernorClient) Application(_ context.Context, _ string) (*v1alpha1.Application, error) {
	if m.err != nil {
		return nil, m.err
	}

	if m.resp == nil {
		return nil, ErrMissingMockedResponse
	}

	out := v1alpha1.Application{}
	if err := json.Unmarshal(m.resp, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (m mockGovernorClient) Applications(_ context.Context) ([]*v1alpha1.Application, error) {
	if m.err != nil {
		return nil, m.err
	}

	if m.resp == nil {
		return nil, ErrMissingMockedResponse
	}

	out := []*v1alpha1.Application{}
	if err := json.Unmarshal(m.resp, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func (m mockGovernorClient) ApplicationTypes(_ context.Context) ([]*v1alpha1.ApplicationType, error) {
	if m.err != nil {
		return nil, m.err
	}

	if m.resp == nil {
		return nil, ErrMissingMockedResponse
	}

	out := []*v1alpha1.ApplicationType{}
	if err := json.Unmarshal(m.resp, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func (m mockGovernorClient) ApplicationGroups(_ context.Context, _ string) ([]*v1alpha1.Group, error) {
	if m.err != nil {
		return nil, m.err
	}

	if m.resp == nil {
		return nil, ErrMissingMockedResponse
	}

	out := []*v1alpha1.Group{}
	if err := json.Unmarshal(m.resp, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func (m mockGovernorClient) Group(_ context.Context, _ string, _ bool) (*v1alpha1.Group, error) {
	if m.err != nil {
		return nil, m.err
	}

	if m.resp == nil {
		return nil, ErrMissingMockedResponse
	}

	out := v1alpha1.Group{}
	if err := json.Unmarshal(m.resp, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (m mockGovernorClient) GroupMembers(_ context.Context, _ string) ([]*v1alpha1.GroupMember, error) {
	if m.err != nil {
		return nil, m.err
	}

	if m.resp == nil {
		return nil, ErrMissingMockedResponse
	}

	out := []*v1alpha1.GroupMember{}
	if err := json.Unmarshal(m.resp, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func (m mockGovernorClient) User(_ context.Context, _ string, _ bool) (*v1alpha1.User, error) {
	if m.err != nil {
		return nil, m.err
	}

	if m.resp == nil {
		return nil, ErrMissingMockedResponse
	}

	out := v1alpha1.User{}
	if err := json.Unmarshal(m.resp, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

func (m mockGovernorClient) URL() string {
	return "https://governor.example.com"
}

func TestNew(t *testing.T) {
	reconciler := New()

	reconcilerType := reflect.TypeOf(reconciler).String()
	if reconcilerType != "*reconciler.Reconciler" {
		t.Errorf("expected type to be '*reconciler.Reconciler', got %s", reconcilerType)
	}

	reconciler = New(WithLogger(zap.NewExample()))
	if reconciler.Logger.Core().Enabled(zap.DebugLevel) != true {
		t.Error("expected logger debug level to be 'true', got 'false'")
	}

	reconciler = New(WithDryRun(true))
	if reconciler.dryrun != true {
		t.Errorf("expected reconciler dryrun to be 'true', got %t", reconciler.dryrun)
	}

	reconciler = New(WithInterval(10 * time.Minute))
	if reconciler.interval != 10*time.Minute {
		t.Errorf("expected reconciler interval to be '10m0s', got %s", reconciler.interval)
	}

	reconciler = New(WithQueue("NATS"))
	if reconciler.queue != "NATS" {
		t.Errorf("expected reconciler queue to be 'NATS', got %s", reconciler.queue)
	}
}
