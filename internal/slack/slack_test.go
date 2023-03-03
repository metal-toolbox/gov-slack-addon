package slack

import (
	"reflect"
	"testing"

	"go.uber.org/zap"
)

// type mockSlackService struct {
// 	Error error
// }

func TestNewClient(t *testing.T) {
	client := NewClient()

	clientType := reflect.TypeOf(client).String()
	if clientType != "*slack.Client" {
		t.Errorf("expected type to be '*slackql.Client', got %s", clientType)
	}

	client = NewClient(WithToken("test-token"))
	if client.token != "test-token" {
		t.Errorf("expected client token to be 'test-token', got %s", client.token)
	}

	client = NewClient(WithLogger(zap.NewExample()))
	if client.logger.Core().Enabled(zap.DebugLevel) != true {
		t.Error("expected logger debug level to be 'true', got 'false'")
	}
}
