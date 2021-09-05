package protofilter

import (
	"github.com/v2fly/v2ray-core/v4/common/platform/filesystem"
	"github.com/v2fly/v2ray-core/v4/common/protoext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

func FilterProtoConfig(config proto.Message) error {
	messageProtoReflect := config.ProtoReflect()
	return filterMessage(messageProtoReflect)
}

func filterMessage(message protoreflect.Message) error {
	var err error
	type fileRead struct {
		filename string
		field    string
	}
	var fileReadingQueue []fileRead
	message.Range(func(descriptor protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		v2extension, ferr := protoext.GetFieldOptions(descriptor)
		if ferr != nil {
			if v2extension.Forbidden {
				if value.Bool() {
					err = newError("a forbidden value is set ", descriptor.FullName())
					return false
				}
			}

			if v2extension.ConvertTimeReadFileInto != "" {
				fileReadingQueue = append(fileReadingQueue, fileRead{
					filename: value.String(),
					field:    v2extension.ConvertTimeReadFileInto,
				})
			}
		}

		switch descriptor.Kind() {
		case protoreflect.MessageKind:
			if descriptor.IsMap() {
				err = filterMap(value.Map())
				break
			}
			if descriptor.IsList() {
				err = filterList(value.List())
				break
			}
			err = filterMessage(value.Message())
		}
		return true
	})

	for _, v := range fileReadingQueue {
		file, err := filesystem.ReadFile(v.filename)
		if err != nil {
			return newError("unable to read file").Base(err)
		}
		field := message.Descriptor().Fields().ByTextName(v.field)
		message.Set(field, protoreflect.ValueOf(file))
	}
	return nil
}

func filterMap(mapValue protoreflect.Map) error {
	var err error
	mapValue.Range(func(key protoreflect.MapKey, value protoreflect.Value) bool {
		err = filterMessage(value.Message())
		if err != nil {
			return false
		}
		return true
	})
	return err
}

func filterList(listValue protoreflect.List) error {
	var err error
	size := listValue.Len()
	for i := 0; i < size; i++ {
		err = filterMessage(listValue.Get(i).Message())
		if err != nil {
			return err
		}
	}
	return nil
}
