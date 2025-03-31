ifeq ($(GITHUB_SHA),)
	export GIT_VERSION := $(shell git describe --tags --always --dirty)
else
	export GIT_VERSION := $(strip $(GITHUB_SHA))	
endif

Ver=3.0.1_$(GIT_VERSION)
export Version=$(strip $(Ver))
export LDFlag="-s -w -X main.Version=$(Version)"


gateway:
	go build -ldflags $(LDFlag) -o dist/$(Version)/elink ./main
	cp -rf main/samples dist/$(Version)/

release:gateway router dat

sjw74:
	CGO_ENABLED=1 \
       CC=/opt/openeuler/oecore-x86_64/sysroots/x86_64-openeulersdk-linux/usr/bin/x86_64-openeuler-linux-gcc\
       GOOS=linux \
       GOARCH=amd64 \
        go build -ldflags $(LDFlag) -o dist/$(Version)/elink_embed ./main

router: 
	CGO_ENABLED=1 \
       CC=/opt/openwrt-sdk/staging_dir/toolchain-x86_64_gcc-11.3.0_musl/bin/x86_64-openwrt-linux-gcc\
       GOOS=linux \
       GOARCH=amd64 \
        go build -ldflags $(LDFlag) -o dist/$(Version)/elink_embed ./main

dat:
	echo ">>> Download latest geoip..."
	curl -s -L -o dist/$(Version)/geoip.dat "https://github.com/v2fly/geoip/raw/release/geoip.dat"
	echo ">>> Download latest geosite..."
	curl -s -L -o dist/$(Version)/geosite.dat "https://github.com/v2fly/domain-list-community/raw/release/dlc.dat"

clean:
	rm -rf dist/*

