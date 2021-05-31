package nacos

import (
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewNacosConfigClient(t *testing.T) {
	scs := []constant.ServerConfig{
		*constant.NewServerConfig("console.nacos.io", 80),
	}

	cc := constant.ClientConfig{
		TimeoutMs:           5 * 1000,
		NotLoadCacheAtStart: true,
	}

	client1, err := NewNacosConfigClient("nacos", true, scs, cc)
	client2, err := NewNacosConfigClient("nacos", true, scs, cc)
	client3, err := NewNacosConfigClient("nacos", false, scs, cc)
	client4, err := NewNacosConfigClient("test", true, scs, cc)

	assert.Nil(t, err)
	assert.Equal(t, client1, client2)
	assert.Equal(t, client1, client3)
	assert.Equal(t, client1, client4)

}
