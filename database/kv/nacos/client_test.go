package nacos

import "testing"

import (
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/stretchr/testify/assert"
)

func TestNewNamingClient(t *testing.T) {
	scs := make([]constant.ServerConfig, 0, 1)
	scs = append(scs, constant.ServerConfig{IpAddr: "console.nacos.io", Port: 80})

	cc := constant.ClientConfig{
		TimeoutMs:           5 * 1000,
		NotLoadCacheAtStart: true,
	}

	client1, err := NewNamingClient("nacos", true, scs, cc)
	client2, err := NewNamingClient("nacos", true, scs, cc)
	client3, err := NewNamingClient("nacos", false, scs, cc)
	client4, err := NewNamingClient("test", true, scs, cc)

	assert.Nil(t, err)
	assert.Equal(t, client1, client2)
	assert.NotEqual(t, client1, client3)
	assert.NotEqual(t, client1, client4)
}
