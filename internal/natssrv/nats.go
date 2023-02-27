package natssrv

import (
	"fmt"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

const (
	natsSuffixApps    = "apps"
	natsSuffixGroups  = "groups"
	natsSuffixMembers = "members"
)

// NATSClient is a NATS client with some configuration
type NATSClient struct {
	conn       *nats.Conn
	logger     *zap.Logger
	prefix     string
	queueGroup string
	queueSize  int
	subject    string
}

// NATSOption is a functional configuration option for NATS
type NATSOption func(c *NATSClient)

// NewNATSClient configures and establishes a new NATS client connection
func NewNATSClient(opts ...NATSOption) (*NATSClient, error) {
	client := NATSClient{
		logger: zap.NewNop(),
	}

	for _, opt := range opts {
		opt(&client)
	}

	return &client, nil
}

// WithNATSConn sets the nats connection
func WithNATSConn(nc *nats.Conn) NATSOption {
	return func(c *NATSClient) {
		c.conn = nc
	}
}

// WithNATSPrefix sets the nats subscription prefix
func WithNATSPrefix(p string) NATSOption {
	return func(c *NATSClient) {
		c.prefix = p
	}
}

// WithNATSSubject sets the nats subscription subject
func WithNATSSubject(s string) NATSOption {
	return func(c *NATSClient) {
		c.subject = s
	}
}

// WithNATSQueueGroup sets the nats subscription queue group
func WithNATSQueueGroup(q string, s int) NATSOption {
	return func(c *NATSClient) {
		c.queueGroup = q
		c.queueSize = s
	}
}

// WithNATSLogger sets the NATS client logger
func WithNATSLogger(l *zap.Logger) NATSOption {
	return func(c *NATSClient) {
		c.logger = l
	}
}

func (s *Server) registerSubscriptionHandlers() error {
	prefix := s.NATSClient.prefix
	qg := s.NATSClient.queueGroup

	s.Logger.Debug("registering subscription handlers", zap.String("nats.prefix", prefix), zap.String("nats.queue_group", qg))

	n := 1
	for n < s.NATSClient.queueSize {
		// Receive application channel events
		subj := fmt.Sprintf("%s.%s", prefix, natsSuffixApps)
		if _, err := s.NATSClient.conn.QueueSubscribe(subj, qg, s.ApplicationsMessageHandler); err != nil {
			return err
		}

		s.Logger.Debug("added subscriber", zap.String("nats.subscriber_id", fmt.Sprintf("%s-%d", subj, n)))

		// Receive groups channel events
		subj = fmt.Sprintf("%s.%s", prefix, natsSuffixGroups)
		if _, err := s.NATSClient.conn.QueueSubscribe(subj, qg, s.GroupsMessageHandler); err != nil {
			return err
		}

		s.Logger.Debug("added subscriber", zap.String("nats.subscriber_id", fmt.Sprintf("%s-%d", subj, n)))

		// Receive group memberships channel events
		subj = fmt.Sprintf("%s.%s", prefix, natsSuffixMembers)
		if _, err := s.NATSClient.conn.QueueSubscribe(subj, qg, s.MembersMessageHandler); err != nil {
			return err
		}

		s.Logger.Debug("added subscriber", zap.String("nats.subscriber_id", fmt.Sprintf("%s-%d", subj, n)))

		n++
	}

	return nil
}

func (s *Server) shutdownSubscriptions() error {
	// Drain and close the NATS connection
	return s.NATSClient.conn.Drain()
}
