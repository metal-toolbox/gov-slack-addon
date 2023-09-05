package reconciler

import (
	"context"
)

// isSlackApplication returns true if the given governor application id is a slack application.
// In that case it also returns the application name (which should be the Slack workspace name)
func (r *Reconciler) isSlackApplication(ctx context.Context, appID string) (bool, string, error) {
	app, err := r.GovernorClient.Application(ctx, appID)
	if err != nil {
		return false, "", err
	}

	name := app.Name
	if name == "" {
		return false, name, ErrAppNameEmpty
	}

	if app.Type.Slug == r.applicationType {
		return true, name, nil
	}

	return false, name, nil
}
