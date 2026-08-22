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

package gxetcd

import (
	"context"
	"net"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"
	"testing"
	"time"
)

import (
	perrors "github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"

	"go.etcd.io/etcd/server/v3/embed"

	"google.golang.org/grpc/connectivity"
)

const defaultEtcdV3WorkDir = "/tmp/default-dubbo-go-remote.etcd"

// tests dataset
var tests = []struct {
	input struct {
		k string
		v string
	}
}{
	{input: struct {
		k string
		v string
	}{k: "name/name", v: "scott.wang"}},
	{input: struct {
		k string
		v string
	}{k: "name/namePrefix", v: "prefix.scott.wang"}},
	{input: struct {
		k string
		v string
	}{k: "name/namePrefix1", v: "prefix1.scott.wang"}},
	{input: struct {
		k string
		v string
	}{k: "age", v: "27"}},
}

// test dataset prefix
const prefixKey = "name/"
const keyPrefix = "name/name"

type ClientTestSuite struct {
	suite.Suite

	etcdConfig struct {
		name      string
		endpoints []string
		timeout   time.Duration
		heartbeat int
	}

	etcd *embed.Etcd

	client *Client
}

// start etcd server
func (suite *ClientTestSuite) SetupSuite() {
	t := suite.T()

	DefaultListenPeerURLs := "http://localhost:2382"
	DefaultListenClientURLs := "http://localhost:2381"
	lpurl, _ := url.Parse(DefaultListenPeerURLs)
	lcurl, _ := url.Parse(DefaultListenClientURLs)
	cfg := embed.NewConfig()
	cfg.LPUrls = []url.URL{*lpurl}
	cfg.LCUrls = []url.URL{*lcurl}
	cfg.Dir = defaultEtcdV3WorkDir
	e, err := embed.StartEtcd(cfg)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-e.Server.ReadyNotify():
		t.Log("Server is ready!")
	case <-time.After(60 * time.Second):
		e.Server.Stop() // trigger a shutdown
		t.Logf("Server took too long to start!")
	}

	suite.etcd = e
}

// stop etcd server
func (suite *ClientTestSuite) TearDownSuite() {
	suite.etcd.Close()
	if err := os.RemoveAll(defaultEtcdV3WorkDir); err != nil {
		suite.FailNow(err.Error())
	}
}

func (suite *ClientTestSuite) setUpClient() *Client {
	c, err := NewConfigClientWithErr(WithName(suite.etcdConfig.name),
		WithEndpoints(suite.etcdConfig.endpoints...),
		WithTimeout(suite.etcdConfig.timeout),
		WithHeartbeat(suite.etcdConfig.heartbeat))
	if err != nil {
		suite.T().Fatal(err)
	}
	return c
}

// set up a client for suite
func (suite *ClientTestSuite) SetupTest() {
	c := suite.setUpClient()
	if err := c.CleanKV(); err != nil {
		suite.T().Fatal("CleanKV failed:", err)
	}
	suite.client = c
}

func (suite *ClientTestSuite) TestClientClose() {
	c := suite.client
	t := suite.T()

	defer c.Close()
	if c.rawClient.ActiveConnection().GetState() != connectivity.Ready {
		t.Fatal(suite.client.rawClient.ActiveConnection().GetState())
	}
}

func TestNewClientWithOptionsTimeoutsInitialGrant(t *testing.T) {
	const timeout = 100 * time.Millisecond

	embed.SetupGrpcLoggingForTest(false)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	listenerDone := make(chan struct{})
	go func() {
		defer close(listenerDone)
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		defer conn.Close()
		<-ctx.Done()
	}()
	t.Cleanup(func() {
		cancel()
		_ = listener.Close()
		<-listenerDone
	})

	started := time.Now()
	done := make(chan error, 1)
	go func() {
		_, clientErr := NewClientWithOptions(ctx, &Options{
			Endpoints: []string{listener.Addr().String()},
			Timeout:   timeout,
			Heartbeat: 1,
		})
		done <- clientErr
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected client initialization to fail")
		}
		if !strings.Contains(err.Error(), "grant session lease") {
			t.Fatalf("expected grant session lease error, got %v", err)
		}
		if elapsed := time.Since(started); elapsed > time.Second {
			t.Fatalf("client initialization took too long: %s", elapsed)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("client initialization did not respect the grant timeout")
	}
}

func (suite *ClientTestSuite) TestKeepSessionUsesDefaultTTL() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	before, err := suite.client.rawClient.Leases(ctx)
	suite.Require().NoError(err)
	existingLeases := make(map[clientv3.LeaseID]struct{}, len(before.Leases))
	for _, lease := range before.Leases {
		existingLeases[lease.ID] = struct{}{}
	}

	c, err := NewConfigClientWithErr(
		WithName(suite.etcdConfig.name),
		WithEndpoints(suite.etcdConfig.endpoints...),
		WithTimeout(suite.etcdConfig.timeout),
		WithHeartbeat(0),
	)
	suite.Require().NoError(err)
	defer c.Close()

	after, err := c.rawClient.Leases(ctx)
	suite.Require().NoError(err)
	newLeases := make([]clientv3.LeaseID, 0, 1)
	for _, lease := range after.Leases {
		if _, ok := existingLeases[lease.ID]; !ok {
			newLeases = append(newLeases, lease.ID)
		}
	}
	suite.Require().Len(newLeases, 1)

	ttl, err := c.rawClient.TimeToLive(ctx, newLeases[0])
	suite.Require().NoError(err)
	suite.Equal(int64(defaultSessionTTL), ttl.GrantedTTL)

	_, err = c.rawClient.Revoke(ctx, newLeases[0])
	suite.Require().NoError(err)
}

func (suite *ClientTestSuite) TestClientValid() {
	c := suite.client
	t := suite.T()

	if !c.Valid() {
		t.Fatal("client is not valid")
	}
	c.Close()
	if suite.client.Valid() != false {
		t.Fatal("client is valid")
	}
}

func (suite *ClientTestSuite) TestClientDone() {
	c := suite.client

	go func() {
		time.Sleep(2 * time.Second)
		c.Close()
	}()

	c.Wait.Wait()

	if c.Valid() {
		suite.T().Fatal("client should be invalid then")
	}
}

func (suite *ClientTestSuite) TestClientCreateKV() {
	tests := tests

	c := suite.client
	t := suite.T()

	defer suite.client.Close()

	for _, tc := range tests {

		k := tc.input.k
		v := tc.input.v
		expect := tc.input.v

		if err := c.Create(k, v); err != nil {
			t.Fatal(err)
		}

		value, err := c.Get(k)
		if err != nil {
			t.Fatal(err)
		}

		if value != expect {
			t.Fatalf("expect %v but get %v", expect, value)
		}

	}
}

func (suite *ClientTestSuite) TestBatchClientCreateKV() {
	tests := tests

	c := suite.client
	t := suite.T()

	defer suite.client.Close()

	for _, tc := range tests {

		k := tc.input.k
		v := tc.input.v
		expect := tc.input.v
		kList := make([]string, 0, 1)
		vList := make([]string, 0, 1)
		kList = append(kList, k)
		vList = append(vList, v)

		if err := c.BatchCreate(kList, vList); err != nil {
			t.Fatal(err)
		}

		value, err := c.Get(k)
		if err != nil {
			t.Fatal(err)
		}

		if value != expect {
			t.Fatalf("expect %v but get %v", expect, value)
		}
	}
}

func (suite *ClientTestSuite) TestBatchClientGetValAndRevKV() {
	tests := tests

	c := suite.client
	t := suite.T()

	defer suite.client.Close()

	for _, tc := range tests {

		k := tc.input.k
		v := tc.input.v
		expect := tc.input.v
		kList := make([]string, 0, 1)
		vList := make([]string, 0, 1)
		kList = append(kList, k)
		vList = append(vList, v)

		if err := c.BatchCreate(kList, vList); err != nil {
			t.Fatal(err)
		}

		value, revision, err := c.getValAndRev(k)
		if err != nil {
			t.Fatal(err)
		}

		err = c.UpdateWithRev(k, k, revision)
		if err != nil {
			t.Fatal(err)
		}

		err = c.Update(k, k)
		if err != nil {
			t.Fatal(err)
		}

		if value != expect {
			t.Fatalf("expect %v but get %v", expect, value)
		}
	}
}

func (suite *ClientTestSuite) TestClientDeleteKV() {
	tests := tests
	c := suite.client
	t := suite.T()

	defer c.Close()

	for _, tc := range tests {

		k := tc.input.k
		v := tc.input.v
		expect := ErrKVPairNotFound

		if err := c.Put(k, v); err != nil {
			t.Fatal(err)
		}

		if err := c.Delete(k); err != nil {
			t.Fatal(err)
		}

		_, err := c.Get(k)
		if perrors.Cause(err) == expect {
			continue
		}

		if err != nil {
			t.Fatal(err)
		}
	}
}

func (suite *ClientTestSuite) TestClientGetChildrenKVList() {
	tests := tests

	c := suite.client
	t := suite.T()

	var expectKList []string
	var expectVList []string

	for _, tc := range tests {

		k := tc.input.k
		v := tc.input.v

		if strings.Contains(k, prefixKey) {
			expectKList = append(expectKList, k)
			expectVList = append(expectVList, v)
		}

		if err := c.Create(k, v); err != nil {
			t.Fatal(err)
		}
	}

	kList, vList, err := c.GetChildrenKVList(prefixKey)
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(expectKList, kList) && reflect.DeepEqual(expectVList, vList) {
		return
	}

	t.Fatalf("expect keylist %v but got %v expect valueList %v but got %v ", expectKList, kList, expectVList, vList)
}

func (suite *ClientTestSuite) TestClientWatch() {
	tests := tests

	c := suite.client
	t := suite.T()

	go func() {
		time.Sleep(time.Second)
		for _, tc := range tests {

			k := tc.input.k
			v := tc.input.v

			if err := c.Create(k, v); err != nil {
				assert.Error(t, err)
			}

			if err := c.delete(k); err != nil {
				assert.Error(t, err)
			}
		}

		c.Close()
	}()

	wc, err := c.WatchWithOption(keyPrefix)
	if err != nil {
		assert.Error(t, err)
	}

	events := make([]mvccpb.Event, 0)
	var eCreate, eDelete mvccpb.Event

	for e := range wc {
		for _, event := range e.Events {
			events = append(events, (mvccpb.Event)(*event))
			if event.Type == mvccpb.PUT {
				eCreate = (mvccpb.Event)(*event)
			}
			if event.Type == mvccpb.DELETE {
				eDelete = (mvccpb.Event)(*event)
			}
			t.Logf("type IsCreate %v k %s v %s", event.IsCreate(), event.Kv.Key, event.Kv.Value)
		}
	}

	assert.Equal(t, 2, len(events))
	assert.Contains(t, events, eCreate)
	assert.Contains(t, events, eDelete)
}

func (suite *ClientTestSuite) TestClientRegisterTemp() {
	c := suite.client
	observeC := suite.setUpClient()
	t := suite.T()

	go func() {
		time.Sleep(2 * time.Second)
		err := c.RegisterTemp("scott/wang", "test")
		if err != nil {
			assert.Error(t, err)
		}
		c.Close()
	}()

	completePath := path.Join("scott", "wang")
	wc, err := observeC.watchWithOption(completePath)
	if err != nil {
		assert.Error(t, err)
	}

	events := make([]mvccpb.Event, 0)
	var eCreate, eDelete mvccpb.Event

	for e := range wc {
		for _, event := range e.Events {
			events = append(events, (mvccpb.Event)(*event))
			if event.Type == mvccpb.DELETE {
				eDelete = (mvccpb.Event)(*event)
				t.Logf("complete key (%s) is delete", completePath)
				observeC.Close()
				break
			}
			eCreate = (mvccpb.Event)(*event)
			t.Logf("type IsCreate %v k %s v %s", event.IsCreate(), event.Kv.Key, event.Kv.Value)
		}
	}

	assert.Equal(t, 2, len(events))
	assert.Contains(t, events, eCreate)
	assert.Contains(t, events, eDelete)
}

func TestClientSuite(t *testing.T) {
	suite.Run(t, &ClientTestSuite{
		etcdConfig: struct {
			name      string
			endpoints []string
			timeout   time.Duration
			heartbeat int
		}{
			name:      "test",
			endpoints: []string{"localhost:2381"},
			timeout:   time.Second,
			heartbeat: 1,
		},
	})
}
