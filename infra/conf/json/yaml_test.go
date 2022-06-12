package json_test

import (
	"encoding/json"
	"testing"

	. "github.com/v2fly/v2ray-core/v5/infra/conf/json"
)

func TestYMLToJSON_V2Style(t *testing.T) {
	input := `
log:
  loglevel: debug
inbounds:
- port: 10800
  listen: 127.0.0.1
  protocol: socks
  settings:
    udp: true
outbounds:
- protocol: vmess
  settings:
    vnext:
    - address: example.com
      port: 443
      users:
      - id: '98a15fa6-2eb1-edd5-50ea-cfc428aaab78'
  streamSettings:
    network: tcp
    security: tls
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
	bs, err := FromYAML([]byte(input))
	if err != nil {
		t.Error(err)
	}
	m := make(map[string]interface{})
	json.Unmarshal(bs, &m)
	assertResult(t, m, expected)
}

func TestYMLToJSON_ValueTypes(t *testing.T) {
	input := `
boolean: 
    - TRUE
    - FALSE
    - true
    - false
float:
    - 3.14
    - 6.8523015e+5
int:
    - 123
    - 0b1010_0111_0100_1010_1110
null:
    nodeName: 'node'
    parent: ~  # ~ for null
string:
    - 哈哈
    - 'Hello world'
    - newline
      newline2    # multi-line string
date:
    - 2018-02-17    # yyyy-MM-dd
datetime: 
    -  2018-02-17T15:02:31+08:00    # ISO 8601 time
mixed:
    - true
    - false
    - 1
    - 0
    - null
    - hello
# arbitrary keys
1: 0
true: false
TRUE: TRUE
"str": "hello"
`
	expected := `
{
    "boolean": [true, false, true, false],
    "float": [3.14, 685230.15],
    "int": [123, 685230],
    "null": {
        "nodeName": "node",
        "parent": null
    },
    "string": ["哈哈", "Hello world",  "newline newline2"],
    "date": ["2018-02-17T00:00:00Z"],
    "datetime": ["2018-02-17T15:02:31+08:00"],
    "mixed": [true,false,1,0,null,"hello"],
    "1": 0,
    "true": true,
    "str": "hello"
}
`
	bs, err := FromYAML([]byte(input))
	if err != nil {
		t.Error(err)
	}
	m := make(map[string]interface{})
	json.Unmarshal(bs, &m)
	assertResult(t, m, expected)
}
