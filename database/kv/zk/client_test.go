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
	"crypto/rand"
	"encoding/hex"
	"fmt"
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

// testRootPath is a zookeeper path unique to this test run, generated once
// per package invocation by newTestRootPath. Every test derives its node
// paths from it and cleanupZkPath only ever clears this path, so tests keep
// working against a shared, long-lived zookeeper server without touching
// data created by other test runs or other clients.
var testRootPath = newTestRootPath()

// newTestRootPath returns a randomly named root path of the form
// "/gost-test-<hex>", used to give each test run its own zookeeper
// namespace.
func newTestRootPath() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Sprintf("newTestRootPath: %v", err))
	}
	return "/gost-test-" + hex.EncodeToString(buf)
}

// testZkAddr returns the zookeeper address these tests should run against,
// reading the ZK_ADDR environment variable and falling back to
// 127.0.0.1:2181 when it is unset. It verifies the server is reachable and
// waits for a probe connection to reach zk.StateHasSession, skipping the
// calling test if either check fails or times out.
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

	probe, _, err := zk.Connect([]string{addr}, time.Second)
	if err != nil {
		t.Skipf("skipping: could not connect to zookeeper at %s: %v", addr, err)
		return ""
	}
	defer probe.Close()
	if err := waitForSession(probe, sessionEstablishTimeout, sessionPollInterval); err != nil {
		t.Skipf("skipping: zookeeper at %s never reported a session: %v", addr, err)
		return ""
	}

	return addr
}

// cleanupZkPath best-effort recursively removes zkPath and all of its
// descendants, letting tests clear their run-scoped root (testRootPath)
// before creating nodes so runs are repeatable against a shared, long-lived
// real zk server instead of a freshly booted test cluster.
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

// verifyEventStateOrder asserts that c delivers session events whose states
// match expectedStates, in order, failing t if the channel closes early or
// the states diverge.
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

// Test_getZookeeperClient verifies that NewZookeeperClient returns the same
// *ZookeeperClient for repeated calls with the same name when share is true,
// and distinct clients otherwise.
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

// Test_Close verifies that Close decrements a shared client's reference
// count and only closes the underlying connection once that count reaches
// zero, while non-shared clients close independently of one another.
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

// Test_newMockZookeeperClient verifies that NewMockZookeeperClient connects
// successfully and that its event channel delivers the expected sequence of
// session state events.
func Test_newMockZookeeperClient(t *testing.T) {
	testZkAddr(t)

	z, event, err := NewMockZookeeperClient("test", 15*time.Second)
	assert.NoError(t, err)
	defer z.Close()
	states := []zk.State{zk.StateConnecting, zk.StateConnected, zk.StateHasSession}
	verifyEventStateOrder(t, event, states, "event channel")
}

// TestCreate verifies that Create creates a node along with any missing
// ancestor nodes.
func TestCreate(t *testing.T) {
	testZkAddr(t)

	z, event, err := NewMockZookeeperClient("test", 15*time.Second)
	assert.NoError(t, err)
	defer z.Close()
	defer cleanupZkPath(z, testRootPath)

	states := []zk.State{zk.StateConnecting, zk.StateConnected, zk.StateHasSession}
	verifyEventStateOrder(t, event, states, "event channel")

	cleanupZkPath(z, testRootPath)
	err = z.Create(testRootPath + "/test2/test3/test4")
	assert.NoError(t, err)
}

// TestCreateDelete verifies that a node created with Create can subsequently
// be removed with Delete.
func TestCreateDelete(t *testing.T) {
	testZkAddr(t)

	z, event, err := NewMockZookeeperClient("test", 15*time.Second)
	assert.NoError(t, err)
	defer z.Close()
	defer cleanupZkPath(z, testRootPath)

	states := []zk.State{zk.StateConnecting, zk.StateConnected, zk.StateHasSession}
	verifyEventStateOrder(t, event, states, "event channel")

	cleanupZkPath(z, testRootPath)
	err = z.Create(testRootPath + "/test2/test3/test4")
	assert.NoError(t, err)
	err = z.Delete(testRootPath + "/test2/test3/test4")
	assert.NoError(t, err)
	// verifyEventOrder(t, event, []zk.EventType{zk.EventNodeCreated}, "event channel")
}

// TestRegisterTemp verifies that RegisterTemp creates an ephemeral child
// node at the expected path.
func TestRegisterTemp(t *testing.T) {
	testZkAddr(t)

	z, event, err := NewMockZookeeperClient("test", 15*time.Second)
	assert.NoError(t, err)
	defer z.Close()
	defer cleanupZkPath(z, testRootPath)

	cleanupZkPath(z, testRootPath)
	err = z.Create(testRootPath + "/test2/test3")
	assert.NoError(t, err)

	tmpath, err := z.RegisterTemp(testRootPath+"/test2/test3", "test4")
	assert.NoError(t, err)
	assert.Equal(t, testRootPath+"/test2/test3/test4", tmpath)
	states := []zk.State{zk.StateConnecting, zk.StateConnected, zk.StateHasSession}
	verifyEventStateOrder(t, event, states, "event channel")
}

// TestRegisterTempSeq verifies that RegisterTempSeq creates an ephemeral
// sequential child node, starting at sequence number 0 for a freshly created
// parent.
func TestRegisterTempSeq(t *testing.T) {
	testZkAddr(t)

	z, event, err := NewMockZookeeperClient("test", 15*time.Second)
	assert.NoError(t, err)
	defer z.Close()
	defer cleanupZkPath(z, testRootPath)

	cleanupZkPath(z, testRootPath)
	err = z.Create(testRootPath + "/test2/test3")
	assert.NoError(t, err)
	tmpath, err := z.RegisterTempSeq(testRootPath+"/test2/test3", []byte("test"))
	assert.NoError(t, err)
	assert.Equal(t, testRootPath+"/test2/test3/0000000000", tmpath)
	states := []zk.State{zk.StateConnecting, zk.StateConnected, zk.StateHasSession}
	verifyEventStateOrder(t, event, states, "event channel")
}

// Test_UnregisterEvent verifies that UnregisterEvent removes a previously
// registered event channel without panicking.
func Test_UnregisterEvent(t *testing.T) {
	client := &ZookeeperClient{}
	client.eventRegistry = make(map[string][]chan zk.Event)
	mockEvent := make(chan zk.Event, 1)
	var array []chan zk.Event
	array = append(array, mockEvent)
	client.eventRegistry["test"] = array
	client.UnregisterEvent("test", mockEvent)
}
