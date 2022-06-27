package taggedfeatures

import (
	"context"
	"reflect"
	"sync"

	"github.com/v2fly/v2ray-core/v5/common/task"
	"github.com/v2fly/v2ray-core/v5/features"
)

type Holder struct {
	access     *sync.RWMutex
	features   map[string]features.Feature
	memberType reflect.Type
	ctx        context.Context
}

func NewHolder(ctx context.Context, memberType interface{}) *Holder {
	return &Holder{
		ctx:        ctx,
		access:     &sync.RWMutex{},
		features:   make(map[string]features.Feature),
		memberType: reflect.TypeOf(memberType),
	}
}

func (h *Holder) GetFeaturesByTag(tag string) (features.Feature, error) {
	h.access.RLock()
	defer h.access.RUnlock()
	feature, ok := h.features[tag]
	if !ok {
		return nil, newError("unable to find feature with tag")
	}
	return feature, nil
}

func (h *Holder) AddFeaturesByTag(tag string, feature features.Feature) error {
	h.access.Lock()
	defer h.access.Unlock()
	featureType := reflect.TypeOf(feature.Type())
	if !featureType.AssignableTo(h.memberType) {
		return newError("feature is not assignable to the base type")
	}
	h.features[tag] = feature
	return nil
}

func (h *Holder) RemoveFeaturesByTag(tag string) error {
	h.access.Lock()
	defer h.access.Unlock()
	delete(h.features, tag)
	return nil
}

func (h *Holder) GetFeaturesTag() ([]string, error) {
	h.access.RLock()
	defer h.access.RUnlock()
	var ret []string
	for key := range h.features {
		ret = append(ret, key)
	}
	return ret, nil
}

func (h *Holder) Start() error {
	h.access.Lock()
	defer h.access.Unlock()
	var startTasks []func() error
	for _, v := range h.features {
		startTasks = append(startTasks, v.Start)
	}
	return task.Run(h.ctx, startTasks...)
}

func (h *Holder) Close() error {
	h.access.Lock()
	defer h.access.Unlock()

	var closeTasks []func() error
	for _, v := range h.features {
		closeTasks = append(closeTasks, v.Close)
	}
	return task.Run(h.ctx, closeTasks...)
}
