package subscriptionmanager

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/sha3"

	"github.com/v2fly/v2ray-core/v5/app/subscription/containers"
	"github.com/v2fly/v2ray-core/v5/app/subscription/documentfetcher"
	"github.com/v2fly/v2ray-core/v5/app/subscription/specs"
)

func (s *SubscriptionManagerImpl) updateSubscription(subscriptionName string) error {
	var trackedSub *trackedSubscription
	if trackedSubFound, found := s.trackedSubscriptions[subscriptionName]; !found {
		return newError("not found")
	} else {
		trackedSub = trackedSubFound
	}
	importSource := trackedSub.importSource
	docFetcher, err := documentfetcher.GetFetcher("http")
	if err != nil {
		return newError("failed to get fetcher: ", err)
	}
	if strings.HasPrefix(importSource.Url, "data:") {
		docFetcher, err = documentfetcher.GetFetcher("dataurl")
		if err != nil {
			return newError("failed to get fetcher: ", err)
		}
	}

	downloadedDocument, err := docFetcher.DownloadDocument(s.ctx, importSource)
	if err != nil {
		return newError("failed to download document: ", err)
	}

	trackedSub.originalDocument = downloadedDocument

	container, err := containers.TryAllParsers(trackedSub.originalDocument, "")
	if err != nil {
		return newError("failed to parse document: ", err)
	}

	trackedSub.originalContainer = container

	parsedDocument := &specs.SubscriptionDocument{}
	parsedDocument.Metadata = container.Metadata

	trackedSub.originalServerConfig = make(map[string]*originalServerConfig)

	for _, server := range trackedSub.originalContainer.ServerSpecs {
		documentHash := sha3.Sum256(server.Content)
		serverConfigHashName := fmt.Sprintf("%x", documentHash)
		parsed, err := s.converter.TryAllConverters(server.Content, "outbound", server.KindHint)
		if err != nil {
			trackedSub.originalServerConfig["!!!"+serverConfigHashName] = &originalServerConfig{data: server.Content}
			continue
		}
		s.polyfillServerConfig(parsed, serverConfigHashName)
		parsedDocument.Server = append(parsedDocument.Server, parsed)
		trackedSub.originalServerConfig[parsed.Id] = &originalServerConfig{data: server.Content}
	}
	newError("new subscription document fetched and parsed from ", subscriptionName).AtInfo().WriteToLog()
	if err := s.applySubscriptionTo(subscriptionName, parsedDocument); err != nil {
		return newError("failed to apply subscription: ", err)
	}
	trackedSub.currentDocument = parsedDocument
	trackedSub.currentDocumentExpireTime = time.Now().Add(time.Second * time.Duration(importSource.DefaultExpireSeconds))
	return nil
}

func (s *SubscriptionManagerImpl) polyfillServerConfig(document *specs.SubscriptionServerConfig, hash string) {
	document.Id = hash

	if document.Metadata == nil {
		document.Metadata = make(map[string]string)
	}

	if id, ok := document.Metadata[ServerMetadataID]; !ok || id == "" {
		document.Metadata[ServerMetadataID] = document.Id
	} else {
		document.Id = document.Metadata[ServerMetadataID]
	}

	if fqn, ok := document.Metadata[ServerMetadataFullyQualifiedName]; !ok || fqn == "" {
		document.Metadata[ServerMetadataFullyQualifiedName] = hash
	}

	if tagName, ok := document.Metadata[ServerMetadataTagName]; !ok || tagName == "" {
		document.Metadata[ServerMetadataTagName] = document.Metadata[ServerMetadataID]
	}
	document.Metadata[ServerMetadataTagName] = s.restrictTagName(document.Metadata[ServerMetadataTagName])
}

func (s *SubscriptionManagerImpl) restrictTagName(tagName string) string {
	newTagName := &strings.Builder{}
	somethingRemoved := false
	for _, c := range tagName {
		if (unicode.IsLetter(c) || unicode.IsNumber(c)) && c < 128 {
			newTagName.WriteRune(c)
		} else {
			somethingRemoved = true
		}
	}
	newTagNameString := newTagName.String()
	if len(newTagNameString) > 24 {
		newTagNameString = newTagNameString[:15]
		somethingRemoved = true
	}
	if somethingRemoved {
		hashedTagName := sha3.Sum256([]byte(tagName))
		hashedTagNameString := fmt.Sprintf("%x", hashedTagName)
		newTagNameString = newTagNameString + "_" + hashedTagNameString[:8]
	}
	return newTagNameString
}
