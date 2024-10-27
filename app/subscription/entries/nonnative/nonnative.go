package nonnative

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"regexp"
	"strings"
)

//go:generate go run github.com/v2fly/v2ray-core/v5/common/errors/errorgen

func ExtractAllValuesFromBytes(bytes []byte) AbstractNonNativeLink {
	link := AbstractNonNativeLink{}
	link.fromBytes(bytes)
	return link
}

type jsonDocument map[string]json.RawMessage

type AbstractNonNativeLink struct {
	Values map[string]string
}

func (a *AbstractNonNativeLink) fromBytes(bytes []byte) {
	a.Values = make(map[string]string)
	content := string(bytes)
	content = strings.Trim(content, " \n\t\r")
	a.extractValue(content, "root")
}

func (a *AbstractNonNativeLink) extractValue(content, prefix string) {
	if content == "" {
		return
	}

	{
		// check if the content is a link
		match, err := regexp.Match("[a-zA-Z0-9]+:((\\/\\/)|\\?)", []byte(content))
		if err != nil {
			panic(err)
		}
		if match {
			// if so, parse as link
			parsedURL, err := url.Parse(content)
			// if process is successful, then continue to parse every element of the link
			if err == nil {
				a.Values[prefix+"_!kind"] = "link"
				a.extractLink(parsedURL, prefix)
				return
			}
		}
	}
	{
		// check if it is base64
		content = strings.Trim(content, "=")
		decoded, err := base64.RawStdEncoding.DecodeString(content)
		if err == nil {
			a.Values[prefix+"_!kind"] = "base64"
			a.Values[prefix+"_!rawContent"] = string(decoded)
			a.extractValue(string(decoded), prefix+"_!base64")
			return
		}
	}
	{
		// check if it is base64url
		content = strings.Trim(content, "=")
		decoded, err := base64.RawURLEncoding.DecodeString(content)
		if err == nil {
			a.Values[prefix+"_!kind"] = "base64url"
			a.Values[prefix+"_!rawContent"] = string(decoded)
			a.extractValue(string(decoded), prefix+"_!base64")
			return
		}
	}
	{
		// check if it is json
		var doc jsonDocument
		if err := json.Unmarshal([]byte(content), &doc); err == nil {
			a.Values[prefix+"_!kind"] = "json"
			a.extractJSON(&doc, prefix)
			return
		}
	}
}

func (a *AbstractNonNativeLink) extractLink(content *url.URL, prefix string) {
	a.Values[prefix+"_!link"] = content.String()
	a.Values[prefix+"_!link_protocol"] = content.Scheme
	a.Values[prefix+"_!link_host"] = content.Host
	a.extractValue(content.Host, prefix+"_!link_host")
	a.Values[prefix+"_!link_path"] = content.Path
	a.Values[prefix+"_!link_query"] = content.RawQuery
	a.Values[prefix+"_!link_fragment"] = content.Fragment
	a.Values[prefix+"_!link_userinfo"] = content.User.String()
	a.extractValue(content.User.String(), prefix+"_!link_userinfo_!value")
	a.Values[prefix+"_!link_opaque"] = content.Opaque
}

func (a *AbstractNonNativeLink) extractJSON(content *jsonDocument, prefix string) {
	for key, value := range *content {
		switch value[0] {
		case '{':
			a.extractValue(string(value), prefix+"_!json_"+key)
		case '"':
			var unquoted string
			if err := json.Unmarshal(value, &unquoted); err == nil {
				a.Values[prefix+"_!json_"+key+"_!unquoted"] = unquoted
			}
			fallthrough
		default:
			a.Values[prefix+"_!json_"+key] = string(value)
		}
	}
}
