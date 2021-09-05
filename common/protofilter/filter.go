package protofilter

import (
	"context"
	"github.com/v2fly/v2ray-core/v4/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v4/common/environment/filesystemcap"
	"github.com/v2fly/v2ray-core/v4/common/protoext"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"io"
)

//go:generate go run github.com/v2fly/v2ray-core/v4/common/errors/errorgen

func FilterProtoConfig(ctx context.Context, config proto.Message) error {
	messageProtoReflect := config.ProtoReflect()
	return filterMessage(ctx, messageProtoReflect)
}

func filterMessage(ctx context.Context, message protoreflect.Message) error {
	var err error
	type fileRead struct {
		filename string
		field    string
	}
	var fileReadingQueue []fileRead
	message.Range(func(descriptor protoreflect.FieldDescriptor, value protoreflect.Value) bool {
		v2extension, ferr := protoext.GetFieldOptions(descriptor)
		if ferr == nil {
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
				err = filterMap(ctx, value.Map())
				break
			}
			if descriptor.IsList() {
				err = filterList(ctx, value.List())
				break
			}
			err = filterMessage(ctx, value.Message())
		}
		return true
	})

	if err != nil {
		return err
	}

	fsenvironment := envctx.EnvironmentFromContext(ctx)
	fsifce := fsenvironment.(filesystemcap.FileSystemCapabilitySet)
	for _, v := range fileReadingQueue {
		field := message.Descriptor().Fields().ByTextName(v.field)
		if v.filename == "" {
			continue
		}

		if len(message.Get(field).Bytes()) > 0 {
			continue
		}

		file, err := fsifce.OpenFileForRead()(v.filename)
		if err != nil {
			return newError("unable to open file").Base(err)
		}
		fileContent, err := io.ReadAll(file)
		if err != nil {
			return newError("unable to read file").Base(err)
		}
		file.Close()
		message.Set(field, protoreflect.ValueOf(fileContent))
	}
	return nil
}

func filterMap(ctx context.Context, mapValue protoreflect.Map) error {
	var err error
	mapValue.Range(func(key protoreflect.MapKey, value protoreflect.Value) bool {
		err = filterMessage(ctx, value.Message())
		if err != nil {
			return false
		}
		return true
	})
	return err
}

func filterList(ctx context.Context, listValue protoreflect.List) error {
	var err error
	size := listValue.Len()
	for i := 0; i < size; i++ {
		err = filterMessage(ctx, listValue.Get(i).Message())
		if err != nil {
			return err
		}
	}
	return nil
}
