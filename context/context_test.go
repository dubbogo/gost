package gxcontext

import (
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestValuesContext_General(t *testing.T) {
	vc := NewValuesContext(nil)
	assert.NotNil(t, vc)

	key := "hello"
	value := "world"
	v, ok := vc.Get(key)
	assert.Nil(t, v)
	assert.False(t, ok)

	vc.Set(key, value)
	v, ok = vc.Get(key)
	assert.Equal(t, v.(string), value)
	assert.True(t, ok)
}
