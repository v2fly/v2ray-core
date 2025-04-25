package geodata

import (
	"strings"

	"github.com/ghxhy/v2ray-core/v5/app/router/routercommon"
)

type AttributeList struct {
	matcher []AttributeMatcher
}

func (al *AttributeList) Match(domain *routercommon.Domain) bool {
	for _, matcher := range al.matcher {
		if !matcher.Match(domain) {
			return false
		}
	}
	return true
}

func (al *AttributeList) IsEmpty() bool {
	return len(al.matcher) == 0
}

func parseAttrs(attrs []string) *AttributeList {
	al := new(AttributeList)
	for _, attr := range attrs {
		trimmedAttr := strings.ToLower(strings.TrimSpace(attr))
		if len(trimmedAttr) == 0 {
			continue
		}
		al.matcher = append(al.matcher, BooleanMatcher(trimmedAttr))
	}
	return al
}

type AttributeMatcher interface {
	Match(*routercommon.Domain) bool
}

type BooleanMatcher string

func (m BooleanMatcher) Match(domain *routercommon.Domain) bool {
	for _, attr := range domain.Attribute {
		if strings.EqualFold(attr.GetKey(), string(m)) {
			return true
		}
	}
	return false
}
