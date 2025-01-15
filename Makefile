ifeq ($(GITHUB_SHA),)
	export GIT_VERSION := $(shell git describe --tags --always --dirty)
else
	export GIT_VERSION := $(strip $(GITHUB_SHA))	
endif

Ver=3.0.1_$(GIT_VERSION)
export Version=$(strip $(Ver))
export LDFlag="-s -w -X main.Version=$(Version)"


all:
	go build -ldflags $(LDFlag) -o dist/$(Version)/proxier ./main