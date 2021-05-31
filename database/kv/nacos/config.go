package nacos

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"sync"
	"sync/atomic"
)

import "github.com/nacos-group/nacos-sdk-go/clients/config_client"

type NacosConfigClient struct {
	name         string
	sync.RWMutex // for Client
	client       *config_client.IConfigClient
	Config       vo.NacosClientParam //conn config
	valid        uint32
	once         sync.Once
	share        bool
	activeNumber uint32
}

var (
	configClientPool     nacosConfigClientPool
	configClientPoolOnce sync.Once
)

type nacosConfigClientPool struct {
	sync.Mutex
	configClient map[string]*NacosConfigClient
}

func initNacosConfigClientPool() {
	configClientPool.configClient = make(map[string]*NacosConfigClient)
}

func (n *NacosConfigClient) newConfigClient() error {
	client, err := clients.NewConfigClient(n.Config)
	if err != nil {
		return err
	}
	atomic.StoreUint32(&n.valid, 1)
	n.activeNumber++
	n.client = &client
	return nil
}

// NewNacosConfigClient create config client
func NewNacosConfigClient(name string, share bool, sc []constant.ServerConfig,
	cc constant.ClientConfig) (*NacosConfigClient, error) {

	configClient := &NacosConfigClient{
		name:         name,
		activeNumber: 0,
		share:        share,
		Config:       vo.NacosClientParam{ClientConfig: &cc, ServerConfigs: sc},
	}
	if !share {
		return configClient, configClient.newConfigClient()
	}
	configClientPoolOnce.Do(initNacosConfigClientPool)
	configClientPool.Lock()
	defer configClientPool.Unlock()
	if client, ok := configClientPool.configClient[name]; ok {
		return client, nil
	}

	err := configClient.newConfigClient()
	if err == nil {
		configClientPool.configClient[name] = configClient
	}
	return configClient, err
}

// Client Get NacosConfigClient
func (n *NacosConfigClient) Client() *config_client.IConfigClient {
	return n.client
}

// SetClient Set NacosConfigClient
func (n *NacosConfigClient) SetClient(client *config_client.IConfigClient) {

	n.Lock()
	n.client = client
	n.Unlock()
}

// NacosClientValid Get nacos client valid status
func (n *NacosConfigClient) NacosClientValid() bool {

	return atomic.LoadUint32(&n.valid) == 1
}
