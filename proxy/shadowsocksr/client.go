package shadowsocksr

import (
    "context"
    "crypto/aes"
    "crypto/cipher"
    "crypto/md5"
    "crypto/sha1"
    "encoding/binary"
    "hash/crc32"
    "math/rand"
    "strings"
    "time"

    core "github.com/v2fly/v2ray-core/v5"
    "github.com/v2fly/v2ray-core/v5/common"
    "github.com/v2fly/v2ray-core/v5/common/buf"
    "github.com/v2fly/v2ray-core/v5/common/crypto"
    "github.com/v2fly/v2ray-core/v5/common/net"
    "github.com/v2fly/v2ray-core/v5/common/protocol"
    "github.com/v2fly/v2ray-core/v5/common/retry"
    "github.com/v2fly/v2ray-core/v5/common/session"
    "github.com/v2fly/v2ray-core/v5/common/signal"
    "github.com/v2fly/v2ray-core/v5/common/task"
    "github.com/v2fly/v2ray-core/v5/features/policy"
    "github.com/v2fly/v2ray-core/v5/transport"
    "github.com/v2fly/v2ray-core/v5/transport/internet"
)

// Client is a ShadowsocksR client
type Client struct {
    serverPicker  protocol.ServerPicker
    policyManager policy.Manager
}

// NewClient creates a new ShadowsocksR client
func NewClient(ctx context.Context, config *ClientConfig) (*Client, error) {
    serverList := protocol.NewServerList()
    for _, rec := range config.Server {
        s, err := protocol.NewServerSpecFromPB(rec)
        if err != nil {
            return nil, newError("failed to parse server spec").Base(err)
        }
        serverList.AddServer(s)
    }
    if serverList.Size() == 0 {
        return nil, newError("0 server")
    }

    v := core.MustFromContext(ctx)
    client := &Client{
        serverPicker:  protocol.NewRoundRobinServerPicker(serverList),
        policyManager: v.GetFeature(policy.ManagerType()).(policy.Manager),
    }
    return client, nil
}

// Process implements OutboundHandler.Process()
func (c *Client) Process(ctx context.Context, link *transport.Link, dialer internet.Dialer) error {
    outbound := session.OutboundFromContext(ctx)
    if outbound == nil || !outbound.Target.IsValid() {
        return newError("target not specified")
    }
    
    destination := outbound.Target
    network := destination.Network

    var server *protocol.ServerSpec
    var conn internet.Connection

    err := retry.ExponentialBackoff(5, 100).On(func() error {
        server = c.serverPicker.PickServer()
        dest := server.Destination()
        dest.Network = network
        rawConn, err := dialer.Dial(ctx, dest)
        if err != nil {
            return err
        }
        conn = rawConn
        return nil
    })

    if err != nil {
        return newError("failed to find an available destination").AtWarning().Base(err)
    }

    defer conn.Close()

    request := &protocol.RequestHeader{
        Version: Version,
        Command: protocol.RequestCommandTCP,
        Address: destination.Address,
        Port:    destination.Port,
        User:    server.PickUser(),
    }
    if destination.Network == net.Network_UDP {
        request.Command = protocol.RequestCommandUDP
    }

    user := request.User
    account := user.Account.(*Account)
    
    // Initialize SSR context
    ssrCtx := &ConnContext{
        Protocol:      initializeProtocol(account),
        Obfs:         initializeObfs(account),
        EncryptMethod: account.CipherType,
        EncryptKey:   passwordToKey(account.Password, getCipherKeyLen(account.CipherType)),
        IV:           make([]byte, getCipherIVLen(account.CipherType)),
        ClientID:     make([]byte, 4),
        ConnectionID: rand.Uint32() % 0xFFFFFF,
        TCPMss:      1460,
        UserKey:     []byte(account.Password),
    }

    rand.Read(ssrCtx.IV)
    rand.Read(ssrCtx.ClientID)
    
    // Initialize encryption
    if err := ssrCtx.initEncryption(); err != nil {
        return newError("failed to initialize encryption").Base(err)
    }

    sessionPolicy := c.policyManager.ForLevel(user.Level)
    ctx, cancel := context.WithCancel(ctx)
    timer := signal.CancelAfterInactivity(ctx, cancel, sessionPolicy.Timeouts.ConnectionIdle)

    // Handle TCP
    if request.Command == protocol.RequestCommandTCP {
        return c.handleTCP(ctx, request, conn, link, timer, sessionPolicy, ssrCtx)
    }

    // Handle UDP
    if request.Command == protocol.RequestCommandUDP {
        return c.handleUDP(ctx, request, conn, link, timer, sessionPolicy, ssrCtx)
    }

    return nil
}

func (c *Client) handleTCP(ctx context.Context, request *protocol.RequestHeader, 
    conn internet.Connection, link *transport.Link, timer *signal.ActivityTimer, 
    sessionPolicy policy.Session, ssrCtx *ConnContext) error {

    requestDone := func() error {
        defer timer.SetTimeout(sessionPolicy.Timeouts.DownlinkOnly)

        // 1. Write SSR handshake
        if err := writeHandshake(request, conn, ssrCtx); err != nil {
            return newError("failed to write handshake").Base(err)
        }

        // 2. Write request data
        bufferedWriter := buf.NewBufferedWriter(buf.NewWriter(conn))
        writer, err := WriteTCPRequest(request, bufferedWriter, ssrCtx)
        if err != nil {
            return newError("failed to write request").Base(err)
        }

        // 3. Write payload
        if err := buf.Copy(link.Reader, writer, buf.UpdateActivity(timer)); err != nil {
            return newError("failed to transfer request").Base(err)
        }

        return nil
    }

    responseDone := func() error {
        defer timer.SetTimeout(sessionPolicy.Timeouts.UplinkOnly)

        reader, err := ReadTCPResponse(request.User, conn, ssrCtx)
        if err != nil {
            return err
        }

        if err := buf.Copy(reader, link.Writer, buf.UpdateActivity(timer)); err != nil {
            return newError("failed to transfer response").Base(err)
        }

        return nil
    }

    responseDoneAndCloseWriter := task.OnSuccess(responseDone, task.Close(link.Writer))
    if err := task.Run(ctx, requestDone, responseDoneAndCloseWriter); err != nil {
        return newError("connection ends").Base(err)
    }

    return nil
}

func (c *Client) handleUDP(ctx context.Context, request *protocol.RequestHeader,
    conn internet.Connection, link *transport.Link, timer *signal.ActivityTimer,
    sessionPolicy policy.Session, ssrCtx *ConnContext) error {

    writer := &UDPWriter{
        Writer:  conn,
        Request: request,
        Context: ssrCtx,
    }

    requestDone := func() error {
        if err := buf.Copy(link.Reader, writer, buf.UpdateActivity(timer)); err != nil {
            return newError("failed to transport all UDP request").Base(err)
        }
        return nil
    }

    responseDone := func() error {
        defer timer.SetTimeout(sessionPolicy.Timeouts.UplinkOnly)

        reader := &UDPReader{
            Reader:  conn,
            Request: request,
            Context: ssrCtx,
        }

        if err := buf.Copy(reader, link.Writer, buf.UpdateActivity(timer)); err != nil {
            return newError("failed to transport all UDP response").Base(err)
        }
        return nil
    }

    responseDoneAndCloseWriter := task.OnSuccess(responseDone, task.Close(link.Writer))
    if err := task.Run(ctx, requestDone, responseDoneAndCloseWriter); err != nil {
        return newError("connection ends").Base(err)
    }

    return nil
}

func writeHandshake(request *protocol.RequestHeader, conn internet.Connection, ctx *ConnContext) error {
    var authData [7]byte
    rand.Read(authData[:1])

    key := append(ctx.IV, ctx.UserKey...)
    hmacData := ctx.hmacMD5(key, authData[:1])
    copy(authData[1:], hmacData[:6])

    if _, err := conn.Write(authData[:]); err != nil {
        return err
    }

    return nil
}

func init() {
    common.Must(common.RegisterConfig((*ClientConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
        return NewClient(ctx, config.(*ClientConfig))
    }))
}
