module github.com/v2fly/v2ray-core/v5

go 1.24.0

toolchain go1.25.6

require (
	github.com/adrg/xdg v0.5.3
	github.com/apernet/quic-go v0.48.2-0.20241104191913-cb103fcecfe7
	github.com/go-chi/chi/v5 v5.2.4
	github.com/go-chi/render v1.0.3
	github.com/go-playground/validator/v10 v10.30.1
	github.com/golang-collections/go-datastructures v0.0.0-20150211160725-59788d5eb259
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.4
	github.com/google/go-cmp v0.7.0
	github.com/google/gopacket v1.1.19
	github.com/gorilla/websocket v1.5.3
	github.com/improbable-eng/grpc-web v0.15.0
	github.com/jhump/protoreflect v1.17.0
	github.com/miekg/dns v1.1.70
	github.com/mustafaturan/bus v1.0.2
	github.com/pelletier/go-toml v1.9.5
	github.com/pion/dtls/v2 v2.2.12
	github.com/pion/transport/v2 v2.2.10
	github.com/pires/go-proxyproto v0.8.1
	github.com/quic-go/quic-go v0.55.0
	github.com/refraction-networking/utls v1.8.2
	github.com/seiflotfy/cuckoofilter v0.0.0-20220411075957-e3b120b3f5fb
	github.com/stretchr/testify v1.11.1
	github.com/v2fly/BrowserBridge v0.0.0-20210430233438-0570fc1d7d08
	github.com/v2fly/VSign v0.0.0-20201108000810-e2adc24bf848
	github.com/v2fly/hysteria/core/v2 v2.0.0-20250113081444-b0a0747ac7ab
	github.com/v2fly/ss-bloomring v0.0.0-20210312155135-28617310f63e
	github.com/v2fly/struc v0.0.0-20241227015403-8e8fa1badfd6
	github.com/vincent-petithory/dataurl v1.0.0
	github.com/xiaokangwang/VLite v0.0.0-20220418190619-cff95160a432
	go.starlark.net v0.0.0-20230612165344-9532f5667272
	go4.org/netipx v0.0.0-20230303233057-f1b76eb4bb35
	golang.org/x/crypto v0.47.0
	golang.org/x/net v0.49.0
	golang.org/x/sync v0.19.0
	golang.org/x/sys v0.40.0
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.11
	gopkg.in/yaml.v3 v3.0.1
	gvisor.dev/gvisor v0.0.0-20231020174304-b8a429915ff1
	h12.io/socks v1.0.3
	lukechampine.com/blake3 v1.4.1
)

require (
	github.com/aead/cmac v0.0.0-20160719120800-7af84192f0b1 // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/andybalholm/brotli v1.0.6 // indirect
	github.com/boljen/go-bitmap v0.0.0-20151001105940-23cd2fb0ce7d // indirect
	github.com/bufbuild/protocompile v0.14.1 // indirect
	github.com/cenkalti/backoff/v4 v4.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/dgryski/go-metro v0.0.0-20211217172704-adc40b04c140 // indirect
	github.com/ebfe/bcrypt_pbkdf v0.0.0-20140212075826-3c8d2dcb253a // indirect
	github.com/gabriel-vasile/mimetype v1.4.12 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/google/btree v1.1.2 // indirect
	github.com/google/pprof v0.0.0-20240320155624-b11c3daa6f07 // indirect
	github.com/klauspost/compress v1.17.4 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/klauspost/reedsolomon v1.11.7 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lunixbochs/struc v0.0.0-20200707160740-784aaebc1d40 // indirect
	github.com/mustafaturan/monoton v1.0.0 // indirect
	github.com/onsi/ginkgo/v2 v2.17.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible // indirect
	github.com/pion/logging v0.2.2 // indirect
	github.com/pion/randutil v0.1.0 // indirect
	github.com/pion/sctp v1.8.7 // indirect
	github.com/pion/transport/v3 v3.0.7 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/quic-go/qpack v0.5.1 // indirect
	github.com/riobard/go-bloom v0.0.0-20200614022211-cdc8013cb5b3 // indirect
	github.com/rs/cors v1.7.0 // indirect
	github.com/secure-io/siv-go v0.0.0-20180922214919-5ff40651e2c4 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/xtaci/smux v1.5.24 // indirect
	go.uber.org/mock v0.5.2 // indirect
	golang.org/x/exp v0.0.0-20240506185415-9bf2ced13842 // indirect
	golang.org/x/mod v0.31.0 // indirect
	golang.org/x/text v0.33.0 // indirect
	golang.org/x/time v0.5.0 // indirect
	golang.org/x/tools v0.40.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251029180050-ab9386a59fda // indirect
	nhooyr.io/websocket v1.8.6 // indirect
)
