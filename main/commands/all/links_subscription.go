package all

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"

	"v2ray.com/core/infra/link"
)

// Subscription represents a subscription config
type Subscription struct {
	Tag    string `json:"tag"`
	URL    string `json:"url"`
	Ignore string `json:"ignore"`
	Select string `json:"select"`
}

func (s *Subscription) String() string {
	return fmt.Sprintf(`Tag: %s
URL: %s
Ignore: %s
Select: %s`,
		s.Tag, s.URL, s.Ignore, s.Select)
}

// SubscriptionConfig represents a subscription json
type SubscriptionConfig struct {
	Subscriptions []*Subscription `json:"subscriptions"`
}

// FetchSubscriptions fetches subscription specified by "conf", and generating json files to "outdir"
func FetchSubscriptions(subscriptions []*Subscription, outdir string, socketMark int32) error {
	filesMap, err := getFilesMap(outdir)
	if err != nil {
		return err
	}

	err = subscriptionsToJSONs(subscriptions, outdir, socketMark, filesMap)
	if err != nil {
		return err
	}
	for _, file := range filesMap {
		rel, err := filepath.Rel(outdir, file)
		if err != nil {
			return err
		}
		fmt.Println("Removed:", rel)
	}
	return nil
}

func subscriptionsToJSONs(subs []*Subscription, outdir string, socketMark int32, filesMap map[string]string) error {
	for _, sub := range subs {
		err := subscriptionToJSONs(sub, outdir, socketMark, filesMap)
		if err != nil {
			return err
		}
	}
	return nil
}

func subscriptionToJSONs(sub *Subscription, outdir string, socketMark int32, filesMap map[string]string) error {
	fmt.Println(sub)
	fmt.Println("Output:", outdir)
	if socketMark != 0 {
		fmt.Println("Sokect mark:", socketMark)
	}
	fmt.Println("Downloading...")
	links, err := LinksFromSubscription(sub.URL)
	if err != nil {
		return err
	}
	fmt.Printf("%v link(s) found...\n", len(links))

	links, err = filterLinks(links, sub.Ignore, sub.Select)
	if err != nil {
		return err
	}
	for _, link := range links {
		out := link.ToOutbound()
		out.Tag = asFileName(sub.Tag, link.Tag())
		filename := out.Tag + ".json"
		content, err := outbound2JSON(out, socketMark)
		if err != nil {
			return err
		}
		err = writeFile(outdir, filename, content, filesMap)
		if err != nil {
			return err
		}
	}
	return nil
}

// LinksFromSubscription downloads and parses links from a subscription URL
func LinksFromSubscription(url string) ([]link.Link, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	decoded, err := base64Decode(string(body))
	if err != nil {
		return nil, err
	}
	content := string(decoded)
	links := make([]link.Link, 0)
	for _, line := range strings.Split(content, "\n") {
		line = strings.Trim(line, " ")
		if line == "" {
			continue
		}
		link, err := link.Parse(line)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}
	return links, nil
}

func filterLinks(links []link.Link, exclude string, include string) ([]link.Link, error) {
	lks := make([]link.Link, 0)
	var (
		err        error
		regExclude *regexp.Regexp
		regInclude *regexp.Regexp
	)
	if exclude != "" {
		regExclude, err = regexp.Compile("(?i)" + exclude)
		if err != nil {
			return nil, err
		}
	}
	if include != "" {
		regInclude, err = regexp.Compile("(?i)" + include)
		if err != nil {
			return nil, err
		}
	}
	for _, l := range links {
		tag := l.Tag()
		if regExclude != nil && regExclude.Match([]byte(tag)) {
			fmt.Println(tag)
			continue
		}
		if regInclude != nil && !regInclude.Match([]byte(tag)) {
			continue
		}
		lks = append(lks, l)
	}
	return lks, nil
}
