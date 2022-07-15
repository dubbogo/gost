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
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"
	"time"
)

import (
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"

	"github.com/stretchr/testify/assert"
)

func TestStructAlign(t *testing.T) {
	typ := reflect.TypeOf(NacosConfigClient{})
	fmt.Printf("Struct is %d bytes long\n", typ.Size())
	n := typ.NumField()
	for i := 0; i < n; i++ {
		field := typ.Field(i)
		fmt.Printf("%s: at offset %v, size=%d, align=%d\n",
			field.Name, field.Offset, field.Type.Size(),
			field.Type.Align())
	}
}

//TestNewNacosConfigClient config client
func TestNewNacosConfigClient(t *testing.T) {

	scs := []constant.ServerConfig{*constant.NewServerConfig("console.nacos.io", 80)}
	cc := constant.ClientConfig{TimeoutMs: 5 * 1000, NotLoadCacheAtStart: true}

	client1, err := NewNacosConfigClient("nacos", true, scs, cc)
	assert.Nil(t, err)
	client2, err := NewNacosConfigClient("nacos", true, scs, cc)
	assert.Nil(t, err)
	client3, err := NewNacosConfigClient("nacos", false, scs, cc)
	assert.Nil(t, err)
	client4, err := NewNacosConfigClient("test", true, scs, cc)
	assert.Nil(t, err)

	assert.Equal(t, client1, client2)
	assert.Equal(t, client1.activeCount, uint32(2))
	assert.Equal(t, client1.NacosClientValid(), true)
	assert.Equal(t, client3.activeCount, uint32(1))
	assert.Equal(t, client4.activeCount, uint32(1))

	client1.Close()
	assert.Equal(t, client1.activeCount, uint32(1))
	client1.Close()

	assert.Equal(t, client1.NacosClientValid(), false)
	assert.Nil(t, client1.Client())

	client1.Close()
	assert.Equal(t, client1.NacosClientValid(), false)
	assert.Nil(t, client1.Client())
}

func TestPublishConfig(t *testing.T) {

	scs := []constant.ServerConfig{*constant.NewServerConfig("console.nacos.io", 80)}

	cc := constant.ClientConfig{
		AppName:             "nacos",
		NamespaceId:         "14e01fa8-a4aa-44cc-ad5b-c768f3c62bd5", //namespace id
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		RotateTime:          "1h",
		MaxAge:              3,
		LogLevel:            "debug",
	}

	client, clientErr := NewNacosConfigClient("nacos", true, scs, cc)

	assert.Nil(t, clientErr)

	defer client.Close()

	t.Run("publishConfig", func(t *testing.T) {
		//publish config
		//config key=dataId+group+namespaceId
		push, err := client.Client().PublishConfig(vo.ConfigParam{
			DataId:  "nacos-config",
			Group:   "dubbo",
			Content: "dubbo-go nb",
		})
		assert.Nil(t, err)
		assert.Equal(t, push, true)
	})

	t.Run("getConfig", func(t *testing.T) {
		//get config
		cfg, err := client.Client().GetConfig(vo.ConfigParam{
			DataId: "nacos-config",
			Group:  "dubbo",
		})
		assert.Nil(t, err)
		fmt.Println("GetConfig,config :", cfg)
	})

	t.Run("listenConfig", func(t *testing.T) {
		randomizer := rand.New(rand.NewSource(time.Now().UnixNano()))
		key := strconv.Itoa(randomizer.Intn(100))
		//Listen config change,key=dataId+group+namespaceId.
		err := client.Client().ListenConfig(vo.ConfigParam{
			DataId: "nacos-config" + key,
			Group:  "dubbo",
			OnChange: func(namespace, group, dataId, data string) {
				assert.Equal(t, data, "test-listen")
				fmt.Println("config changed group:" + group + ", dataId:" + dataId + ", content:" + data)
			},
		})
		assert.Nil(t, err)

		_, err = client.Client().PublishConfig(vo.ConfigParam{
			DataId:  "nacos-config" + key,
			Group:   "dubbo",
			Content: "test-listen",
		})
		assert.Nil(t, err)

		time.Sleep(2 * time.Second)
		_, err = client.Client().DeleteConfig(vo.ConfigParam{
			DataId: "nacos-config" + key,
			Group:  "dubbo"})
		assert.Nil(t, err)
	})

	t.Run("searchConfig", func(t *testing.T) {
		searchPage, err := client.Client().SearchConfig(vo.SearchConfigParam{
			Search:   "accurate",
			DataId:   "",
			Group:    "dubbo",
			PageNo:   1,
			PageSize: 10,
		})
		fmt.Printf("Search config:%+v \n", searchPage)
		assert.Nil(t, err)
	})
}
