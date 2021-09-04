package registry

import (
	"github.com/v2fly/v2ray-core/v4/common/protoext"
	"google.golang.org/protobuf/proto"
)

type implementationRegistry struct {
	implSet map[string]*implementationSet
}

func (i *implementationRegistry) RegisterImplementation(name string, opt *protoext.MessageOpt) {
	interfaceType := opt.GetType()[0]
	implSet, found := i.implSet[interfaceType]
	if !found {
		implSet = newImplementationSet()
		i.implSet[interfaceType] = implSet
	}
	implSet.RegisterImplementation(name, opt)
}

func (i *implementationRegistry) FindImplementationByAlias(interfaceType, alias string) (string, error) {
	implSet, found := i.implSet[interfaceType]
	if !found {
		return "", newError("cannot find implemention unknown interface type")
	}
	return implSet.FindImplementationByAlias(alias)
}

func newImplementationRegistry() *implementationRegistry {
	return &implementationRegistry{implSet: map[string]*implementationSet{}}
}

var globalImplementationRegistry = newImplementationRegistry()

func RegisterImplementation(proto proto.Message) error {
	msgDesc := proto.ProtoReflect().Type().Descriptor()
	fullName := string(msgDesc.FullName())
	msgOpts, err := protoext.GetMessageOptions(msgDesc)
	if err != nil {
		return newError("unable to find message options").Base(err)
	}
	globalImplementationRegistry.RegisterImplementation(fullName, msgOpts)
	return nil
}

func FindImplementationByAlias(interfaceType, alias string) (string, error) {
	return globalImplementationRegistry.FindImplementationByAlias(interfaceType, alias)
}
