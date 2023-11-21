package containers

func TryAllParsers(rawConfig []byte, prioritizedParser string) (*Container, error) {
	if prioritizedParser != "" {
		if parser, found := knownParsers[prioritizedParser]; found {
			container, err := parser.ParseSubscriptionContainerDocument(rawConfig)
			if err == nil {
				return container, nil
			}
		}
	}

	for _, parser := range knownParsers {
		container, err := parser.ParseSubscriptionContainerDocument(rawConfig)
		if err == nil {
			return container, nil
		}
	}
	return nil, newError("no parser found for config")
}
