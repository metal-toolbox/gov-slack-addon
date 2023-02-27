package reconciler

// // applicationTypeFilter is the name of the governor application type that we use to filter interesting events
// const applicationTypeFilter = "slack"

// // isSlackApplication returns true if the given governor application id is a slack application.
// // In that case it also returns the application name (which is the Slack organization slug)
// func (r *Reconciler) isSlackApplication(ctx context.Context, appID string) (bool, string, error) {
// 	app, err := r.GovernorClient.Application(ctx, appID)
// 	if err != nil {
// 		return false, "", err
// 	}

// 	name := strings.ToLower(app.Slug)
// 	if name == "" {
// 		return false, name, ErrAppNameEmpty
// 	}

// 	if app.Type == applicationTypeFilter {
// 		return true, name, nil
// 	}

// 	return false, name, nil
// }
