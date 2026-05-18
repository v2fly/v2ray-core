package gdocsviewer

import (
	"bytes"
	"testing"

	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request/assembler/simple"
	roundtripper "github.com/v2fly/v2ray-core/v5/transport/internet/request/roundtripper/gdocsviewer"
)

func TestBuildClientRequestConfig(t *testing.T) {
	key := bytes.Repeat([]byte{8}, 32)
	config := buildClientRequestConfig(&Config{
		ViewerUrl:            "https://viewer.example/viewer",
		TextUrl:              "https://viewer.example/viewerng/text",
		OriginUrl:            "https://origin.example/g",
		ViewerHostHeader:     "docs.example",
		UserAgent:            "ua",
		AllowHttp:            true,
		H2PoolSize:           4,
		MaxViewerBodyBytes:   1234,
		MinRequestIntervalMs: 5,
		MaxRequestBytes:      777,
		SharedKey:            key,
		OriginUrlReplacementRules: []*OriginUrlReplacementRule{{
			Name:    "abc",
			Pattern: "[a-z0-9]{14}",
		}},
		RequestHeaders: map[string]string{
			"Accept-Language": "en-US,en;q=0.9",
		},
	})

	assembler, err := serial.GetInstanceOf(config.Assembler)
	if err != nil {
		t.Fatal(err)
	}
	simpleConfig := assembler.(*simple.ClientConfig)
	if simpleConfig.MaxWriteSize != 777 ||
		simpleConfig.InitialPollingIntervalMs != 5 ||
		simpleConfig.MinPollingIntervalMs != 5 ||
		simpleConfig.MaxPollingIntervalMs != defaultMaxPollingIntervalMs ||
		simpleConfig.FailedRetryIntervalMs != defaultFailedRetryIntervalMs {
		t.Fatalf("unexpected simple client config %+v", simpleConfig)
	}

	rt, err := serial.GetInstanceOf(config.Roundtripper)
	if err != nil {
		t.Fatal(err)
	}
	rtConfig := rt.(*roundtripper.ClientConfig)
	if rtConfig.OriginUrl != "https://origin.example/g" ||
		rtConfig.ViewerHostHeader != "docs.example" ||
		rtConfig.UserAgent != "ua" ||
		!rtConfig.AllowHttp ||
		rtConfig.H2PoolSize != 4 ||
		rtConfig.MaxViewerBodyBytes != 1234 ||
		rtConfig.MinRequestIntervalMs != 5 ||
		!bytes.Equal(rtConfig.SharedKey, key) ||
		len(rtConfig.OriginUrlReplacementRules) != 1 ||
		rtConfig.OriginUrlReplacementRules[0].Name != "abc" ||
		rtConfig.OriginUrlReplacementRules[0].Pattern != "[a-z0-9]{14}" ||
		rtConfig.RequestHeaders["Accept-Language"] != "en-US,en;q=0.9" {
		t.Fatalf("unexpected gdocsviewer client config %+v", rtConfig)
	}
}

func TestBuildServerRequestConfig(t *testing.T) {
	key := bytes.Repeat([]byte{8}, 32)
	config := buildServerRequestConfig(&Config{
		PathPrefix:       "/g",
		MaxRequestBytes:  777,
		MaxResponseBytes: 888,
		SharedKey:        key,
	})

	assembler, err := serial.GetInstanceOf(config.Assembler)
	if err != nil {
		t.Fatal(err)
	}
	simpleConfig := assembler.(*simple.ServerConfig)
	if simpleConfig.MaxWriteSize != 888 || simpleConfig.PollingResponseWaitMs != defaultPollingResponseWaitMs {
		t.Fatalf("unexpected simple server config %+v", simpleConfig)
	}

	rt, err := serial.GetInstanceOf(config.Roundtripper)
	if err != nil {
		t.Fatal(err)
	}
	rtConfig := rt.(*roundtripper.ServerConfig)
	if rtConfig.PathPrefix != "/g" ||
		rtConfig.MaxRequestBytes != 777 ||
		rtConfig.MaxResponseBytes != 888 ||
		!bytes.Equal(rtConfig.SharedKey, key) {
		t.Fatalf("unexpected gdocsviewer server config %+v", rtConfig)
	}
}
