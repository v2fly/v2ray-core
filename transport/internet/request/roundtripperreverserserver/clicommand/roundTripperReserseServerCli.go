package clicommand

import (
	"bytes"
	"context"
	"flag"
	"os"

	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/common/environment/systemnetworkimpl"
	"github.com/v2fly/v2ray-core/v5/main/commands/all/engineering"
	"github.com/v2fly/v2ray-core/v5/main/commands/base"
	"github.com/v2fly/v2ray-core/v5/transport/internet/request/roundtripperreverserserver"

	"google.golang.org/protobuf/encoding/protojson"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

var config *string

var cmdRTTReverseServer = &base.Command{
	UsageLine: "{{.Exec}} engineering request-rtt-reverse-server",
	Flag: func() flag.FlagSet {
		fs := flag.NewFlagSet("", flag.ExitOnError)
		config = fs.String("c", "", "")
		return *fs
	}(),
	Run: func(cmd *base.Command, args []string) {
		err := cmd.Flag.Parse(args)
		if err != nil {
			base.Fatalf("failed to parse flags: %v", err)
		}
		fd, err := os.Open(*config)
		if err != nil {
			base.Fatalf("failed to open config file %q: %v", *config, err)
		}
		defer func() { _ = fd.Close() }()
		content := bytes.NewBuffer(nil)
		_, err = content.ReadFrom(fd)
		if err != nil {
			base.Fatalf("failed to read config file %q: %v", *config, err)
		}
		var reverserConfig roundtripperreverserserver.Config
		err = protojson.Unmarshal(content.Bytes(), &reverserConfig)
		if err != nil {
			base.Fatalf("failed to unmarshal JSON config from %q: %v", *config, err)
		}

		ctx := context.Background()
		systemNetworkImpl := systemnetworkimpl.NewSystemNetworkDefault()
		ctx = envctx.ContextWithEnvironment(ctx, systemNetworkImpl)

		server, err := roundtripperreverserserver.NewReverser(ctx, &reverserConfig)
		if err != nil {
			base.Fatalf("failed to create RTT reverse server: %v", err)
		}
		_ = server
		select {}
	},
}

func init() {
	engineering.AddCommand(cmdRTTReverseServer)
}
