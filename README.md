<div>
<img width="190" height="210" align="left"  src="https://raw.githubusercontent.com/v2fly/v2fly-github-io/master/docs/.vuepress/public/readme-logo.png" alt="V2Ray"/>
<br>
<h1>Project V</h1> 
<p>Project V is a set of network tools that helps you to build your own computer network.
It secures your network connections and thus protects your privacy.</p>
</div>

[![GitHub Test Badge](https://github.com/v2fly/v2ray-core/workflows/Test/badge.svg)](https://github.com/v2fly/v2ray-core/actions)
[![codecov.io](https://codecov.io/gh/v2fly/v2ray-core/branch/master/graph/badge.svg?branch=master)](https://codecov.io/gh/v2fly/v2ray-core?branch=master)
[![codebeat](https://goreportcard.com/badge/github.com/v2fly/v2ray-core)](https://goreportcard.com/report/github.com/v2fly/v2ray-core)
[![Downloads](https://img.shields.io/github/downloads/v2fly/v2ray-core/total.svg)]()

## Related Links
 - [Documentation](https://www.v2fly.org/) and [Newcomer's Instructions](https://www.v2fly.org/guide/start.html)
 - Welcome to translate V2Ray via: **[Transifex](https://www.transifex.com/v2fly/public/)**

## Installation

V2Ray binaries are directly available in {releases/latest} as well as some major distros' repositories, including Debian, Arch Linux, macOS (homebrew), etc. If you are willing to package V2Ray for other distros, you are also welcome to seek for help via our issues.


### FHS-install-script
_Maintainers: [@IceCodeNew](https://github.com/IceCodeNew)_

```
bash <(curl -L https://raw.githubusercontent.com/v2fly/fhs-install-v2ray/master/install-release.sh)
```

### Docker
_Maintainers wanted._

```
docker pull v2fly/v2fly-core
```

### Arch Linux
_Maintainers: [@felixonmars](https://github.com/felixonmars)_

```
pacman -S v2ray
```

### Debian
_Maintainers: [@rogers0](https://github.com/rogers0) [@ymshenyu](https://github.com/ymshenyu)_

```
coming soon
```

### macOS
_Maintainers: [@kidonng](https://github.com/kidonng)_

```
brew install v2ray
```

### Windows
_Maintainers: [@kidonng](https://github.com/kidonng)_

```
scoop install v2ray
or
choco install v2ray
```

## License

[The MIT License (MIT)](https://raw.githubusercontent.com/v2fly/v2ray-core/master/LICENSE)

## Credits

This repo relies on the following third-party projects:

- In production:
  - [gorilla/websocket](https://github.com/gorilla/websocket)
  - [lucas-clemente/quic-go](https://github.com/lucas-clemente/quic-go)
  - [pires/go-proxyproto](https://github.com/pires/go-proxyproto)
  - [seiflotfy/cuckoofilter](https://github.com/seiflotfy/cuckoofilter)
  - [google/starlark-go](https://github.com/google/starlark-go)
- For testing only:
  - [miekg/dns](https://github.com/miekg/dns)
  - [h12w/socks](https://github.com/h12w/socks)
