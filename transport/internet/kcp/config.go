package kcp

import (
	"crypto/cipher"

	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/serial"
	"github.com/v2fly/v2ray-core/v5/transport/internet"
)

const protocolName = "mkcp"

// GetMTUValue returns the value of MTU settings.
func (c *Config) GetMTUValue() uint32 {
	if c == nil || c.Mtu == 0 {
		return 1350
	}
	return c.Mtu
}

// GetTTIValue returns the value of TTI settings.
func (c *Config) GetTTIValue() uint32 {
	if c == nil || c.Tti == 0 {
		return 50
	}
	return c.Tti
}

// GetUplinkCapacityValue returns the value of UplinkCapacity settings.
func (c *Config) GetUplinkCapacityValue() uint32 {
	if c == nil || c.UplinkCapacity == 0 {
		return 5
	}
	return c.UplinkCapacity
}

// GetDownlinkCapacityValue returns the value of DownlinkCapacity settings.
func (c *Config) GetDownlinkCapacityValue() uint32 {
	if c == nil || c.DownlinkCapacity == 0 {
		return 20
	}
	return c.DownlinkCapacity
}

// GetWriteBufferSize returns the size of WriterBuffer in bytes.
func (c *Config) GetWriteBufferSize() uint32 {
	if c == nil || c.WriteBuffer == 0 {
		return 2 * 1024 * 1024
	}
	return c.WriteBuffer * 1024 * 1024
}

// GetReadBufferSize returns the size of ReadBuffer in bytes.
func (c *Config) GetReadBufferSize() uint32 {
	if c == nil || c.ReadBuffer == 0 {
		return 2 * 1024 * 1024
	}
	return c.ReadBuffer * 1024 * 1024
}

// GetSecurity returns the security settings.
func (c *Config) GetSecurity() (cipher.AEAD, error) {
	if c.Seed != "" {
		return NewAEADAESGCMBasedOnSeed(c.Seed), nil
	}
	return NewSimpleAuthenticator(), nil
}

func (c *Config) GetPackerHeader() (internet.PacketHeader, error) {
	if c.HeaderConfig != nil {
		rawConfig, err := serial.GetInstanceOf(c.HeaderConfig)
		if err != nil {
			return nil, err
		}

		return internet.CreatePacketHeader(rawConfig)
	}
	return nil, nil
}

func (c *Config) GetSendingInFlightSize() uint32 {
	size := c.GetUplinkCapacityValue() * 1024 * 1024 / c.GetMTUValue() / (1000 / c.GetTTIValue())
	if size < 8 {
		size = 8
	}
	return size
}

func (c *Config) GetSendingBufferSize() uint32 {
	return c.GetWriteBufferSize() / c.GetMTUValue()
}

func (c *Config) GetReceivingInFlightSize() uint32 {
	size := c.GetDownlinkCapacityValue() * 1024 * 1024 / c.GetMTUValue() / (1000 / c.GetTTIValue())
	if size < 8 {
		size = 8
	}
	return size
}

func (c *Config) GetReceivingBufferSize() uint32 {
	return c.GetReadBufferSize() / c.GetMTUValue()
}

func init() {
	common.Must(internet.RegisterProtocolConfigCreator(protocolName, func() interface{} {
		return new(Config)
	}))
}
