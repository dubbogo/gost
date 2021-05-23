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
