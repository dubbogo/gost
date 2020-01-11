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

package gxnet

import (
	"net"
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestGetLocalIP(t *testing.T) {
	ip, err := GetLocalIP()
	assert.NoError(t, err)
	t.Log(ip)
}

func TestIsSameAddr(t *testing.T) {
	addr1 := net.TCPAddr{
		IP:   []byte("192.168.0.1"),
		Port: 80,
		Zone: "cn/shanghai",
	}
	addr2 := net.TCPAddr{
		IP:   []byte("192.168.0.2"),
		Port: 80,
		Zone: "cn/shanghai",
	}

	assert.True(t, IsSameAddr(&addr1, &addr1))
	assert.False(t, IsSameAddr(&addr1, &addr2))

	addr1.IP = []byte("2001:4860:0:2001::68")
	addr1.Zone = ""

	addr2.IP = []byte("2001:4860:0000:2001:0000:0000:0000:0068")
	addr2.Zone = ""
	assert.True(t, IsSameAddr(&addr1, &addr1))
}
