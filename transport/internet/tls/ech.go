//go:build go1.23
// +build go1.23

package tls

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"io"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/miekg/dns"
)

func ApplyECH(c *Config, config *tls.Config) error {
	var ECHConfig []byte
	var err error

	if len(c.EchConfig) > 0 {
		ECHConfig, err = base64.StdEncoding.DecodeString(c.EchConfig)
		if err != nil {
			return newError("invalid ECH config")
		}
	} else { // ECH config > DOH lookup
		if c.ServerName == "" {
			return newError("Using DOH for ECH needs serverName")
		}
		ECHRecord, err := QueryRecord(c.ServerName, c.Ech_DOHserver)
		if err != nil {
			return err
		}
		ECHConfig, _ = base64.StdEncoding.DecodeString(ECHRecord)
	}

	config.EncryptedClientHelloConfigList = ECHConfig
	return nil
}

type record struct {
	record string
	expire time.Time
}

var (
	dnsCache = make(map[string]record)
	mutex    sync.RWMutex
)

func QueryRecord(domain string, server string) (string, error) {
	rec, found := dnsCache[domain]
	if found && rec.expire.After(time.Now()) {
		return rec.record, nil
	}
	mutex.Lock()
	defer mutex.Unlock()
	record, ttl, err := dohQuery(server, domain)
	if err != nil {
		return "", err
	}
	// Use TTL for good, but many HTTPS records have TTL 60, too short
	if ttl < 600 {
		ttl = 600
	}
	rec.record = record
	rec.expire = time.Now().Add(time.Second * time.Duration(ttl))
	dnsCache[domain] = rec
	return record, nil
}

func dohQuery(server string, domain string) (string, uint32, error) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypeHTTPS)
	msg, err := m.Pack()
	if err != nil {
		return "", 0, err
	}
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("POST", server, bytes.NewReader(msg))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/dns-message")
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return "", 0, newError("query failed with response code:", resp.StatusCode)
	}
	respMsg := new(dns.Msg)
	err = respMsg.Unpack(respBody)
	if err != nil {
		return "", 0, err
	}
	if len(respMsg.Answer) > 0 {
		re := regexp.MustCompile(`ech="([^"]+)"`)
		match := re.FindStringSubmatch(respMsg.Answer[0].String())
		if match[1] != "" {
			return match[1], respMsg.Answer[0].Header().Ttl, nil
		}
	}
	return "", 0, newError("no ech record found")
}