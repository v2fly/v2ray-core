package roundtripperreverserserver

import (
	"context"
	"sync"
	"time"

	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request"
)

func NewReverserImpl() (request.ReverserImpl, error) {
	r := &ReverserImpl{
		timeoutDuration: 5 * time.Minute,
	}

	// configure periodic cleaner
	r.periodicCleaner = &task.Periodic{
		Interval: time.Second * 30,
		Execute: func() error {
			now := time.Now()
			r.serverPublicKeyToStateMap.Range(func(k, v interface{}) bool {
				ss, ok := v.(*serverState)
				if !ok {
					return true
				}
				if now.Sub(ss.lastSeen) > r.timeoutDuration {
					r.serverPublicKeyToStateMap.Delete(k)
					r.serverPrivateKeyToPublicKey.Delete(ss.privateKey)
				}
				return true
			})
			return nil
		},
	}

	// start the cleaner; return error if Start fails
	if err := r.periodicCleaner.Start(); err != nil {
		return nil, err
	}

	return r, nil
}

type ReverserImpl struct {
	serverPublicKeyToStateMap   sync.Map
	serverPrivateKeyToPublicKey sync.Map

	clientTemporaryKeyToStateMap sync.Map

	// cleanup fields
	periodicCleaner *task.Periodic
	timeoutDuration time.Duration
}

// StopCleanup stops the periodic cleaner (if running).
func (r *ReverserImpl) StopCleanup() error {
	if r.periodicCleaner == nil {
		return nil
	}
	return r.periodicCleaner.Close()
}

func (r *ReverserImpl) OnOtherRoundTrip(ctx context.Context, req request.Request, opts ...request.RoundTripperOption) (resp request.Response, err error) {
	_ = ctx
	_ = opts
	routingKey := req.ConnectionTag
	if len(routingKey) != 32 {
		return request.Response{}, newError("invalid routing key")
	}
	sourceKey := routingKey[:16]
	destKey := routingKey[16:]

	if _, ok := r.serverPrivateKeyToPublicKey.Load(string(sourceKey)); ok {
		stateInterface, clientOk := r.clientTemporaryKeyToStateMap.Load(string(destKey))
		if clientOk {
			state := stateInterface.(*clientState)
			message := &reverserMessage{
				Data: req.Data,
			}
			select {
			case state.messageQueue <- message:
			default:
				return request.Response{}, newError("client message queue full")
			}
			return request.Response{}, nil
		}
		return request.Response{}, newError("no client found for the given routing key")
	}

	// try if this is a client to server message
	stateInterface, _ := r.clientTemporaryKeyToStateMap.LoadOrStore(string(sourceKey), &clientState{
		messageQueue: make(chan *reverserMessage, 1),
	})
	state := stateInterface.(*clientState)
	defer func() {
		r.clientTemporaryKeyToStateMap.Delete(string(sourceKey))
	}()

	serverStateInterface, ok := r.serverPublicKeyToStateMap.Load(string(destKey))
	if ok {
		serverStateInst := serverStateInterface.(*serverState)
		message := &reverserMessage{
			Data: req.Data,
		}
		timeOutTimer := time.NewTimer(time.Second * 25)
		select {
		case serverStateInst.messageQueue <- message:
		case <-timeOutTimer.C:
			return request.Response{}, newError("server message queue full timeout")
		}

		select {
		case respMessage := <-state.messageQueue:
			timeOutTimer.Stop()
			return request.Response{
				Data: respMessage.Data,
			}, nil
		case <-timeOutTimer.C:
			return request.Response{}, newError("client message queue empty timeout")
		}
	}

	return request.Response{}, newError("no server found for the given routing key")
}

func (r *ReverserImpl) OnAuthenticatedServerIntentRoundTrip(ctx context.Context, serverPublic []byte, req request.Request, opts ...request.RoundTripperOption) (resp request.Response, err error) {
	_ = ctx
	_ = opts
	if len(req.ConnectionTag) != 16 {
		return request.Response{}, newError("invalid server private key")
	}

	if len(serverPublic) != 16 {
		return request.Response{}, newError("invalid server public key")
	}

	serverPrivate := req.ConnectionTag
	// store mapping from private -> public
	r.serverPrivateKeyToPublicKey.Store(string(serverPrivate), string(serverPublic))
	stateInterface, _ := r.serverPublicKeyToStateMap.LoadOrStore(string(serverPublic), &serverState{
		messageQueue: make(chan *reverserMessage, 16),
		lastSeen:     time.Now(),
	})
	state := stateInterface.(*serverState)
	state.lastSeen = time.Now()
	timeOutTimer := time.NewTimer(time.Minute)
	select {
	case message := <-state.messageQueue:
		timeOutTimer.Stop()
		return request.Response{
			Data: message.Data,
		}, nil
	case <-timeOutTimer.C:
		return request.Response{}, nil
	}
}

func (r *ReverserImpl) __CleanupNow____TestOnly() {
	// trigger cleanup immediately for testing purposes
	r.periodicCleaner.Execute()
}

type clientState struct {
	messageQueue chan *reverserMessage
}

type serverState struct {
	messageQueue chan *reverserMessage
	lastSeen     time.Time
	privateKey   string
}

type reverserMessage struct {
	Data []byte
}
