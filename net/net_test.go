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
	"strings"
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

func TestListenOnTCPRandomPort(t *testing.T) {
	l, err := ListenOnTCPRandomPort("")
	assert.Nil(t, err)
	t.Logf("a tcp server listen on a random addr:%s", l.Addr())
	l.Close()

	localIP, err := GetLocalIP()
	if err == nil {
		l, err = ListenOnTCPRandomPort(localIP)
		assert.Nil(t, err)
		assert.True(t, strings.Contains(l.Addr().String(), localIP))
		t.Logf("a tcp server listen on a random addr:%s", l.Addr())
		l.Close()
	}
}

func TestListenOnUDPRandomPort(t *testing.T) {
	l, err := ListenOnUDPRandomPort("")
	assert.Nil(t, err)
	t.Logf("a udp peer listen on a random addr:%s", l.LocalAddr())
	l.Close()

	localIP, err := GetLocalIP()
	if err == nil {
		l, err = ListenOnUDPRandomPort(localIP)
		assert.Nil(t, err)
		assert.True(t, strings.Contains(l.LocalAddr().String(), localIP))
		t.Logf("a udp server listen on a random addr:%s", l.LocalAddr())
		l.Close()
	}
}

func TestMatchIpIpv4Equal(t *testing.T) {
	assert.True(t, MatchIP("192.168.0.1:8080", "192.168.0.1", "8080"))
	assert.False(t, MatchIP("192.168.0.1:8081", "192.168.0.1", "8080"))
	assert.True(t, MatchIP("*", "192.168.0.1", "8080"))
	assert.True(t, MatchIP("*", "192.168.0.1", ""))
	assert.True(t, MatchIP("*.*.*.*", "192.168.0.1", "8080"))
	assert.False(t, MatchIP("*", "", ""))
}

func TestMatchIpIpv4Subnet(t *testing.T) {
	assert.True(t, MatchIP("206.0.68.0/23", "206.0.68.123", "8080"))
	assert.False(t, MatchIP("206.0.68.0/23", "207.0.69.123", "8080"))
}

func TestMatchIpIpv4Range(t *testing.T) {
	assert.True(t, MatchIP("206.*.68.0", "206.0.68.0", "8080"))
	assert.False(t, MatchIP("206.*.68.0", "206.0.69.0", "8080"))
	assert.True(t, MatchIP("206.0.68-69.0", "206.0.68.0", "8080"))
	assert.False(t, MatchIP("206.0.68-69.0", "206.0.70.0", "8080"))
}

func TestMatchIpIpv6Equal(t *testing.T) {
	assert.True(t, MatchIP("[1fff:0:a88:85a3::ac1f]:8080", "1fff:0:a88:85a3::ac1f", "8080"))
	assert.False(t, MatchIP("[1fff:0:a88:85a3::ac1f]:8081", "1fff:0:a88:85a3::ac1f", "8080"))
	assert.True(t, MatchIP("*", "1fff:0:a88:85a3::ac1f", "8080"))
	assert.True(t, MatchIP("*", "1fff:0:a88:85a3::ac1f", ""))
	assert.True(t, MatchIP("*.*.*.*", "1fff:0:a88:85a3::ac1f", "8080"))
	assert.False(t, MatchIP("*", "", ""))
}

func TestMatchIpIpv6Subnet(t *testing.T) {
	assert.True(t, MatchIP("1fff:0:a88:85a3::ac1f/64", "1fff:0000:0a88:85a3:0000:0000:0000:0000", "8080"))
	assert.False(t, MatchIP("1fff:0:a88:85a3::ac1f/64", "2fff:0000:0a88:85a3:0000:0000:0000:0000", "8080"))
}

func TestMatchIpIpv6Range(t *testing.T) {
	assert.True(t, MatchIP("234e:0:4567:0:0:0:3d:*", "234e:0:4567:0:0:0:3d:4", "8080"))
	assert.False(t, MatchIP("234e:0:4567:0:0:0:3d:*", "234e:0:4567:0:0:0:2d:4", "8080"))
	assert.True(t, MatchIP("234e:0:4567:0:0:0:3d:1-2", "234e:0:4567:0:0:0:3d:1", "8080"))
	assert.False(t, MatchIP("234e:0:4567:0:0:0:3d:1-2", "234e:0:4567:0:0:0:3d:3", "8080"))
}

func TestHostAddress(t *testing.T) {
	assert.Equal(t, "127.0.0.1:8080", HostAddress("127.0.0.1", 8080))
}

func TestWSHostAddress(t *testing.T) {
	assert.Equal(t, "ws://127.0.0.1:8080ws", WSHostAddress("127.0.0.1", 8080, "ws"))
}

func TestWSSHostAddress(t *testing.T) {
	assert.Equal(t, "wss://127.0.0.1:8080wss", WSSHostAddress("127.0.0.1", 8080, "wss"))
}

func TestHostAddress2(t *testing.T) {
	assert.Equal(t, "127.0.0.1:8080", HostAddress2("127.0.0.1", "8080"))
}

func TestWSHostAddress2(t *testing.T) {
	assert.Equal(t, "ws://127.0.0.1:8080ws", WSHostAddress2("127.0.0.1", "8080", "ws"))
}

func TestWSSHostAddress2(t *testing.T) {
	assert.Equal(t, "wss://127.0.0.1:8080wss", WSSHostAddress2("127.0.0.1", "8080", "wss"))
	assert.False(t, WSSHostAddress2("127.0.0.1", "8080", "wss") == "wss://127.0.0.1:8081wss")
}

func TestHostPort(t *testing.T) {
	host, port, err := HostPort("127.0.0.1:8080")
	assert.Equal(t, "127.0.0.1", host)
	assert.Equal(t, "8080", port)
	assert.Nil(t, err)
}
