package natssrv

import (
	"context"

	"github.com/metal-toolbox/auditevent"
	"github.com/metal-toolbox/gov-slack-addon/internal/auctx"
	"github.com/metal-toolbox/governor-api/pkg/events/v1alpha1"
	"github.com/metal-toolbox/governor-extension-sdk/pkg/eventrouter"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

// ApplicationsLink handles a group being linked to a slack application, it
// creates the corresponding slack user group and syncs its members.
func (p *Processor) ApplicationsLink(ctx context.Context, payload *v1alpha1.Event) error {
	ctx, span := p.tracer.Start(ctx, "process-applink-create")
	defer span.End()

	logger := p.logger.With(zap.String("governor.group.id", payload.GroupID))

	if payload.GroupID == "" {
		logger.Error("bad event payload", zap.Error(ErrEventMissingGroupID))
		return ErrEventMissingGroupID
	}

	logger.Info("create application link event")

	if err := p.reconciler.CreateUserGroup(ctx, payload.GroupID, payload.ApplicationID); err != nil {
		logger.Error("error creating user group", zap.Error(err))
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	if err := p.reconciler.UpdateUserGroupMembers(ctx, payload.GroupID, payload.ApplicationID); err != nil {
		logger.Error("error setting user group members", zap.Error(err))
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	return nil
}

// ApplicationUnlink handles a group being unlinked from a slack application:
// it deletes/disables the corresponding slack user group.
func (p *Processor) ApplicationUnlink(ctx context.Context, payload *v1alpha1.Event) error {
	ctx, span := p.tracer.Start(ctx, "process-applink-delete")
	defer span.End()

	logger := p.logger.With(zap.String("governor.group.id", payload.GroupID))

	if payload.GroupID == "" {
		logger.Error("bad event payload", zap.Error(ErrEventMissingGroupID))
		return ErrEventMissingGroupID
	}

	logger.Info("delete application link event")

	if err := p.reconciler.DeleteUserGroup(ctx, payload.GroupID, payload.ApplicationID); err != nil {
		logger.Error("error deleting user group", zap.Error(err))
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	return nil
}

// MemberCreate handles a member being added to a group.
func (p *Processor) MemberCreate(ctx context.Context, payload *v1alpha1.Event) error {
	ctx, span := p.tracer.Start(ctx, "process-member-create")
	defer span.End()

	logger := p.logger.With(zap.String("governor.group.id", payload.GroupID), zap.String("governor.user.id", payload.UserID))

	if payload.GroupID == "" {
		logger.Error("bad event payload", zap.Error(ErrEventMissingGroupID))
		return ErrEventMissingGroupID
	}

	logger.Info("create group member event")

	if err := p.reconciler.AddUserGroupMember(ctx, payload.GroupID, payload.UserID); err != nil {
		logger.Error("error adding user group member", zap.Error(err))
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	return nil
}

// MemberDelete handles a member being removed from a group.
func (p *Processor) MemberDelete(ctx context.Context, payload *v1alpha1.Event) error {
	ctx, span := p.tracer.Start(ctx, "process-member-delete")
	defer span.End()

	logger := p.logger.With(zap.String("governor.group.id", payload.GroupID), zap.String("governor.user.id", payload.UserID))

	if payload.GroupID == "" {
		logger.Error("bad event payload", zap.Error(ErrEventMissingGroupID))
		return ErrEventMissingGroupID
	}

	logger.Info("delete group member event")

	if err := p.reconciler.RemoveUserGroupMember(ctx, payload.GroupID, payload.UserID); err != nil {
		logger.Error("error removing user group member", zap.Error(err))
		span.SetStatus(codes.Error, err.Error())

		return err
	}

	return nil
}

func (p *Processor) auditMiddleware(next eventrouter.Handler) eventrouter.Handler {
	return func(ctx context.Context, e *v1alpha1.Event) error {
		subject := eventrouter.GetSubjectFromContext(ctx)

		ctx = auctx.WithAuditEvent(
			ctx,
			auditevent.NewAuditEventWithID(
				e.AuditID,
				"", // eventType to be populated later
				auditevent.EventSource{
					Type:  "NATS",
					Value: subject,
					Extra: map[string]interface{}{
						"nats.subject": subject,
					},
				},
				auditevent.OutcomeSucceeded,
				map[string]string{
					"event": "governor",
				},
				"gov-slack-addon",
			),
		)

		return next(ctx, e)
	}
}
