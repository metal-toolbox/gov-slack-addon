package natssrv

import (
	"context"
	"encoding/json"

	"github.com/metal-toolbox/auditevent"
	"github.com/metal-toolbox/gov-slack-addon/internal/auctx"
	"github.com/metal-toolbox/governor-api/pkg/events/v1alpha1"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

// ApplicationsMessageHandler handles messages for governor applications events (linking/unlinking a group)
func (s *Server) ApplicationsMessageHandler(m *nats.Msg) {
	payload, err := s.unmarshalPayload(m)
	if err != nil {
		s.Logger.Warn("unable to unmarshal governor payload", zap.Error(err))
		return
	}

	if payload.GroupID == "" {
		s.Logger.Error("bad event payload", zap.Error(ErrEventMissingGroupID))
		return
	}

	logger := s.Logger.With(zap.String("governor.group.id", payload.GroupID))

	ctx := context.Background()

	switch payload.Action {
	case v1alpha1.GovernorEventCreate:
		// when a group is linked to a slack app, we'll create the group in slack and set the members
		logger.Info("create application link event")

		ctx = auctx.WithAuditEvent(ctx, s.auditEventNATS(m.Subject, payload))

		if err := s.Reconciler.CreateUserGroup(ctx, payload.GroupID, payload.ApplicationID); err != nil {
			logger.Error("error creating user group", zap.Error(err))
			return
		}

		if err := s.Reconciler.UpdateUserGroupMembers(ctx, payload.GroupID, payload.ApplicationID); err != nil {
			logger.Error("error setting user group members", zap.Error(err))
			return
		}

	case v1alpha1.GovernorEventDelete:
		// when a group is unlinked from a slack app, we'll delete/disable the group in slack
		logger.Info("delete application link event")

		ctx = auctx.WithAuditEvent(ctx, s.auditEventNATS(m.Subject, payload))

		if err := s.Reconciler.DeleteUserGroup(ctx, payload.GroupID, payload.ApplicationID); err != nil {
			logger.Error("error deleting user group", zap.Error(err))
			return
		}

	default:
		logger.Warn("unexpected action in governor event", zap.String("governor.action", payload.Action))
		return
	}
}

// MembersMessageHandler handles messages for governor members events (adding/removing members)
func (s *Server) MembersMessageHandler(m *nats.Msg) {
	payload, err := s.unmarshalPayload(m)
	if err != nil {
		s.Logger.Warn("unable to unmarshal governor payload", zap.Error(err))
		return
	}

	if payload.GroupID == "" {
		s.Logger.Error("bad event payload", zap.Error(ErrEventMissingGroupID))
		return
	}

	logger := s.Logger.With(zap.String("governor.group.id", payload.GroupID), zap.String("governor.user.id", payload.UserID))

	ctx := context.Background()

	switch payload.Action {
	case v1alpha1.GovernorEventCreate:
		logger.Info("create group member event")

		ctx = auctx.WithAuditEvent(ctx, s.auditEventNATS(m.Subject, payload))

		if err := s.Reconciler.AddUserGroupMember(ctx, payload.GroupID, payload.UserID); err != nil {
			logger.Error("error adding user group member", zap.Error(err))
			return
		}

	case v1alpha1.GovernorEventDelete:
		logger.Info("delete group member event")

		ctx = auctx.WithAuditEvent(ctx, s.auditEventNATS(m.Subject, payload))

		if err := s.Reconciler.RemoveUserGroupMember(ctx, payload.GroupID, payload.UserID); err != nil {
			logger.Error("error removing user group member", zap.Error(err))
			return
		}

	default:
		logger.Warn("unexpected action in governor event", zap.String("governor.action", payload.Action))
		return
	}
}

func (s *Server) unmarshalPayload(m *nats.Msg) (*v1alpha1.Event, error) {
	s.Logger.Debug("received a message:", zap.String("nats.data", string(m.Data)), zap.String("nats.subject", m.Subject))

	payload := v1alpha1.Event{}
	if err := json.Unmarshal(m.Data, &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}

// auditEventNATS returns a stub NATS audit event
func (s *Server) auditEventNATS(natsSubj string, event *v1alpha1.Event) *auditevent.AuditEvent {
	return auditevent.NewAuditEventWithID(
		event.AuditID,
		"", // eventType to be populated later
		auditevent.EventSource{
			Type:  "NATS",
			Value: s.NATSClient.conn.ConnectedUrlRedacted(),
			Extra: map[string]interface{}{
				"nats.subject":    natsSubj,
				"nats.queuegroup": s.NATSClient.queueGroup,
			},
		},
		auditevent.OutcomeSucceeded,
		map[string]string{
			"event": "governor",
		},
		"gov-slack-addon",
	)
}
