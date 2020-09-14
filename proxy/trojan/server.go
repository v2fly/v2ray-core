package trojan

import (
	"context"
	"io"
	"time"

	"v2ray.com/core"
	"v2ray.com/core/common"
	"v2ray.com/core/common/buf"
	"v2ray.com/core/common/errors"
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

	buffer := buf.New()
	defer buffer.Release()

	n, err := buffer.ReadFrom(conn)
	if err != nil {
		return newError("failed to read first request").Base(err)
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

	bufferedReader := &buf.BufferedReader{
		Reader: buf.NewReader(conn),
		Buffer: buf.MultiBuffer{buffer},
	}

	if fallbackEnabled && shouldFallback {
		return s.fallback(ctx, sessionPolicy, bufferedReader, buf.NewWriter(conn))
	} else if shouldFallback {
		return newError("invalid protocol or invalid user")
	}

	dest, err := ReadHeader(bufferedReader)
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
	sessionPolicy = s.policyManager.ForLevel(user.Level)

	if destination.Network == net.Network_UDP { // handle udp request
		for {
			dest, mb, err := ReadPacket(bufferedReader)
			if dest != nil && !mb.IsEmpty() {
				destination := *dest
				log.ContextWithAccessMessage(ctx, &log.AccessMessage{
					From:   conn.RemoteAddr(),
					To:     destination,
					Status: log.AccessAccepted,
					Reason: "",
					Email:  user.Email,
				})
				newError("received request for ", destination).WriteToLog(session.ExportIDToError(ctx))

				// send every udp packet seperately
				werr := s.transferRequest(ctx, sessionPolicy, dispatcher, destination, &buf.MultiBufferContainer{MultiBuffer: mb}, &PacketWriter{Writer: conn, Target: destination})
				if werr != nil {
					return werr
				}
			}

			if err != nil {
				if errors.Cause(err) != io.EOF {
					return err
				}

				return nil
			}
		}
	} else { // handle tcp request

		log.ContextWithAccessMessage(ctx, &log.AccessMessage{
			From:   conn.RemoteAddr(),
			To:     destination,
			Status: log.AccessAccepted,
			Reason: "",
			Email:  user.Email,
		})

		newError("received request for ", destination).WriteToLog(session.ExportIDToError(ctx))
		return s.transferRequest(ctx, sessionPolicy, dispatcher, destination, bufferedReader, buf.NewWriter(conn))
	}
}

func (s *Server) transferRequest(ctx context.Context, sessionPolicy policy.Session,
	dispatcher routing.Dispatcher,
	destination net.Destination,
	clientReader buf.Reader,
	clientWriter buf.Writer) error {

	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, sessionPolicy.Timeouts.ConnectionIdle)
	ctx = policy.ContextWithBufferPolicy(ctx, sessionPolicy.Buffer)

	link, err := dispatcher.Dispatch(ctx, destination)
	if err != nil {
		return newError("failed to dispatch request to ", destination).Base(err)
	}

	requestDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.DownlinkOnly)

		if err := buf.Copy(clientReader, link.Writer, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to transfer request").Base(err)
		}
		return nil
	}

	responseDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.UplinkOnly)

		if err := buf.Copy(link.Reader, clientWriter, buf.UpdateActivity(timer)); err != nil {
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

func (s *Server) fallback(ctx context.Context, sessionPolicy policy.Session, requestReader buf.Reader, responseWriter buf.Writer) error {
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
