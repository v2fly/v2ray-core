module github.com/v2fly/v2ray-core/v5

go 1.26.10

toolchain go1.27.3

require (
	github.com/adrg/xdg v0.5.3
	github.com/apernet/quic-go v0.59.1-0.20260217092621-db4786c77a22
	github.com/go-chi/chi/v5 v5.5.0
	github.com/go-chi/render v1.0.3
	github.com/go-playground/validator/v10 v10.30.5
	github.com/golang-collections/go-datastructures v0.0.0-20150211160725-59788d5eb259
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.4
	github.com/google/go-cmp v0.7.0
	github.com/google/gopacket v1.1.19
	github.com/gorilla/websocket v1.5.3
	github.com/improbable-eng/grpc-web v0.15.0
	github.com/jhump/protoreflect v1.19.0
	github.com/lunixbochs/struc v0.0.0-20241101090106-8d528fa2c543
	github.com/miekg/dns v1.1.73
	github.com/mustafaturan/bus v1.0.2
	github.com/pelletier/go-toml v1.9.5
	github.com/pion/dtls/v3 v3.1.15
	github.com/pion/ice/v4 v4.4.8
	github.com/pion/logging v0.2.4
	github.com/pion/stun/v3 v3.1.7
	github.com/pion/stun/v4 v4.0.1
	github.com/pion/transport/v2 v2.2.10
	github.com/pion/transport/v5 v5.1.1
	github.com/pion/webrtc/v4 v4.2.24
	github.com/pires/go-proxyproto v0.16.0
	github.com/quic-go/quic-go v0.64.1
	github.com/refraction-networking/utls v1.8.2
	github.com/seiflotfy/cuckoofilter v0.0.0-20240715131351-a2f2c23f1771
	github.com/stretchr/testify v1.12.1
	github.com/v2fly/BrowserBridge v0.0.0-20210430233438-0570fc1d7d08
	github.com/v2fly/VSign v0.0.0-20201108000810-e2adc24bf848
	github.com/v2fly/hysteria/core/v2 v2.0.0-20260220231229-39018a43855e
	github.com/v2fly/ss-bloomring v0.0.0-20210312155135-28617310f63e
	github.com/v2fly/struc v0.0.0-20241227015403-8e8fa1badfd6
	github.com/vincent-petithory/dataurl v1.0.0
	github.com/xiaokangwang/VLite v0.0.0-20231225174116-75fa4b06e9f2
	github.com/xssnick/raptorq v1.6.0
	github.com/xtaci/smux v1.5.57
	go.starlark.net v0.0.0-20261016124120-4811d7e34ea3
	go4.org/netipx v0.0.0-20260823151212-3075585bcbeb
	golang.org/x/crypto v0.58.0
	golang.org/x/net v0.61.0
	golang.org/x/sync v0.24.0
	golang.org/x/sys v0.49.0
	golang.zx2c4.com/wireguard v0.0.0-20261006164505-2631ce99a06f
	google.golang.org/grpc v1.84.0
	google.golang.org/protobuf v1.36.12
	gopkg.in/yaml.v3 v3.0.1
	gvisor.dev/gvisor v0.0.0-20260218152508-eed1cebf761e
	h12.io/socks v1.0.3
	lukechampine.com/blake3 v1.4.1
)

require (
	github.com/aead/cmac v0.0.0-20160719120800-7af84192f0b1 // indirect
	github.com/ajg/form v1.9.0 // indirect
	github.com/andybalholm/brotli v1.2.6 // indirect
	github.com/boljen/go-bitmap v0.0.0-20151001105940-23cd2fb0ce7d // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/desertbit/timer v1.0.1 // indirect
	github.com/dgryski/go-metro v0.0.0-20261008173524-5059e0b1da03 // indirect
	github.com/ebfe/bcrypt_pbkdf v0.0.0-20140212075826-3c8d2dcb253a // indirect
	github.com/gabriel-vasile/mimetype v1.4.16 // indirect
	github.com/go-playground/locales v0.14.2 // indirect
	github.com/go-playground/universal-translator v0.18.2 // indirect
	github.com/google/btree v1.1.3 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jhump/protoreflect/v2 v2.0.0 // indirect
	github.com/klauspost/compress v1.20.3 // indirect
	github.com/klauspost/cpuid/v2 v2.4.1 // indirect
	github.com/klauspost/reedsolomon v1.14.2 // indirect
	github.com/leodido/go-urn v1.5.0 // indirect
	github.com/mustafaturan/monoton v1.0.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pion/datachannel v1.6.3 // indirect
	github.com/pion/dtls/v2 v2.2.12 // indirect
	github.com/pion/interceptor v0.1.49 // indirect
	github.com/pion/mdns/v2 v2.2.2 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/rtcp v1.2.19 // indirect
	github.com/pion/rtp v1.10.5 // indirect
	github.com/pion/sctp v1.12.0 // indirect
	github.com/pion/sdp/v3 v3.0.20 // indirect
	github.com/pion/srtp/v3 v3.1.3 // indirect
	github.com/pion/transport/v3 v3.1.1 // indirect
	github.com/pion/transport/v4 v4.1.0 // indirect
	github.com/pion/turn/v5 v5.1.2 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/riobard/go-bloom v0.0.0-20200614022211-cdc8013cb5b3 // indirect
	github.com/rs/cors v1.11.1 // indirect
	github.com/secure-io/siv-go v0.0.0-20180922214919-5ff40651e2c4 // indirect
	github.com/stretchr/objx v0.8.0 // indirect
	github.com/wlynxg/anet v0.0.5 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/exp v0.0.0-20261014143641-f0251eb4aa3e // indirect
	golang.org/x/mobile v0.0.0-20261011034528-197770444867 // indirect
	golang.org/x/mod v0.42.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	golang.org/x/time v0.16.0 // indirect
	golang.org/x/tools v0.52.0 // indirect
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261019180522-ce908d29299b // indirect
	nhooyr.io/websocket v1.8.17 // indirect
)

tool golang.org/x/mobile/cmd/gobind
