package protostorage

import (
	"context"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/v2fly/v2ray-core/v5/features/extension/storage"
)

type ProtoPersistentStorage interface {
	PutProto(ctx context.Context, key string, pb proto.Message) error
	GetProto(ctx context.Context, key string, pb proto.Message) error
}

type protoStorage struct {
	storage    storage.ScopedPersistentStorage
	textFormat bool
}

func (p *protoStorage) PutProto(ctx context.Context, key string, pb proto.Message) error {
	if !p.textFormat {
		data, err := proto.Marshal(pb)
		if err != nil {
			return err
		}
		return p.storage.Put(ctx, []byte(key), data)
	} else {
		protojsonStr := protojson.Format(pb)
		return p.storage.Put(ctx, []byte(key), []byte(protojsonStr))
	}
}

func (p *protoStorage) GetProto(ctx context.Context, key string, pb proto.Message) error {
	data, err := p.storage.Get(ctx, []byte(key))
	if err != nil {
		return err
	}
	if !p.textFormat {
		return proto.Unmarshal(data, pb)
	}
	return protojson.Unmarshal(data, pb)
}

func NewProtoStorage(storage storage.ScopedPersistentStorage, textFormat bool) ProtoPersistentStorage {
	return &protoStorage{
		storage:    storage,
		textFormat: textFormat,
	}
}
