package filesystemstorage

import (
	"bytes"
	"context"
	"io"
	"path/filepath"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/app/persistentstorage/protostorage"
	"github.com/v2fly/v2ray-core/v5/common"
	"github.com/v2fly/v2ray-core/v5/common/environment"
	"github.com/v2fly/v2ray-core/v5/common/environment/envctx"
	"github.com/v2fly/v2ray-core/v5/features/extension/storage"
)

func newFileSystemStorage(ctx context.Context, config *Config) storage.ScopedPersistentStorageService {
	appEnvironment := envctx.EnvironmentFromContext(ctx).(environment.AppEnvironment)
	fss := &fileSystemStorage{
		fs:              appEnvironment,
		pathRoot:        config.InstanceName,
		currentLocation: nil,
		config:          config,
	}

	protoStorageInst := protostorage.NewProtoStorage(fss, config.Protojson)
	fss.proto = protoStorageInst
	return fss
}

type fileSystemStorage struct {
	fs    environment.FileSystemCapabilitySet
	proto protostorage.ProtoPersistentStorage

	pathRoot        string
	currentLocation []string
	config          *Config
}

func (f *fileSystemStorage) Type() interface{} {
	return storage.ScopedPersistentStorageServiceType
}

func (f *fileSystemStorage) Start() error {
	return nil
}

func (f *fileSystemStorage) Close() error {
	return nil
}

func (f *fileSystemStorage) PutProto(ctx context.Context, key string, pb proto.Message) error {
	return f.proto.PutProto(ctx, key, pb)
}

func (f *fileSystemStorage) GetProto(ctx context.Context, key string, pb proto.Message) error {
	return f.proto.GetProto(ctx, key, pb)
}

func (f *fileSystemStorage) ScopedPersistentStorageEngine() {
}

func (f *fileSystemStorage) Put(ctx context.Context, key []byte, value []byte) error {
	finalPath := filepath.Join(f.pathRoot, filepath.Join(f.currentLocation...), string(key))
	if value == nil {
		return f.fs.RemoveFile()(finalPath)
	}
	writer, err := f.fs.OpenFileForWrite()(finalPath)
	if err != nil {
		return err
	}
	defer writer.Close()
	_, err = io.Copy(writer, io.NopCloser(bytes.NewReader(value)))
	return err
}

func (f *fileSystemStorage) Get(ctx context.Context, key []byte) ([]byte, error) {
	finalPath := filepath.Join(f.pathRoot, filepath.Join(f.currentLocation...), string(key))
	reader, err := f.fs.OpenFileForRead()(finalPath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

func (f *fileSystemStorage) List(ctx context.Context, keyPrefix []byte) ([][]byte, error) {
	res, err := f.fs.ReadDir()(filepath.Join(f.pathRoot, filepath.Join(f.currentLocation...)))
	if err != nil {
		return nil, err
	}
	var result [][]byte
	for _, entry := range res {
		if !entry.IsDir() && bytes.HasPrefix([]byte(entry.Name()), keyPrefix) {
			result = append(result, []byte(entry.Name()))
		}
	}
	return result, nil
}

func (f *fileSystemStorage) Clear(ctx context.Context) {
	allFile, err := f.List(ctx, []byte{})
	if err != nil {
		return
	}
	for _, file := range allFile {
		_ = f.Put(ctx, file, nil)
	}
}

func (f *fileSystemStorage) NarrowScope(ctx context.Context, key []byte) (storage.ScopedPersistentStorage, error) {
	escapedKey := strings.ReplaceAll(string(key), "/", "_")
	fss := &fileSystemStorage{
		fs:              f.fs,
		pathRoot:        f.pathRoot,
		currentLocation: append(f.currentLocation, escapedKey),
		config:          f.config,
	}
	fss.proto = protostorage.NewProtoStorage(fss, f.config.Protojson)
	return fss, nil
}

func (f *fileSystemStorage) DropScope(ctx context.Context, key []byte) error {
	panic("unimplemented")
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return newFileSystemStorage(ctx, config.(*Config)), nil
	}))
}
