package all

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"v2ray.com/core/infra/link"
	"v2ray.com/core/main/commands/base"
)

var cmdLinks = &base.Command{
	UsageLine: "{{.Exec}} links [-t tag] [vmess://...] [vmess://...]",
	Short:     "Fetch and convert V2Ray links",
	Long: `
Fetch, convert V2Ray links and save to files.

{{.LongName}} now supports the following common seen link formats:

	* V2rayN (V2rayNG) link
	* Shadowrocket link
	* Quantumult (X) link

Arguments:

	-o
		Output directory to save config files. It is required when 
		convert multiple links.

	-t
		Tag prefix for the subscription, outbounds will be tagged with 
		"Prefix - Node Name", which is useful for selector filtering.

	-m
		Sets SO_MARK for outbounds, useful for Linux firewall.

	-u
	 	Fetch links from the the subscription URL.

	-c
		Fetch links from subscriptions defined by a config file.

	-i
		Ignore pattern (REGEXP), nodes match this pattern are ignored 
		and display their node names as information. The pattern is 
		prior, if a node matches it, it is ignored even if it matches 
		the select pattern.

	-s
		Select pattern (REGEXP), nodes whose tags match this pattern 
		are selected. Leave blank to select all nodes.

** NOTE **

	Patterns are only applies to subscription links.

Subscriptions Config File Example:
		
	{
		"subscriptions": [{
			"enabled": true,
			"tag": "all.name",
			"url": "https://url.to/subscriptions",
			"ignore": null,
			"match": null
		}]
	}

Examples:

	{{.Exec}} {{.LongName}} vmess://... vmess://...   (1)
	{{.Exec}} {{.LongName}} -t name -u url -o dir     (2)
	{{.Exec}} {{.LongName}} -c path/to/json -o dir    (3)

(1) Convert links and save to current directory.
(2) Fetch and convert links from the subscription url.
(3) Fetch and convert links from multiple subscriptions.
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
	linksMatch      = cmdLinks.Flag.String("s", "", "")
	linksOutdir     = cmdLinks.Flag.String("o", "", "")
	linksSocketMark = cmdLinks.Flag.Int("m", 0, "")
)

func executeLinks(cmd *base.Command, args []string) {
	links := cmd.Flag.Args()
	conf := &SubscriptionConfig{}
	if *linksConfPath != "" {
		data, err := ioutil.ReadFile(*linksConfPath)
		if err != nil {
			base.Fatalf("Failed to read file: %s", err)
		}
		err = json.Unmarshal(data, conf)
		if err != nil {
			base.Fatalf("Failed to load: %s", err)
		}
	}
	if *linksURL != "" {
		if conf.Subscriptions == nil {
			conf.Subscriptions = make([]*Subscription, 0)
		}
		conf.Subscriptions = append(conf.Subscriptions, &Subscription{
			Tag:    *linksTag,
			URL:    *linksURL,
			Match:  *linksMatch,
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
			base.Fatalf("Failed to fetch subscriptions: %s", err)
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
