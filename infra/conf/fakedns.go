package conf

import (
	"github.com/golang/protobuf/proto"
	"v2ray.com/core/app/dns/fakedns"
)

type FakeDnsConfig struct {
	IPPool  string `json:"ipPool"`
	LruSize int64  `json:"poolSize"`
}

func (f FakeDnsConfig) Build() (proto.Message, error) {
	return &fakedns.FakeDnsPool{
		IpPool:  f.IPPool,
		LruSize: f.LruSize,
	}, nil
}

type FakeDnsPostProcessingStage struct {
}

func (f FakeDnsPostProcessingStage) Process(conf *Config) error {
	if conf.DNSConfig != nil && conf.DNSConfig.FakeConfig != nil && *conf.DNSConfig.FakeConfig {
		//Add a Fake DNS Config if there is none
		conf.FakeDns = &FakeDnsConfig{
			IPPool:  "240.0.0.0/8",
			LruSize: 65535,
		}
	}
	found := false
	//Check if there is a Outbound with necessary sniffer on
	var inbounds []InboundDetourConfig

	if conf.InboundConfig != nil {
		inbounds = append(inbounds, *conf.InboundConfig)
	}

	if len(conf.InboundDetours) > 0 {
		inbounds = append(inbounds, conf.InboundDetours...)
	}

	if len(conf.InboundConfigs) > 0 {
		inbounds = append(inbounds, conf.InboundConfigs...)
	}
	for _, v := range inbounds {
		if v.SniffingConfig != nil && v.SniffingConfig.Enabled != true && v.SniffingConfig.DestOverride != nil {
			for _, dov := range *v.SniffingConfig.DestOverride {
				if dov == "fakedns" {
					found = true
				}
			}
		}
	}
	if !found {
		newError("Defined Fake DNS but haven't enabled fake dns sniffing at any inbound.").AtWarning().WriteToLog()
	}
	return nil
}
