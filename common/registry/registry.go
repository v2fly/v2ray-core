package registry

import (
	"github.com/v2fly/v2ray-core/v4/common/protoext"
	"google.golang.org/protobuf/proto"
)

type implementationRegistry struct {
	implSet map[string]*implementationSet
}

func (i *implementationRegistry) RegisterImplementation(name string, opt *protoext.MessageOpt, loader CustomLoader) {
	interfaceType := opt.GetType()[0]
	implSet, found := i.implSet[interfaceType]
	if !found {
		implSet = newImplementationSet()
		i.implSet[interfaceType] = implSet
	}
	implSet.RegisterImplementation(name, opt, loader)
}

func (i *implementationRegistry) FindImplementationByAlias(interfaceType, alias string) (string, CustomLoader, error) {
	implSet, found := i.implSet[interfaceType]
	if !found {
		return "", nil, newError("cannot find implemention unknown interface type")
	}
	return implSet.FindImplementationByAlias(alias)
}

func newImplementationRegistry() *implementationRegistry {
	return &implementationRegistry{implSet: map[string]*implementationSet{}}
}

var globalImplementationRegistry = newImplementationRegistry()

// RegisterImplementation register an implementation of a type of interface
// loader(CustomLoader) is a private API, its interface is subject to breaking changes
func RegisterImplementation(proto proto.Message, loader CustomLoader) error {
	msgDesc := proto.ProtoReflect().Type().Descriptor()
	fullName := string(msgDesc.FullName())
	msgOpts, err := protoext.GetMessageOptions(msgDesc)
	if err != nil {
		return newError("unable to find message options").Base(err)
	}
	globalImplementationRegistry.RegisterImplementation(fullName, msgOpts, loader)
	return nil
}

func FindImplementationByAlias(interfaceType, alias string) (string, CustomLoader, error) {
	return globalImplementationRegistry.FindImplementationByAlias(interfaceType, alias)
}
