package shadowsocksr

import (
    "context"
    "time"
    
    core "github.com/v2fly/v2ray-core/v5"
    "github.com/v2fly/v2ray-core/v5/common"
    "github.com/v2fly/v2ray-core/v5/common/buf"
    "github.com/v2fly/v2ray-core/v5/common/log"
    "github.com/v2fly/v2ray-core/v5/common/net"
    "github.com/v2fly/v2ray-core/v5/common/protocol"
    udp_proto "github.com/v2fly/v2ray-core/v5/common/protocol/udp"
    "github.com/v2fly/v2ray-core/v5/common/session"
    "github.com/v2fly/v2ray-core/v5/common/signal"
    "github.com/v2fly/v2ray-core/v5/common/task"
    "github.com/v2fly/v2ray-core/v5/features/policy"
    "github.com/v2fly/v2ray-core/v5/features/routing"
    "github.com/v2fly/v2ray-core/v5/transport/internet"
    "github.com/v2fly/v2ray-core/v5/transport/internet/udp"
)

type Server struct {
    config        *ServerConfig
    user          *protocol.MemoryUser
    policyManager policy.Manager
    // SSR specific fields
    ssrContext    *ConnContext
}

func NewServer(ctx context.Context, config *ServerConfig) (*Server, error) {
    if config.GetUser() == nil {
        return nil, newError("user is not specified")
    }

    mUser, err := config.User.ToMemoryUser()
    if err != nil {
        return nil, newError("failed to parse user account").Base(err)
    }

    v := core.MustFromContext(ctx)
    s := &Server{
        config:        config,
        user:          mUser,
        policyManager: v.GetFeature(policy.ManagerType()).(policy.Manager),
    }

    // Initialize SSR context
    account := config.User.Account.(*Account)
    s.ssrContext = &ConnContext{
        Protocol:      initializeProtocol(account),
        Obfs:         initializeObfs(account),
        EncryptMethod: account.CipherType,
        EncryptKey:   passwordToKey(account.Password, getCipherKeyLen(account.CipherType)),
        UserKey:      []byte(account.Password),
    }

    return s, nil
}

func (s *Server) Network() []net.Network {
    list := s.config.Network
    if len(list) == 0 {
        list = append(list, net.Network_TCP)
    }
    if s.config.UdpEnabled {
        list = append(list, net.Network_UDP)
    }
    return list
}

func (s *Server) Process(ctx context.Context, network net.Network, conn internet.Connection, dispatcher routing.Dispatcher) error {
    switch network {
    case net.Network_TCP:
        return s.handleConnection(ctx, conn, dispatcher)
    case net.Network_UDP:
        return s.handlerUDPPayload(ctx, conn, dispatcher)
    default:
        return newError("unknown network: ", network)
    }
}

func (s *Server) handleConnection(ctx context.Context, conn internet.Connection, dispatcher routing.Dispatcher) error {
    sessionPolicy := s.policyManager.ForLevel(s.user.Level)
    conn.SetReadDeadline(time.Now().Add(sessionPolicy.Timeouts.Handshake))

    // Read and verify SSR handshake
    if err := s.readHandshake(conn); err != nil {
        log.Record(&log.AccessMessage{
            From:   conn.RemoteAddr(),
            To:     "",
            Status: log.AccessRejected,
            Reason: err,
        })
        return newError("failed SSR handshake from: ", conn.RemoteAddr()).Base(err)
    }

    bufferedReader := &buf.BufferedReader{
        Reader: &SSRReader{
            Reader:  buf.NewReader(conn),
            Context: s.ssrContext,
        },
    }

    request, bodyReader, err := ReadTCPSession(s.user, bufferedReader)
    if err != nil {
        log.Record(&log.AccessMessage{
            From:   conn.RemoteAddr(),
            To:     "",
            Status: log.AccessRejected,
            Reason: err,
        })
        return newError("failed to create request from: ", conn.RemoteAddr()).Base(err)
    }
    conn.SetReadDeadline(time.Time{})

    inbound := session.InboundFromContext(ctx)
    if inbound == nil {
        panic("no inbound metadata")
    }
    inbound.User = s.user

    dest := request.Destination()
    ctx = log.ContextWithAccessMessage(ctx, &log.AccessMessage{
        From:   conn.RemoteAddr(),
        To:     dest,
        Status: log.AccessAccepted,
        Reason: "",
        Email:  request.User.Email,
    })
    newError("tunnelling request to ", dest).WriteToLog(session.ExportIDToError(ctx))

    ctx, cancel := context.WithCancel(ctx)
    timer := signal.CancelAfterInactivity(ctx, cancel, sessionPolicy.Timeouts.ConnectionIdle)

    ctx = policy.ContextWithBufferPolicy(ctx, sessionPolicy.Buffer)
    link, err := dispatcher.Dispatch(ctx, dest)
    if err != nil {
        return err
    }

    responseDone := func() error {
        defer timer.SetTimeout(sessionPolicy.Timeouts.UplinkOnly)

        bufferedWriter := buf.NewBufferedWriter(&SSRWriter{
            Writer:  buf.NewWriter(conn),
            Context: s.ssrContext,
        })

        responseWriter, err := WriteTCPResponse(request, bufferedWriter)
        if err != nil {
            return newError("failed to write response").Base(err)
        }

        if err := buf.Copy(link.Reader, responseWriter, buf.UpdateActivity(timer)); err != nil {
            return newError("failed to transport all TCP response").Base(err)
        }

        return nil
    }

    requestDone := func() error {
        defer timer.SetTimeout(sessionPolicy.Timeouts.DownlinkOnly)

        if err := buf.Copy(bodyReader, link.Writer, buf.UpdateActivity(timer)); err != nil {
            return newError("failed to transport all TCP request").Base(err)
        }

        return nil
    }

    requestDoneAndCloseWriter := task.OnSuccess(requestDone, task.Close(link.Writer))
    if err := task.Run(ctx, requestDoneAndCloseWriter, responseDone); err != nil {
        common.Interrupt(link.Reader)
        common.Interrupt(link.Writer)
        return newError("connection ends").Base(err)
    }

    return nil
}

func (s *Server) readHandshake(conn internet.Connection) error {
    buffer := make([]byte, 7) // SSR handshake size
    if _, err := io.ReadFull(conn, buffer); err != nil {
        return newError("failed to read handshake").Base(err)
    }

    // Verify HMAC
    key := append(s.ssrContext.IV, s.ssrContext.UserKey...)
    hmacData := s.ssrContext.hmacMD5(key, buffer[:1])
    if !hmac.Equal(hmacData[:6], buffer[1:7]) {
        return newError("invalid handshake authentication")
    }

    return nil
}

// SSR specific reader/writer
type SSRReader struct {
    Reader  buf.Reader
    Context *ConnContext
}

type SSRWriter struct {
    Writer  buf.Writer
    Context *ConnContext
}

func (r *SSRReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
    mb, err := r.Reader.ReadMultiBuffer()
    if err != nil {
        return nil, err
    }

    // Process with protocol and obfs
    if r.Context.Protocol != nil {
        mb, err = r.Context.Protocol.DecodePacket(mb)
        if err != nil {
            return nil, err
        }
    }

    if r.Context.Obfs != nil {
        mb, err = r.Context.Obfs.ServerDecode(mb)
        if err != nil {
            return nil, err
        }
    }

    return mb, nil
}

func (w *SSRWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
    // Process with protocol and obfs
    if w.Context.Protocol != nil {
        mb, err := w.Context.Protocol.EncodePacket(mb)
        if err != nil {
            return err
        }
    }

    if w.Context.Obfs != nil {
        mb, err := w.Context.Obfs.ServerEncode(mb)
        if err != nil {
            return err
        }
    }

    return w.Writer.WriteMultiBuffer(mb)
}

func init() {
    common.Must(common.RegisterConfig((*ServerConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
        return NewServer(ctx, config.(*ServerConfig))
    }))
}
