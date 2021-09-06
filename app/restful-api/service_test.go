package restful_api

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestTypeReturnAnonymousType(t *testing.T) {
	service := restfulService{}
	serviceType := service.Type()
	assert.Empty(t, reflect.TypeOf(serviceType).Name(), "must return anonymous type")
}
