package registry

import (
	"bytes"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/v2fly/v2ray-core/v4/common/protoext"
	"github.com/v2fly/v2ray-core/v4/common/serial"
	protov2 "google.golang.org/protobuf/proto"
	"reflect"
	"strings"
	"sync"
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

func (i *implementationRegistry) findImplementationByAlias(interfaceType, alias string) (string, CustomLoader, error) {
	implSet, found := i.implSet[interfaceType]
	if !found {
		return "", nil, newError("cannot find implemention unknown interface type")
	}
	return implSet.findImplementationByAlias(alias)
}

func (i *implementationRegistry) LoadImplementationByAlias(interfaceType, alias string, data []byte) (proto.Message, error) {
	var implementationFullName string

	if strings.HasPrefix(alias, "#") {
		// skip resolution for full name
		implementationFullName = alias
	} else {
		registryResult, customLoader, err := i.findImplementationByAlias(interfaceType, alias)
		if err != nil {
			return nil, newError("unable to find implementation").Base(err)
		}
		if customLoader != nil {
			return customLoader(data, i)
		}
		implementationFullName = registryResult
	}
	implementationConfigInstance, err := serial.GetInstance(implementationFullName)
	if err != nil {
		return nil, newError("unable to create implementation config instance").Base(err)
	}

	unmarshaler := jsonpb.Unmarshaler{AllowUnknownFields: false}
	err = unmarshaler.Unmarshal(bytes.NewReader(data), implementationConfigInstance.(proto.Message))
	if err != nil {
		return nil, newError("unable to parse json content").Base(err)
	}

	return implementationConfigInstance.(proto.Message), nil

}

func newImplementationRegistry() *implementationRegistry {
	return &implementationRegistry{implSet: map[string]*implementationSet{}}
}

var globalImplementationRegistry = newImplementationRegistry()

var initialized = &sync.Once{}

type registerRequest struct {
	proto  interface{}
	loader CustomLoader
}

var registerRequests []registerRequest

// RegisterImplementation register an implementation of a type of interface
// loader(CustomLoader) is a private API, its interface is subject to breaking changes
func RegisterImplementation(proto interface{}, loader CustomLoader) error {
	registerRequests = append(registerRequests, registerRequest{
		proto:  proto,
		loader: loader,
	})
	return nil
}

func registerImplementation(proto interface{}, loader CustomLoader) error {
	protoReflect := reflect.New(reflect.TypeOf(proto).Elem())
	var proto2 protov2.Message
	assignMessage := func(msg protov2.Message) {
		proto2 = msg
	}
	reflect.ValueOf(assignMessage).Call([]reflect.Value{protoReflect})
	msgDesc := proto2.ProtoReflect().Descriptor()
	fullName := string(msgDesc.FullName())
	msgOpts, err := protoext.GetMessageOptions(msgDesc)
	if err != nil {
		return newError("unable to find message options").Base(err)
	}
	globalImplementationRegistry.RegisterImplementation(fullName, msgOpts, loader)
	return nil
}

type LoadByAlias interface {
	LoadImplementationByAlias(interfaceType, alias string, data []byte) (proto.Message, error)
}

func LoadImplementationByAlias(interfaceType, alias string, data []byte) (proto.Message, error) {
	initialized.Do(func() {
		for _, v := range registerRequests {
			registerImplementation(v.proto, v.loader)
		}
	})
	return globalImplementationRegistry.LoadImplementationByAlias(interfaceType, alias, data)
}
