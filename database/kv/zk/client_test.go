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

package gxzookeeper

import (
	"net"
	"os"
	"path"
	"testing"
	"time"
)

import (
	"github.com/go-zookeeper/zk"

	"github.com/stretchr/testify/assert"
)

// zkTestAddrEnvKey is the environment variable used to point these tests at
// a real zookeeper server.
const zkTestAddrEnvKey = "ZK_ADDR"

// defaultZkTestAddr is used when ZK_ADDR is not set.
const defaultZkTestAddr = "127.0.0.1:2181"

// testZkAddr returns the zookeeper address these tests should run against.
//
// github.com/go-zookeeper/zk (the upstream package this module now depends
// on) does not expose an importable TestCluster/StartTestCluster API the
// way the previous github.com/dubbogo/go-zookeeper fork did (upstream only
// has it in the module's own _test.go files, which cannot be imported), so
// these tests can no longer boot an in-process zk server on demand. Instead
// they require a real, reachable zookeeper server; point ZK_ADDR at one
// (defaults to 127.0.0.1:2181). When no server is reachable, the calling
// test is skipped rather than failed.
func testZkAddr(t *testing.T) string {
	addr := os.Getenv(zkTestAddrEnvKey)
	if addr == "" {
		addr = defaultZkTestAddr
	}
	conn, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Skipf("skipping: no zookeeper server reachable at %s (set %s to run this test): %v", addr, zkTestAddrEnvKey, err)
		return ""
	}
	_ = conn.Close()
	return addr
}

// cleanupZkPath best-effort recursively removes a zk path (and its
// children) so tests that assert on Create() are repeatable against a
// long-lived real zk server instead of a freshly booted test cluster.
func cleanupZkPath(z *ZookeeperClient, zkPath string) {
	if z == nil || z.Conn == nil {
		return
	}
	children, _, err := z.Conn.Children(zkPath)
	if err == nil {
		for _, c := range children {
			cleanupZkPath(z, path.Join(zkPath, c))
		}
	}
	_ = z.Conn.Delete(zkPath, -1)
}

func verifyEventStateOrder(t *testing.T, c <-chan zk.Event, expectedStates []zk.State, source string) {
	for _, state := range expectedStates {
		for {
			event, ok := <-c
			if !ok {
				t.Fatalf("unexpected channel close for %s", source)
			}
			if event.Type != zk.EventSession {
				continue
			}

			if event.State != state {
				t.Fatalf("mismatched state order from %s, expected %v, received %v", source, state, event.State)
			}
			break
		}
	}
}

func Test_getZookeeperClient(t *testing.T) {
	addr := testZkAddr(t)
	address := []string{addr}

	client1, err := NewZookeeperClient("test1", address, true, WithZkTimeOut(3*time.Second))
	assert.Nil(t, err)
	client2, err := NewZookeeperClient("test1", address, true, WithZkTimeOut(3*time.Second))
	assert.Nil(t, err)
	client3, err := NewZookeeperClient("test2", address, true, WithZkTimeOut(3*time.Second))
	assert.Nil(t, err)
	client4, err := NewZookeeperClient("test2", address, false, WithZkTimeOut(3*time.Second))
	assert.Nil(t, err)
	if client1 != client2 {
		t.Fatalf("NewZookeeperClient failed")
	}
	if client1 == client3 {
		t.Fatalf("NewZookeeperClient failed")
	}
	if client3 == client4 {
		t.Fatalf("NewZookeeperClient failed")
	}
	client1.Close()
	client2.Close()
	client3.Close()
	client4.Close()
}

func Test_Close(t *testing.T) {
	addr := testZkAddr(t)
	address := []string{addr}

	client1, err := NewZookeeperClient("test1", address, true, WithZkTimeOut(3*time.Second))
	assert.Nil(t, err)
	client2, err := NewZookeeperClient("test1", address, true, WithZkTimeOut(3*time.Second))
	assert.Nil(t, err)
	if client1 != client2 {
		t.Fatalf("NewZookeeperClient failed")
	}
	client1.Close()
	client3, err := NewZookeeperClient("test1", address, true, WithZkTimeOut(3*time.Second))
	assert.Nil(t, err)
	if client2 != client3 {
		t.Fatalf("NewZookeeperClient failed")
	}
	client2.Close()
	assert.Equal(t, client1.activeNumber, uint32(1))
	client1.Close()
	assert.Equal(t, client1.activeNumber, uint32(0))
	client4, err := NewZookeeperClient("test1", address, true, WithZkTimeOut(3*time.Second))
	assert.Nil(t, err)
	assert.Equal(t, client4.activeNumber, uint32(1))
	if client4 == client3 {
		t.Fatalf("NewZookeeperClient failed")
	}
	client5, err := NewZookeeperClient("test1", address, false, WithZkTimeOut(3*time.Second))
	assert.Nil(t, err)
	client6, err := NewZookeeperClient("test1", address, false, WithZkTimeOut(3*time.Second))
	assert.Nil(t, err)
	if client5 == client6 {
		t.Fatalf("NewZookeeperClient failed")
	}
	client5.Close()
	assert.Equal(t, client5.activeNumber, uint32(0))
	assert.Equal(t, client5.Conn, (*zk.Conn)(nil))
	assert.NotEqual(t, client6.Conn, nil)
	client6.Close()
	assert.Equal(t, client6.activeNumber, uint32(0))
	assert.Equal(t, client6.Conn, (*zk.Conn)(nil))
	client4.Close()
}

func Test_newMockZookeeperClient(t *testing.T) {
	testZkAddr(t)

	z, event, err := NewMockZookeeperClient("test", 15*time.Second)
	assert.NoError(t, err)
	defer z.Close()
	states := []zk.State{zk.StateConnecting, zk.StateConnected, zk.StateHasSession}
	verifyEventStateOrder(t, event, states, "event channel")
}

func TestCreate(t *testing.T) {
	testZkAddr(t)

	z, event, err := NewMockZookeeperClient("test", 15*time.Second)
	assert.NoError(t, err)
	defer z.Close()

	states := []zk.State{zk.StateConnecting, zk.StateConnected, zk.StateHasSession}
	verifyEventStateOrder(t, event, states, "event channel")

	cleanupZkPath(z, "/test1")
	err = z.Create("/test1/test2/test3/test4")
	assert.NoError(t, err)
}

func TestCreateDelete(t *testing.T) {
	testZkAddr(t)

	z, event, err := NewMockZookeeperClient("test", 15*time.Second)
	assert.NoError(t, err)
	defer z.Close()

	states := []zk.State{zk.StateConnecting, zk.StateConnected, zk.StateHasSession}
	verifyEventStateOrder(t, event, states, "event channel")

	cleanupZkPath(z, "/test1")
	err = z.Create("/test1/test2/test3/test4")
	assert.NoError(t, err)
	err = z.Delete("/test1/test2/test3/test4")
	assert.NoError(t, err)
	// verifyEventOrder(t, event, []zk.EventType{zk.EventNodeCreated}, "event channel")
}

func TestRegisterTemp(t *testing.T) {
	testZkAddr(t)

	z, event, err := NewMockZookeeperClient("test", 15*time.Second)
	assert.NoError(t, err)
	defer z.Close()

	cleanupZkPath(z, "/test1")
	err = z.Create("/test1/test2/test3")
	assert.NoError(t, err)

	tmpath, err := z.RegisterTemp("/test1/test2/test3", "test4")
	assert.NoError(t, err)
	assert.Equal(t, "/test1/test2/test3/test4", tmpath)
	states := []zk.State{zk.StateConnecting, zk.StateConnected, zk.StateHasSession}
	verifyEventStateOrder(t, event, states, "event channel")
}

func TestRegisterTempSeq(t *testing.T) {
	testZkAddr(t)

	z, event, err := NewMockZookeeperClient("test", 15*time.Second)
	assert.NoError(t, err)
	defer z.Close()

	cleanupZkPath(z, "/test1")
	err = z.Create("/test1/test2/test3")
	assert.NoError(t, err)
	tmpath, err := z.RegisterTempSeq("/test1/test2/test3", []byte("test"))
	assert.NoError(t, err)
	assert.Equal(t, "/test1/test2/test3/0000000000", tmpath)
	states := []zk.State{zk.StateConnecting, zk.StateConnected, zk.StateHasSession}
	verifyEventStateOrder(t, event, states, "event channel")
}

func Test_UnregisterEvent(t *testing.T) {
	client := &ZookeeperClient{}
	client.eventRegistry = make(map[string][]chan zk.Event)
	mockEvent := make(chan zk.Event, 1)
	var array []chan zk.Event
	array = append(array, mockEvent)
	client.eventRegistry["test"] = array
	client.UnregisterEvent("test", mockEvent)
}
