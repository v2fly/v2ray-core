package outbound

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"hash/crc64"
	"sync"

	core "github.com/v2fly/v2ray-core/v5"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/buf"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/platform"
	"github.com/v2fly/v2ray-core/v5/common/protocol"
	"github.com/v2fly/v2ray-core/v5/common/retry"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/common/session"
	"github.com/v2fly/v2ray-core/v5/common/signal"
	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/features/policy"
	"github.com/v2fly/v2ray-core/v5/proxy"
	"github.com/v2fly/v2ray-core/v5/proxy/vmess"
	"github.com/v2fly/v2ray-core/v5/proxy/vmess/encoding"
	"github.com/v2fly/v2ray-core/v5/transport"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

type serverByDestination struct {
	sync.Mutex
	cache map[net.Destination]*protocol.ServerSpec
}

func newServerByDestination() *serverByDestination {
	return &serverByDestination{
		cache: make(map[net.Destination]*protocol.ServerSpec),
	}
}

func (v *serverByDestination) addNoLock(u *protocol.ServerSpec) bool {
	if _, found := v.cache[u.Destination()]; found {
		return false
	}
	v.cache[u.Destination()] = u
	return true
}

func (v *serverByDestination) Add(u *protocol.ServerSpec) bool {
	v.Lock()
	defer v.Unlock()

	return v.addNoLock(u)
}

func (v *serverByDestination) Get(destination net.Destination) (*protocol.ServerSpec, bool) {
	v.Lock()
	defer v.Unlock()

	server, found := v.cache[destination]
	if !found {
		return nil, false
	}
	return server, true
}

func (v *serverByDestination) Remove(destination net.Destination) bool {
	v.Lock()
	defer v.Unlock()

	if _, found := v.cache[destination]; !found {
		return false
	}
	delete(v.cache, destination)
	return true
}

// Handler is an outbound connection handler for VMess protocol.
type Handler struct {
	serverList          *protocol.ServerList
	serverByDestination *serverByDestination
	serverPicker        protocol.ServerPicker
	policyManager       policy.Manager
}

// New creates a new VMess outbound handler.
func New(ctx context.Context, config *Config) (*Handler, error) {
	serverList := protocol.NewServerList()
	serverCache := newServerByDestination()
	for _, rec := range config.Receiver {
		s, err := protocol.NewServerSpecFromPB(rec)
		if err != nil {
			return nil, newError("failed to parse server spec").Base(err)
		}
		if !serverCache.Add(s) {
			return nil, newError("duplicated destination: ", s.Destination())
		}
		serverList.AddServer(s)
	}

	v := core.MustFromContext(ctx)
	handler := &Handler{
		serverList:          serverList,
		serverByDestination: serverCache,
		serverPicker:        protocol.NewRoundRobinServerPicker(serverList),
		policyManager:       v.GetFeature(policy.ManagerType()).(policy.Manager),
	}

	return handler, nil
}

// Process implements proxy.Outbound.Process().
func (h *Handler) Process(ctx context.Context, link *transport.Link, dialer internet.Dialer) error {
	var rec *protocol.ServerSpec
	var conn internet.Connection

	err := retry.ExponentialBackoff(5, 200).On(func() error {
		rec = h.serverPicker.PickServer()
		rawConn, err := dialer.Dial(ctx, rec.Destination())
		if err != nil {
			return err
		}
		conn = rawConn

		return nil
	})
	if err != nil {
		return newError("failed to find an available destination").Base(err).AtWarning()
	}
	defer conn.Close()

	outbound := session.OutboundFromContext(ctx)
	if outbound == nil || !outbound.Target.IsValid() {
		return newError("target not specified").AtError()
	}

	target := outbound.Target
	newError("tunneling request to ", target, " via ", rec.Destination().NetAddr()).WriteToLog(session.ExportIDToError(ctx))

	command := protocol.RequestCommandTCP
	if target.Network == net.Network_UDP {
		command = protocol.RequestCommandUDP
	}
	if target.Address.Family().IsDomain() && target.Address.Domain() == "v1.mux.cool" {
		command = protocol.RequestCommandMux
	}

	user := rec.PickUser()
	request := &protocol.RequestHeader{
		Version: encoding.Version,
		User:    user,
		Command: command,
		Address: target.Address,
		Port:    target.Port,
		Option:  protocol.RequestOptionChunkStream,
	}

	account := request.User.Account.(*vmess.MemoryAccount)
	request.Security = account.Security

	if request.Security == protocol.SecurityType_AES128_GCM || request.Security == protocol.SecurityType_NONE || request.Security == protocol.SecurityType_CHACHA20_POLY1305 {
		request.Option.Set(protocol.RequestOptionChunkMasking)
	}

	if shouldEnablePadding(request.Security) && request.Option.Has(protocol.RequestOptionChunkMasking) {
		request.Option.Set(protocol.RequestOptionGlobalPadding)
	}

	if request.Security == protocol.SecurityType_ZERO {
		request.Security = protocol.SecurityType_NONE
		request.Option.Clear(protocol.RequestOptionChunkStream)
		request.Option.Clear(protocol.RequestOptionChunkMasking)
	}

	if account.AuthenticatedLengthExperiment {
		request.Option.Set(protocol.RequestOptionAuthenticatedLength)
	}

	input := link.Reader
	output := link.Writer

	isAEAD := false
	if !aeadDisabled && len(account.AlterIDs) == 0 {
		isAEAD = true
	}

	hashkdf := hmac.New(sha256.New, []byte("VMessBF"))
	hashkdf.Write(account.ID.Bytes())

	behaviorSeed := crc64.Checksum(hashkdf.Sum(nil), crc64.MakeTable(crc64.ISO))

	session := encoding.NewClientSession(ctx, isAEAD, protocol.DefaultIDHash, int64(behaviorSeed))
	sessionPolicy := h.policyManager.ForLevel(request.User.Level)

	ctx, cancel := context.WithCancel(ctx)
	timer := signal.CancelAfterInactivity(ctx, cancel, sessionPolicy.Timeouts.ConnectionIdle)

	requestDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.DownlinkOnly)

		writer := buf.NewBufferedWriter(buf.NewWriter(conn))
		if err := session.EncodeRequestHeader(request, writer); err != nil {
			return newError("failed to encode request").Base(err).AtWarning()
		}

		bodyWriter, err := session.EncodeRequestBody(request, writer)
		if err != nil {
			return newError("failed to start encoding").Base(err)
		}
		if err := buf.CopyOnceTimeout(input, bodyWriter, proxy.FirstPayloadTimeout); err != nil && err != buf.ErrNotTimeoutReader && err != buf.ErrReadTimeout {
			return newError("failed to write first payload").Base(err)
		}

		if err := writer.SetBuffered(false); err != nil {
			return err
		}

		if err := buf.Copy(input, bodyWriter, buf.UpdateActivity(timer)); err != nil {
			return err
		}

		if request.Option.Has(protocol.RequestOptionChunkStream) && !account.NoTerminationSignal {
			if err := bodyWriter.WriteMultiBuffer(buf.MultiBuffer{}); err != nil {
				return err
			}
		}

		return nil
	}

	responseDone := func() error {
		defer timer.SetTimeout(sessionPolicy.Timeouts.UplinkOnly)

		reader := &buf.BufferedReader{Reader: buf.NewReader(conn)}
		header, err := session.DecodeResponseHeader(reader)
		if err != nil {
			return newError("failed to read header").Base(err)
		}
		h.handleCommand(rec.Destination(), header.Command)

		bodyReader, err := session.DecodeResponseBody(request, reader)
		if err != nil {
			return newError("failed to start encoding response").Base(err)
		}
		return buf.Copy(bodyReader, output, buf.UpdateActivity(timer))
	}

	responseDonePost := task.OnSuccess(responseDone, task.Close(output))
	if err := task.Run(ctx, requestDone, responseDonePost); err != nil {
		return newError("connection ends").Base(err)
	}

	return nil
}

func (h *Handler) AddServer(ctx context.Context, server *protocol.ServerSpec) error {
	if server.Destination().IsValid() && !h.serverByDestination.Add(server) {
		return newError("Server ", server.Destination().String(), " already exists.")
	}
	h.serverList.AddServer(server)
	return nil
}

func (h *Handler) RemoveServer(ctx context.Context, destination net.Destination) error {
	if !destination.IsValid() {
		return newError("Destination must not be valid.")
	}
	if !h.serverByDestination.Remove(destination) {
		return newError("Destination ", destination.String(), " not found.")
	}
	h.serverList.RemoveServer(destination)
	return nil
}

var (
	enablePadding = false
	aeadDisabled  = false
)

func shouldEnablePadding(s protocol.SecurityType) bool {
	return enablePadding || s == protocol.SecurityType_AES128_GCM || s == protocol.SecurityType_CHACHA20_POLY1305 || s == protocol.SecurityType_AUTO
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*Config))
	}))

	common.Must(common.RegisterConfig((*SimplifiedConfig)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		simplifiedClient := config.(*SimplifiedConfig)
		fullClient := &Config{Receiver: []*protocol.ServerEndpoint{
			{
				Address: simplifiedClient.Address,
				Port:    simplifiedClient.Port,
				User: []*protocol.User{
					{
						Account: serial.ToTypedMessage(&vmess.Account{Id: simplifiedClient.Uuid}),
					},
				},
			},
		}}

		return common.CreateObject(ctx, fullClient)
	}))

	const defaultFlagValue = "NOT_DEFINED_AT_ALL"

	paddingValue := platform.NewEnvFlag("v2ray.vmess.padding").GetValue(func() string { return defaultFlagValue })
	if paddingValue != defaultFlagValue {
		enablePadding = true
	}

	isAeadDisabled := platform.NewEnvFlag("v2ray.vmess.aead.disabled").GetValue(func() string { return defaultFlagValue })
	if isAeadDisabled == "true" {
		aeadDisabled = true
	}
}
