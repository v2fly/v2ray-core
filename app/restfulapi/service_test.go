package restfulapi

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeReturnAnonymousType(t *testing.T) {
	service := restfulService{}
	serviceType := service.Type()
	assert.Empty(t, reflect.TypeOf(serviceType).Name(), "must return anonymous type")
}
