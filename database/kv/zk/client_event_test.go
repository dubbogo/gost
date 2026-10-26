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
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/stretchr/testify/require"
)

func TestDefaultHandlerDispatchesConnectedNodeEvents(t *testing.T) {
	kinds := []zk.EventType{
		zk.EventNodeCreated,
		zk.EventNodeDataChanged,
		zk.EventNodeDeleted,
		zk.EventNodeChildrenChanged,
	}
	for _, kind := range kinds {
		t.Run(zk.StateSyncConnected.String()+"/"+kind.String(), func(t *testing.T) {
			assertPrefixDispatch(t, zk.StateSyncConnected, kind)
		})
	}
	// These StateConnected cases record historical behavior. They do not
	// distinguish this fix: the client synthesizes StateConnected with an
	// empty Path, so a real session event never reaches a prefix subscriber.
	for _, kind := range kinds {
		t.Run(zk.StateConnected.String()+"/"+kind.String(), func(t *testing.T) {
			assertPrefixDispatch(t, zk.StateConnected, kind)
		})
	}
}

func assertPrefixDispatch(t *testing.T, state zk.State, kind zk.EventType) {
	t.Helper()
	session := make(chan zk.Event, 1)
	client := &ZookeeperClient{
		Session:       session,
		eventRegistry: make(map[string][]chan zk.Event),
	}
	matched := make(chan zk.Event, 1)
	unrelated := make(chan zk.Event, 1)
	client.RegisterEvent("/dubbo/config/", matched)
	client.RegisterEvent("/other/", unrelated)
	want := zk.Event{Type: kind, State: state, Path: "/dubbo/config/demo.tag-router"}
	session <- want
	close(session)

	(&DefaultHandler{}).HandleZkEvent(client)

	select {
	case got := <-matched:
		require.Equal(t, want, got)
	default:
		t.Fatal("connected node event was not delivered to its prefix subscriber")
	}
	select {
	case got := <-unrelated:
		t.Fatalf("unrelated subscriber received %+v", got)
	default:
	}
}

// A subscriber that nobody reads must not keep eventRegistryLock held.
// RegisterEvent has to return while that send is still blocked. The test
// does not show that later subscribers run concurrently: sends stay sequential.
func TestDispatchDoesNotHoldRegistryLockWhileSubscriberBlocks(t *testing.T) {
	session := make(chan zk.Event, 2)
	client := &ZookeeperClient{
		Session:       session,
		eventRegistry: make(map[string][]chan zk.Event),
	}
	ready := make(chan zk.Event, 1)
	stuck := make(chan zk.Event)
	client.RegisterEvent("/ready/", ready)
	client.RegisterEvent("/dubbo/config/", stuck)

	session <- zk.Event{Type: zk.EventNodeCreated, State: zk.StateSyncConnected, Path: "/ready/up"}
	session <- zk.Event{Type: zk.EventNodeCreated, State: zk.StateSyncConnected, Path: "/dubbo/config/demo.tag-router"}
	close(session)

	finished := make(chan struct{})
	go func() {
		defer close(finished)
		(&DefaultHandler{}).HandleZkEvent(client)
	}()

	select {
	case <-ready:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not deliver the non-blocking event")
	}
	// The next event is already buffered. Give the handler time to block in
	// the unread subscriber's send before taking the registry write lock.
	time.Sleep(100 * time.Millisecond)

	registered := make(chan struct{})
	go func() {
		client.RegisterEvent("/other/", make(chan zk.Event, 1))
		close(registered)
	}()
	select {
	case <-registered:
	case <-time.After(time.Second):
		t.Fatal("RegisterEvent blocked while a subscriber send was in progress")
	}

	select {
	case got := <-stuck:
		require.Equal(t, "/dubbo/config/demo.tag-router", got.Path)
	case <-time.After(2 * time.Second):
		t.Fatal("blocked subscriber did not receive the event")
	}
	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not return after the subscriber drained")
	}
}

func TestRegisterEventIgnoresDuplicateRegistration(t *testing.T) {
	session := make(chan zk.Event, 1)
	client := &ZookeeperClient{
		Session:       session,
		eventRegistry: make(map[string][]chan zk.Event),
	}
	dup := make(chan zk.Event, 2)
	const path = "/dubbo/config/"
	client.RegisterEvent(path, dup)
	client.RegisterEvent(path, dup)
	require.Len(t, client.eventRegistry[path], 1)

	session <- zk.Event{Type: zk.EventNodeDataChanged, State: zk.StateSyncConnected, Path: path + "demo"}
	close(session)
	(&DefaultHandler{}).HandleZkEvent(client)
	require.Len(t, dup, 1)

	client.UnregisterEvent(path, dup)
	_, exists := client.eventRegistry[path]
	require.False(t, exists)
}

func TestUnregisterEventRemovesDuplicateRegistration(t *testing.T) {
	// RegisterEvent no longer stores duplicates, so build the registry
	// directly to cover a slice that already holds the channel twice.
	dup := make(chan zk.Event, 1)
	other := make(chan zk.Event, 1)
	const path = "/dubbo/config/"
	client := &ZookeeperClient{
		eventRegistry: map[string][]chan zk.Event{path: {dup, other, dup}},
	}

	require.NotPanics(t, func() {
		client.UnregisterEvent(path, dup)
	})
	require.Equal(t, []chan zk.Event{other}, client.eventRegistry[path])
}

func TestDisconnectedClearsConnValid(t *testing.T) {
	session := make(chan zk.Event, 1)
	client := &ZookeeperClient{
		Session:       session,
		eventRegistry: make(map[string][]chan zk.Event),
	}
	atomic.StoreUint32(&client.valid, 1)
	session <- zk.Event{Type: zk.EventSession, State: zk.StateDisconnected}
	close(session)

	(&DefaultHandler{}).HandleZkEvent(client)

	require.False(t, client.ZkConnValid())
}
