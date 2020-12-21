package all

import (
	"os"

	"v2ray.com/core/common/cmdarg"
	"v2ray.com/core/infra/conf/serial"
	"v2ray.com/core/infra/link"
	"v2ray.com/core/main/commands/base"
)

var cmdLinks = &base.Command{
	UsageLine: "{{.Exec}} links [-o dir] [\"vmess://...\"] ...",
	Short:     "Fetch and convert V2Ray links",
	Long: `
Fetch, convert V2Ray links to json.

Arguments:

	-o
		Save config files to output directory. Required when convert 
		multiple links.

	-t
		Tag prefix for the subscription, outbounds will be tagged with 
		"Prefix - Node Name", which is useful for selector filtering.

	-m
		Sets SO_MARK for outbounds, useful for Linux firewall.

	-u
		Fetch links from the the subscription URL.

	-i
		Ignore pattern (REGEXP) for '-u' parameter, nodes match this 
		pattern are ignored and display their names as information. 
		The pattern is prior, if a node matches, it is ignored even 
		it matches the select pattern.

	-s
		Select pattern (REGEXP) for '-u' parameter, nodes whose tags 
		match it are selected. Leave blank to select all nodes.

	-c
		Fetch links from subscriptions spcified by a config file.
		Config example:
		
			{
				"subscriptions": [{
					"enabled": true,
					"tag": "all.name",
					"url": "https://url.to/subscriptions",
					"ignore": null,
					"select": null
				}]
			}

		"ignore" and "select" pattern works just like "-i", "-s".

Examples:

	{{.Exec}} {{.LongName}} "vmess://..."                    (1)
	{{.Exec}} {{.LongName}} -o . "vmess://..." "vmess://..." (2)
	{{.Exec}} {{.LongName}} -o . -t name -u url              (3)
	{{.Exec}} {{.LongName}} -o . -c path/to/json             (4)

(1) Convert a link, and print to stdout.
(2) Convert links, and save to current directory.
(3) Fetch, convert and save links from the subscription url.
(4) Fetch, convert and save links from multiple subscriptions.
`,
}

func init() {
	cmdLinks.Run = executeLinks
}

var (
	linksConfPath   = cmdLinks.Flag.String("c", "", "")
	linksTag        = cmdLinks.Flag.String("t", "", "")
	linksURL        = cmdLinks.Flag.String("u", "", "")
	linksIgnore     = cmdLinks.Flag.String("i", "", "")
	linksSelect     = cmdLinks.Flag.String("s", "", "")
	linksOutdir     = cmdLinks.Flag.String("o", "", "")
	linksSocketMark = cmdLinks.Flag.Int("m", 0, "")
)

func executeLinks(cmd *base.Command, args []string) {
	links := cmd.Flag.Args()
	conf := &SubscriptionConfig{}
	if *linksConfPath != "" {
		r, err := cmdarg.LoadArg(*linksConfPath)
		if err != nil {
			base.Fatalf("Failed to load %s: %s", *linksConfPath, err)
		}
		err = serial.DecodeJSON(r, conf)
		if err != nil {
			base.Fatalf("Failed to load %s: %s", *linksConfPath, err)
		}
	}
	if *linksURL != "" {
		if conf.Subscriptions == nil {
			conf.Subscriptions = make([]*Subscription, 0)
		}
		conf.Subscriptions = append(conf.Subscriptions, &Subscription{
			Tag:    *linksTag,
			URL:    *linksURL,
			Select: *linksSelect,
			Ignore: *linksIgnore,
		})
	}

	if len(links) == 0 && len(conf.Subscriptions) == 0 {
		base.Fatalf("No links or subscription specified")
	}

	// if single link and no outdir, outputs to stdout
	// useful for "v2ray links vmess://... | v2ray convert -output=yaml"
	if len(links) == 1 && *linksConfPath == "" && *linksURL == "" && *linksOutdir == "" {
		singleLinksToStdout(links[0], *linksTag, int32(*linksSocketMark))
		return
	}

	if *linksOutdir == "" {
		base.Fatalf("output directory not specified")
	}

	if len(conf.Subscriptions) > 0 {
		err := FetchSubscriptions(conf.Subscriptions, *linksOutdir, int32(*linksSocketMark))
		if err != nil {
			base.Fatalf("%s", err)
		}
	}
	if len(links) > 0 {
		linksToFiles(links, *linksTag, *linksOutdir, int32(*linksSocketMark))
	}
}

func singleLinksToStdout(l string, prefix string, socketMark int32) {
	link, err := link.Parse(l)
	if err != nil {
		base.Fatalf("failed to parse link:%s\n%s", l, err)
	}
	tag := asFileName(prefix, link.Tag())
	out, err := linkToJSON(link, tag, socketMark)
	if err != nil {
		base.Fatalf("failed to convert to json:%s", err)
	}
	if _, err := os.Stdout.Write(out); err != nil {
		base.Fatalf("failed to write stdout: %s", err)
	}
}

func linksToFiles(links []string, prefix string, outdir string, socketMark int32) {
	filesMap, err := getFilesMap(outdir)
	if err != nil {
		base.Fatalf("failed to read %s:%s", outdir, err)
	}
	for _, l := range links {
		link, err := link.Parse(l)
		if err != nil {
			base.Fatalf("failed to parse link:%s\n%s", l, err)
		}
		tag := asFileName(prefix, link.Tag())
		filename := tag + ".json"
		content, err := linkToJSON(link, tag, socketMark)
		if err != nil {
			base.Fatalf("failed to convert to json:%s", err)
		}
		err = writeFile(outdir, filename, content, filesMap)
		if err != nil {
			base.Fatalf("failed to save %s:%s", filename, err)
		}
	}
}
