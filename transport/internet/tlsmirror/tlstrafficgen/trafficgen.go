package tlstrafficgen

import (
	"context"
	cryptoRand "crypto/rand"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"time"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/net"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet/security"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror"
	"github.com/v2fly/v2ray-core/v5/transport/internet/tlsmirror/httponconnection"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

type TrafficGenerator struct {
	config *Config
	ctx    context.Context

	destination net.Destination
	tag         string
}

func NewTrafficGenerator(ctx context.Context, config *Config, destination net.Destination, tag string) *TrafficGenerator {
	return &TrafficGenerator{
		ctx:         ctx,
		config:      config,
		destination: destination,
		tag:         tag,
	}
}

type trafficGeneratorManagedConnectionController struct {
	readyCtx  context.Context
	readyDone context.CancelFunc

	recallCtx  context.Context
	recallDone context.CancelFunc

	invalidatedCtx  context.Context
	invalidatedDone context.CancelFunc
}

func newTrafficGeneratorManagedConnectionController(parent context.Context) *trafficGeneratorManagedConnectionController {
	readyCtx, readyDone := context.WithCancel(parent)
	recallCtx, recallDone := context.WithCancel(parent)
	invalidatedCtx, invalidatedDone := context.WithCancel(parent)
	return &trafficGeneratorManagedConnectionController{
		readyCtx:        readyCtx,
		readyDone:       readyDone,
		recallCtx:       recallCtx,
		recallDone:      recallDone,
		invalidatedCtx:  invalidatedCtx,
		invalidatedDone: invalidatedDone,
	}
}

func (t *trafficGeneratorManagedConnectionController) WaitConnectionReady() context.Context {
	return t.readyCtx
}

func (t *trafficGeneratorManagedConnectionController) RecallTrafficGenerator() error {
	t.recallDone()
	return nil
}

func (t *trafficGeneratorManagedConnectionController) IsConnectionInvalidated() bool {
	return t.invalidatedCtx.Err() != nil
}

// Copied from https://brandur.org/fragments/crypto-rand-float64, Thanks
func randIntN(max int64) int64 {
	nBig, err := cryptoRand.Int(cryptoRand.Reader, big.NewInt(max))
	if err != nil {
		panic(err)
	}
	return nBig.Int64()
}

func randFloat64() float64 {
	return float64(randIntN(1<<53)) / (1 << 53)
}

func (generator *TrafficGenerator) GenerateNextTraffic(ctx context.Context) error {
	transportEnvironment := envctx.EnvironmentFromContext(generator.ctx).(environment.TransportEnvironment)
	dialer := transportEnvironment.OutboundDialer()

	trafficController := newTrafficGeneratorManagedConnectionController(generator.ctx)

	var trafficControllerIfce tlsmirror.TrafficGeneratorManagedConnection = trafficController
	managedConnectionContextValue := context.WithValue(generator.ctx,
		tlsmirror.TrafficGeneratorManagedConnectionContextKey, trafficControllerIfce) // nolint:staticcheck

	defer func() {
		trafficController.invalidatedDone()
		trafficController.readyDone()
	}()

	conn, err := dialer(managedConnectionContextValue, generator.destination, generator.tag)
	if err != nil {
		return newError("failed to dial to destination").Base(err).AtWarning()
	}
	tlsConn, err := generator.tlsHandshake(conn)
	if err != nil {
		return newError("failed to create TLS connection").Base(err).AtWarning()
	}
	getAlpn, ok := tlsConn.(security.ConnectionApplicationProtocol)
	if !ok {
		return newError("TLS connection does not support getting ALPN").AtWarning()
	}
	alpn, err := getAlpn.GetConnectionApplicationProtocol()
	if err != nil {
		return newError("failed to get ALPN from TLS connection").Base(err).AtWarning()
	}
	steps := generator.config.Steps
	currentStep := 0
	httpRoundTripper, err := httponconnection.NewSingleConnectionHTTPTransport(tlsConn, alpn)
	if err != nil {
		return newError("failed to create HTTP transport").Base(err).AtWarning()
	}
	for {
		if currentStep >= len(steps) {
			return tlsConn.Close()
		}

		step := steps[currentStep]

		url := url.URL{
			Scheme: "https",
			Host:   step.Host,
			Path:   step.Path,
		}

		httpReq := &http.Request{Host: url.Hostname(), Method: step.Method, URL: &url}

		if step.Headers != nil {
			header := make(http.Header, len(step.Headers))
			for _, v := range step.Headers {
				if v.Value != "" {
					header.Add(v.Name, v.Value)
				}
				if v.Values != nil {
					for _, value := range v.Values {
						header.Add(v.Name, value)
					}
				}
			}
			httpReq.Header = header
		}

		startTime := time.Now()

		resp, err := httpRoundTripper.RoundTrip(httpReq) //nolint:bodyclose
		if err != nil {
			return newError("failed to send HTTP request").Base(err).AtWarning()
		}

		finishRequest := func() error {
			_, err = io.Copy(io.Discard, resp.Body)
			if err != nil {
				return newError("failed to read HTTP response body").Base(err).AtWarning()
			}
			err = resp.Body.Close()
			if err != nil {
				return newError("failed to close HTTP response body").Base(err).AtWarning()
			}
			return nil
		}

		if step.H2DoNotWaitForDownloadFinish && alpn == "h2" {
			go func() {
				if err := finishRequest(); err != nil {
					newError("failed to finish request in background").Base(err).AtWarning().WriteToLog()
				}
			}()
		} else {
			if err := finishRequest(); err != nil {
				return err
			}
		}

		endTime := time.Now()

		eclipsedTime := endTime.Sub(startTime)
		if step.WaitTime == nil {
			step.WaitTime = &TimeSpec{}
			newError("no wait time specified for step ", currentStep, ", using default 0 seconds").AtWarning().WriteToLog()
		}
		secondToWait := (float64(step.WaitTime.UniformRandomMultiplierNanoseconds)*randFloat64() + float64(step.WaitTime.BaseNanoseconds)) / float64(time.Second)
		if eclipsedTime < time.Duration(secondToWait*float64(time.Second)) {
			waitTime := time.Duration(secondToWait*float64(time.Second)) - eclipsedTime
			newError("waiting for ", waitTime, " after step ", currentStep).AtDebug().WriteToLog()
			waitTimer := time.NewTimer(waitTime)
			select {
			case <-ctx.Done():
				waitTimer.Stop()
				return ctx.Err()
			case <-waitTimer.C:
			}
		} else {
			newError("step ", currentStep, " took too long: ", eclipsedTime, ", expected: ", secondToWait, " seconds").AtWarning().WriteToLog()
		}

		if step.ConnectionReady {
			trafficController.readyDone()
			newError("connection ready for payload traffic").AtInfo().WriteToLog()
		}

		if step.ConnectionRecallExit {
			if trafficController.recallCtx.Err() != nil {
				return tlsConn.Close()
			}
		}

		if step.NextStep == nil {
			currentStep++
		} else {
			overallWeight := int32(0)
			for _, nextStep := range step.NextStep {
				overallWeight += nextStep.Weight
			}
			maxBound := big.NewInt(int64(overallWeight))
			selectionValue, err := cryptoRand.Int(cryptoRand.Reader, maxBound)
			if err != nil {
				return newError("failed to generate random selection value").Base(err).AtWarning()
			}
			selectedValue := int32(selectionValue.Int64())
			currentValue := int32(0)
			matched := false
			for _, nextStep := range step.NextStep {
				if currentValue >= selectedValue {
					currentStep = int(nextStep.GotoLocation)
					matched = true
					break
				}
				currentValue += nextStep.Weight
			}
			if !matched {
				newError("invalid steps jump instruction, check configuration for step ", currentStep).AtError().WriteToLog()
				currentStep++
			}
		}
	}
}

func (generator *TrafficGenerator) tlsHandshake(conn net.Conn) (security.Conn, error) {
	securitySettingInstance, err := serial.GetInstanceOf(generator.config.SecuritySettings)
	if err != nil {
		return nil, newError("failed to get instance of security settings").Base(err)
	}
	securityEngine, err := common.CreateObject(generator.ctx, securitySettingInstance)
	if err != nil {
		return nil, newError("unable to create security engine from security settings").Base(err)
	}
	securityEngineTyped, ok := securityEngine.(security.Engine)
	if !ok {
		return nil, newError("type assertion error when create security engine from security settings")
	}

	return securityEngineTyped.Client(conn)
}
