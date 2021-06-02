/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package nacos

import (
	"sync"
	"sync/atomic"
)

import (
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var (
	namingClientPool nacosClientPool
	clientPoolOnce   sync.Once
)

type nacosClientPool struct {
	sync.Mutex
	namingClient map[string]*NacosNamingClient
}

type NacosNamingClient struct {
	name        string
	clientLock  sync.Mutex // for Client
	client      naming_client.INamingClient
	config      vo.NacosClientParam //conn config
	valid       uint32
	activeCount uint32
	share       bool
}

func initNacosClientPool() {
	namingClientPool.namingClient = make(map[string]*NacosNamingClient)
}

// NewNacosNamingClient create nacos client
func NewNacosNamingClient(name string, share bool, sc []constant.ServerConfig,
	cc constant.ClientConfig) (*NacosNamingClient, error) {

	namingClient := &NacosNamingClient{
		name:        name,
		activeCount: 0,
		share:       share,
		config:      vo.NacosClientParam{ClientConfig: &cc, ServerConfigs: sc},
	}
	if !share {
		return namingClient, namingClient.newNamingClient()
	}
	clientPoolOnce.Do(initNacosClientPool)
	namingClientPool.Lock()
	defer namingClientPool.Unlock()
	if client, ok := namingClientPool.namingClient[name]; ok {
		client.activeCount++
		return client, nil
	}

	err := namingClient.newNamingClient()
	if err == nil {
		namingClientPool.namingClient[name] = namingClient
	}
	return namingClient, err
}

// newNamingClient create NamingClient
func (n *NacosNamingClient) newNamingClient() error {
	client, err := clients.NewNamingClient(n.config)
	if err != nil {
		return err
	}
	n.activeCount++
	atomic.StoreUint32(&n.valid, 1)
	n.client = client
	return nil
}

// Client Get NacosNamingClient
func (n *NacosNamingClient) Client() naming_client.INamingClient {
	return n.client
}

// SetClient Set NacosNamingClient
func (n *NacosNamingClient) SetClient(client naming_client.INamingClient) {
	n.clientLock.Lock()
	n.client = client
	n.clientLock.Unlock()
}

// NacosClientValid Get nacos client valid status
func (n *NacosNamingClient) NacosClientValid() bool {

	return atomic.LoadUint32(&n.valid) == 1
}

// Close close client
func (n *NacosNamingClient) Close() {
	namingClientPool.Lock()
	defer namingClientPool.Unlock()
	if n.client == nil {
		return
	}
	n.activeCount--
	if n.share {
		if n.activeCount == 0 {
			n.client = nil
			atomic.StoreUint32(&n.valid, 0)
			delete(namingClientPool.namingClient, n.name)
		}
	} else {
		n.client = nil
		atomic.StoreUint32(&n.valid, 0)
		delete(namingClientPool.namingClient, n.name)
	}
}
