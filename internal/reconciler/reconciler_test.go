package reconciler

import (
	"reflect"
	"testing"
	"time"

	"go.uber.org/zap"
)

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
