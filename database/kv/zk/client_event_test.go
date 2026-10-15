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
	"testing"

	"github.com/go-zookeeper/zk"
	"github.com/stretchr/testify/require"
)

func TestDefaultHandlerDispatchesConnectedNodeEvents(t *testing.T) {
	for _, state := range []zk.State{zk.StateSyncConnected, zk.StateConnected} {
		for _, kind := range []zk.EventType{zk.EventNodeCreated, zk.EventNodeDataChanged, zk.EventNodeDeleted} {
			t.Run(state.String()+"/"+kind.String(), func(t *testing.T) {
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
			})
		}
	}
}
