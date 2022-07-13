package json_test

import (
	"encoding/json"
	"testing"

	. "github.com/v2fly/v2ray-core/v5/infra/conf/json"
)

func TestTOMLToJSON_V2Style(t *testing.T) {
	input := `
[log]
loglevel = 'debug'

[[inbounds]]
port = 10800
listen = '127.0.0.1'
protocol = 'socks'

[inbounds.settings]
udp = true

[[outbounds]]
protocol = 'vmess'
[[outbounds.settings.vnext]]
port = 443
address = 'example.com'

[[outbounds.settings.vnext.users]]
id = '98a15fa6-2eb1-edd5-50ea-cfc428aaab78'

[outbounds.streamSettings]
network = 'tcp'
security = 'tls'
`
	expected := `
{
    "log": {
        "loglevel": "debug"
    },
    "inbounds": [{
        "port": 10800,
        "listen": "127.0.0.1",
        "protocol": "socks",
        "settings": {
            "udp": true
        }
    }],
    "outbounds": [{
        "protocol": "vmess",
        "settings": {
            "vnext": [{
                "port": 443,
                "address": "example.com",
                "users": [{
                    "id": "98a15fa6-2eb1-edd5-50ea-cfc428aaab78"
                }]
            }]
        },
        "streamSettings": {
            "network": "tcp",
            "security": "tls"
        }
    }]
}
`
	bs, err := FromTOML([]byte(input))
	if err != nil {
		t.Error(err)
	}
	m := make(map[string]interface{})
	json.Unmarshal(bs, &m)
	assertResult(t, m, expected)
}

func TestTOMLToJSON_ValueTypes(t *testing.T) {
	input := `
boolean = [ true, false, true, false ]
float = [ 3.14, 685_230.15 ]
int = [ 123, 685_230 ]
string = [ "哈哈", "Hello world", "newline newline2" ]
date = [ "2018-02-17" ]
datetime = [ "2018-02-17T15:02:31+08:00" ]
1 = 0
true = true
str = "hello"

[null]
nodeName = "node"
`
	expected := `
{
    "boolean": [true, false, true, false],
    "float": [3.14, 685230.15],
    "int": [123, 685230],
    "null": {
        "nodeName": "node"
    },
    "string": ["哈哈", "Hello world",  "newline newline2"],
    "date": ["2018-02-17"],
    "datetime": ["2018-02-17T15:02:31+08:00"],
    "1": 0,
    "true": true,
    "str": "hello"
}
`
	bs, err := FromTOML([]byte(input))
	if err != nil {
		t.Error(err)
	}
	m := make(map[string]interface{})
	json.Unmarshal(bs, &m)
	assertResult(t, m, expected)
}
