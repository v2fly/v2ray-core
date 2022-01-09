NAME=v2ray
BINDIR=build
VERSION=5.0.2
GOBUILD=CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -buildid=" ./main

PLATFORM = \
	android-arm64-v8a \
	dragonfly-64 \
	freebsd-32 \
	freebsd-64 \
	linux-32 \
	linux-64 \
	linux-arm32-v5 \
	linux-arm32-v6 \
	linux-arm32-v7a \
	linux-arm32-v8a \
	linux-mips32 \
	linux-mips32le \
	linux-mips64 \
	linux-mips64le \
	linux-riscv64 \
	macos-64 \
	macos-arm64-v8a \
	openbsd-32 \
	openbsd-64
	
WINDOWS_ARCH = \
	windows-32 \
	windows-64 \
	windows-arm32-v7a \
	windows-arm64-v8a

android-arm64-v8a:
	GOARCH=arm64 GOOS=android $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

dragonfly-64:
	GOARCH=amd64 GOOS=dragonfly $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

freebsd-32:
	GOARCH=386 GOOS=freebsd $(GOBUILD) -o $(BINDIR)/$(NAME)-$@
	
freebsd-64:
	GOARCH=amd64 GOOS=freebsd $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-32:
	GOARCH=386 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-64:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-arm32-v5:
	GOARCH=arm GOOS=linux GOARM=5 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-arm32-v6:
	GOARCH=arm GOOS=linux GOARM=6 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-arm32-v7a:
	GOARCH=arm64 GOOS=linux GOARM=7 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-arm32-v8a:
	GOARCH=arm64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-mips32:
	GOARCH=mips32 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-mips32le:
	GOARCH=mips32le GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-mips64:
	GOARCH=mips64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-mips64le:
	GOARCH=mips64le GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

linux-riscv64:
	GOARCH=riscv64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@
	
macos-64:
	GOARCH=amd64 GOOS=macos $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

macos-arm64-v8a:
	GOARCH=arm64 GOOS=macos $(GOBUILD) -o $(BINDIR)/$(NAME)-$@

openbsd-32:
	GOARCH=386 GOOS=openbsd $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe

openbsd-64:
	GOARCH=amd64 GOOS=openbsd $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe

windows-32:
	GOARCH=386 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe

windows-64:
	GOARCH=amd64 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe

windows-arm32-v7a:
	GOARCH=arm GOOS=windows GOARM=7 $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe

windows-arm64-v8a:
	GOARCH=arm64 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe

gz_releases=$(addsuffix .gz, $(PLATFORM))
zip_releases=$(addsuffix .zip, $(WINDOWS_ARCH))

$(gz_releases): %.gz : %
	chmod +x $(BINDIR)/$(NAME)-$(basename $@)
	gzip -S -f -$(VERSION).gz $(BINDIR)/$(NAME)-$(basename $@)

$(zip_releases): %.zip : %
	zip -m -j $(BINDIR)/$(NAME)-$(basename $@)-$(VERSION).zip $(BINDIR)/$(NAME)-$(basename $@).exe

all-arch: $(PLATFORM) $(WINDOWS_ARCH)

releases: $(gz_releases) $(zip_releases)