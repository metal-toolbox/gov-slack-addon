package natssrv

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"go.equinixmetal.net/governor-api/pkg/events/v1alpha1"
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

	// ctx := context.Background()

	switch payload.Action {
	case v1alpha1.GovernorEventCreate:
		logger.Info("create application event")

		// ctx = auctx.WithAuditEvent(ctx, s.auditEventNATS(m.Subject, payload))

		// TODO

	case v1alpha1.GovernorEventDelete:
		// when a group is unlinked from a slack app, we'll delete the group in slack
		logger.Info("delete application event")

		// ctx = auctx.WithAuditEvent(ctx, s.auditEventNATS(m.Subject, payload))

		// TODO

	default:
		logger.Warn("unexpected action in governor event", zap.String("governor.action", payload.Action))
		return
	}
}

// GroupsMessageHandler handles messages for governor groups events (deleting a group)
func (s *Server) GroupsMessageHandler(m *nats.Msg) {
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

	// ctx := context.Background()

	switch payload.Action {
	case v1alpha1.GovernorEventDelete:
		logger.Info("delete group event")

		// ctx = auctx.WithAuditEvent(ctx, s.auditEventNATS(m.Subject, payload))

		// TODO

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

	// ctx := context.Background()

	switch payload.Action {
	case v1alpha1.GovernorEventCreate:
		logger.Info("create group member event")

		// ctx = auctx.WithAuditEvent(ctx, s.auditEventNATS(m.Subject, payload))

		// TODO

	case v1alpha1.GovernorEventDelete:
		logger.Info("delete group member event")

		// ctx = auctx.WithAuditEvent(ctx, s.auditEventNATS(m.Subject, payload))

		// TODO

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
// func (s *Server) auditEventNATS(natsSubj string, event *v1alpha1.Event) *auditevent.AuditEvent {
// 	return auditevent.NewAuditEventWithID(
// 		event.AuditID,
// 		"", // eventType to be populated later
// 		auditevent.EventSource{
// 			Type:  "NATS",
// 			Value: s.NATSClient.conn.ConnectedUrlRedacted(),
// 			Extra: map[string]interface{}{
// 				"nats.subject":    natsSubj,
// 				"nats.queuegroup": s.NATSClient.queueGroup,
// 			},
// 		},
// 		auditevent.OutcomeSucceeded,
// 		map[string]string{
// 			"event": "governor",
// 		},
// 		"gov-slack-addon",
// 	)
// }
