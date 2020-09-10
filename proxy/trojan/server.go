package trojan

import (
	"context"
	"time"

	"v2ray.com/core"
	"v2ray.com/core/common"
	"v2ray.com/core/common/buf"
	"v2ray.com/core/common/log"
	"v2ray.com/core/common/net"
	"v2ray.com/core/common/protocol"
	"v2ray.com/core/common/retry"
	"v2ray.com/core/common/session"
	"v2ray.com/core/common/signal"
	"v2ray.com/core/common/task"
	"v2ray.com/core/features/policy"
	"v2ray.com/core/features/routing"
	"v2ray.com/core/transport/internet"
)

func init() {
	common.Must(common.RegisterConfig((*ServerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewServer(ctx, config.(*ServerConfig))
	}))
}

// Server is an inbound connection handler that handles messages in trojan protocol.
type Server struct {
	validator     *Validator
	policyManager policy.Manager
	config        *ServerConfig
}

// New creates a new trojan inbound handler.
func NewServer(ctx context.Context, config *ServerConfig) (*Server, error) {

	validator := new(Validator)
	for _, user := range config.Users {
		u, err := user.ToMemoryUser()
		if err != nil {
			return nil, newError("failed to get trojan user").Base(err).AtError()
		}

		validator.Add(u)
	}

	v := core.MustFromContext(ctx)
	server := &Server{
		policyManager: v.GetFeature(policy.ManagerType()).(policy.Manager),
		validator:     validator,
		config:        config,
	}

	return server, nil
}

func (s *Server) Network() []net.Network {
	return []net.Network{net.Network_TCP}
}

func (s *Server) Process(ctx context.Context, network net.Network, conn internet.Connection, dispatcher routing.Dispatcher) error {
	sessionPolicy := s.policyManager.ForLevel(0)
	if err := conn.SetReadDeadline(time.Now().Add(sessionPolicy.Timeouts.Handshake)); err != nil {
		return newError("unable to set read deadline").Base(err).AtWarning()
	}

	buffer := buf.StackNew()
	defer buffer.Release()

	bufferedReader := buf.BufferedReader{Reader: buf.NewReader(conn)}
	n, err := buffer.ReadFullFrom(&bufferedReader, 56)
	if err != nil {
		return newError("unable to read user hash").Base(err)
	}

	var user *protocol.MemoryUser
	fallbackEnabled := s.config.Fallback != nil
	shouldFallback := false
	if n < 56 {
		// invalid protocol
		log.Record(&log.AccessMessage{
			From:   conn.RemoteAddr(),
			To:     "",
			Status: log.AccessRejected,
			Reason: newError("not trojan protocol"),
		})

		shouldFallback = true
	} else {
		user = s.validator.Get(hexString(buffer.BytesTo(56)))
		if user == nil {
			// invalid user, let's fallback
			log.Record(&log.AccessMessage{
				From:   conn.RemoteAddr(),
				To:     "",
				Status: log.AccessRejected,
				Reason: newError("not a valid user"),
			})

			shouldFallback = true
		}
	}

	if fallbackEnabled && shouldFallback {
		return s.fallback(ctx, &buffer, buf.NewReader(conn), buf.NewWriter(conn))
	} else if shouldFallback {
		return newError("invalid protocol or invalid user")
	}

	dest, bodyReader, err := ReadHeader(&bufferedReader)
	if err != nil {
		log.Record(&log.AccessMessage{
			From:   conn.RemoteAddr(),
			To:     "",
			Status: log.AccessRejected,
			Reason: err,
		})
		return newError("failed to create request from: ", conn.RemoteAddr()).Base(err)
	}
	destination := *dest
	conn.SetReadDeadline(time.Time{})

	inbound := session.InboundFromContext(ctx)
	if inbound == nil {
		panic("no inbound metadata")
	}
	inbound.User = user

	log.ContextWithAccessMessage(ctx, &log.AccessMessage{
		From:   conn.RemoteAddr(),
		To:     destination,
		Status: log.AccessAccepted,
		Reason: "",
		Email:  user.Email,
	})
	newError("received request for ", destination).WriteToLog(session.ExportIDToError(ctx))

	sessionPolicy = s.policyManager.ForLevel(user.Level)

	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, sessionPolicy.Timeouts.ConnectionIdle)

	ctx = policy.ContextWithBufferPolicy(ctx, sessionPolicy.Buffer)
	link, err := dispatcher.Dispatch(ctx, destination)
	if err != nil {
		return newError("failed to dispatch request to ", destination).Base(err)
	}

	requestDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.DownlinkOnly)

		if err := buf.Copy(bodyReader, link.Writer, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to transfer request").Base(err)
		}
		return nil
	}

	responseDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.UplinkOnly)

		var writer buf.Writer
		if destination.Network == net.Network_UDP {
			writer = &PacketWriter{Writer: conn, Target: destination}
		} else {
			writer = buf.NewWriter(conn)
		}

		if err := buf.Copy(link.Reader, writer, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to write response").Base(err)
		}
		return nil
	}

	var requestDonePost = task.OnSuccess(requestDone, task.Close(link.Writer))
	if err := task.Run(ctx, requestDonePost, responseDone); err != nil {
		common.Interrupt(link.Reader)
		common.Interrupt(link.Writer)
		return newError("connection ends").Base(err)
	}
	return nil
}

func (s *Server) fallback(ctx context.Context, pre *buf.Buffer, requestReader buf.Reader, responseWriter buf.Writer) error {
	sessionPolicy := s.policyManager.ForLevel(0)
	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, sessionPolicy.Timeouts.ConnectionIdle)
	ctx = policy.ContextWithBufferPolicy(ctx, sessionPolicy.Buffer)

	var conn net.Conn
	var err error
	fb := s.config.Fallback
	if err := retry.ExponentialBackoff(5, 100).On(func() error {
		var dialer net.Dialer
		conn, err = dialer.DialContext(ctx, fb.Type, fb.Dest)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return newError("failed to dial to " + fb.Dest).Base(err).AtWarning()
	}
	defer conn.Close() // nolint: errcheck

	serverReader := buf.NewReader(conn)
	serverWriter := buf.NewWriter(conn)

	requestDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.DownlinkOnly)

		if err := serverWriter.WriteMultiBuffer(buf.MultiBuffer{pre}); err != nil {
			return newError("failed to fallback request payload").Base(err).AtInfo()
		}

		if err := buf.Copy(requestReader, serverWriter, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to fallback request payload").Base(err).AtInfo()
		}
		return nil
	}

	responseDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.UplinkOnly)
		if err := buf.Copy(serverReader, responseWriter, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to deliver response payload").Base(err).AtInfo()
		}
		return nil
	}

	if err := task.Run(ctx, task.OnSuccess(requestDone, task.Close(serverWriter)), task.OnSuccess(responseDone, task.Close(responseWriter))); err != nil {
		common.Interrupt(serverReader)
		common.Interrupt(serverWriter)
		return newError("fallback ends").Base(err).AtInfo()
	}

	return nil
}
