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
	"testing"
)

import (
	"github.com/nacos-group/nacos-sdk-go/common/constant"

	"github.com/stretchr/testify/assert"
)

func TestNewNacosClient(t *testing.T) {

	scs := []constant.ServerConfig{
		*constant.NewServerConfig("console.nacos.io", 8848),
	}

	cc := constant.ClientConfig{
		TimeoutMs:           5 * 1000,
		NotLoadCacheAtStart: true,
	}

	client1, err := NewNacosNamingClient("nacos", true, scs, cc)
	assert.True(t, err == nil && client1 != nil)
	client2, err := NewNacosNamingClient("nacos", true, scs, cc)
	assert.True(t, err == nil && client2 != nil)
	client3, err := NewNacosNamingClient("nacos", false, scs, cc)
	assert.True(t, err == nil && client3 != nil)
	client4, err := NewNacosNamingClient("test", true, scs, cc)
	assert.True(t, err == nil && client4 != nil)

	assert.Equal(t, client1, client2)
	assert.Equal(t, client1.activeCount, uint32(2))
	assert.True(t, client1 != client3)
	assert.True(t, client1 != client4)
}
