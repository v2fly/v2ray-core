ifeq ($(GITHUB_SHA),)
	export GIT_VERSION := $(shell git describe --tags --always --dirty)
else
	export GIT_VERSION := $(strip $(GITHUB_SHA))	
endif

Ver=3.0.1_$(GIT_VERSION)
export Version=$(strip $(Ver))
export LDFlag="-s -w -X main.Version=$(Version)"


all:
	go build -ldflags $(LDFlag) -o dist/$(Version)/elink ./main
	cp -rf main/samples dist/$(Version)/

dat:
	echo ">>> Download latest geoip..."
	curl -s -L -o dist/$(Version)/geoip.dat "https://github.com/v2fly/geoip/raw/release/geoip.dat"
	echo ">>> Download latest geosite..."
	curl -s -L -o dist/$(Version)/geosite.dat "https://github.com/v2fly/domain-list-community/raw/release/dlc.dat"

