package registry

import "github.com/v2fly/v2ray-core/v4/common/protoext"

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

type implementationSet struct {
	AliasLookup map[string]*implementation
}

type implementation struct {
	FullName string
	Alias    []string
}

func (i *implementationSet) RegisterImplementation(name string, opt *protoext.MessageOpt) {
	alias := opt.GetShortName()

	impl := &implementation{
		FullName: name,
		Alias:    alias,
	}

	for _, aliasName := range alias {
		i.AliasLookup[aliasName] = impl
	}
}

func (i *implementationSet) FindImplementationByAlias(alias string) (string, error) {
	impl, found := i.AliasLookup[alias]
	if found {
		return impl.FullName, nil
	}
	return "", newError("cannot find implementation by alias")
}

func newImplementationSet() *implementationSet {
	return &implementationSet{AliasLookup: map[string]*implementation{}}
}
