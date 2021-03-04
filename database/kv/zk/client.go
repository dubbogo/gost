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
	"path"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

import (
	"github.com/dubbogo/go-zookeeper/zk"
	perrors "github.com/pkg/errors"
)

var (
	// ErrNilZkClientConn no conn error
	ErrNilZkClientConn = perrors.New("zookeeper Client{conn} is nil")
	// ErrNilChildren no children error
	ErrNilChildren = perrors.Errorf("has none children")
	// ErrNilNode no node error
	ErrNilNode = perrors.Errorf("node does not exist")
)

var (
	zkClientPool   zookeeperClientPool
	clientPoolOnce sync.Once
)

// ZookeeperClient represents zookeeper Client Configuration
type ZookeeperClient struct {
	name              string
	ZkAddrs           []string
	sync.RWMutex      // for conn
	Conn              *zk.Conn
	activeNumber      uint32
	Timeout           time.Duration
	Wait              sync.WaitGroup
	valid             uint32
	share             bool
	reconnectCh       chan struct{}
	eventRegistry     map[string][]*chan struct{}
	eventRegistryLock sync.RWMutex
	zkEventHandler    ZkEventHandler
	Session           <-chan zk.Event
}

type zookeeperClientPool struct {
	sync.Mutex
	zkClient map[string]*ZookeeperClient
}

// ZkEventHandler interface
type ZkEventHandler interface {
	HandleZkEvent(z *ZookeeperClient)
}

// DefaultHandler is default handler for zk event
type DefaultHandler struct {
}

// StateToString will transfer zk state to string
func StateToString(state zk.State) string {
	switch state {
	case zk.StateDisconnected:
		return "zookeeper disconnected"
	case zk.StateConnecting:
		return "zookeeper connecting"
	case zk.StateAuthFailed:
		return "zookeeper auth failed"
	case zk.StateConnectedReadOnly:
		return "zookeeper connect readonly"
	case zk.StateSaslAuthenticated:
		return "zookeeper sasl authenticated"
	case zk.StateExpired:
		return "zookeeper connection expired"
	case zk.StateConnected:
		return "zookeeper connected"
	case zk.StateHasSession:
		return "zookeeper has Session"
	case zk.StateUnknown:
		return "zookeeper unknown state"
	case zk.State(zk.EventNodeDeleted):
		return "zookeeper node deleted"
	case zk.State(zk.EventNodeDataChanged):
		return "zookeeper node data changed"
	default:
		return state.String()
	}
}

func initZookeeperClientPool() {
	zkClientPool.zkClient = make(map[string]*ZookeeperClient)
}

//NewZookeeperClient will create a ZookeeperClient
func NewZookeeperClient(name string, zkAddrs []string, share bool, opts ...zkClientOption) (*ZookeeperClient, error) {
	if share {
		clientPoolOnce.Do(initZookeeperClientPool)
		zkClientPool.Lock()
		defer zkClientPool.Unlock()
		if zkClient, ok := zkClientPool.zkClient[name]; ok {
			zkClient.activeNumber++
			return zkClient, nil
		}

	}
	newZkClient := &ZookeeperClient{
		name:           name,
		ZkAddrs:        zkAddrs,
		activeNumber:   0,
		share:          share,
		reconnectCh:    make(chan struct{}),
		eventRegistry:  make(map[string][]*chan struct{}),
		Session:        make(<-chan zk.Event),
		zkEventHandler: &DefaultHandler{},
	}
	for _, opt := range opts {
		opt(newZkClient)
	}
	err := newZkClient.createZookeeperConn()
	if err != nil {
		return nil, err
	}
	newZkClient.activeNumber++
	if share {
		zkClientPool.zkClient[name] = newZkClient
	}
	return newZkClient, nil
}

// nolint
func (z *ZookeeperClient) createZookeeperConn() error {
	var err error

	// connect to zookeeper
	z.Conn, z.Session, err = zk.Connect(z.ZkAddrs, z.Timeout)
	if err != nil {
		return err
	}
	atomic.StoreUint32(&z.valid, 1)
	go z.zkEventHandler.HandleZkEvent(z)
	return nil
}

// WithTestCluster sets test cluster for zk Client
func WithTestCluster(ts *zk.TestCluster) Option {
	return func(opt *options) {
		opt.Ts = ts
	}
}

// NewMockZookeeperClient returns a mock Client instance
func NewMockZookeeperClient(name string, timeout time.Duration, opts ...Option) (*zk.TestCluster, *ZookeeperClient, <-chan zk.Event, error) {
	var (
		err error
		z   *ZookeeperClient
		ts  *zk.TestCluster
	)

	z = &ZookeeperClient{
		name:           name,
		ZkAddrs:        []string{},
		Timeout:        timeout,
		share:          false,
		reconnectCh:    make(chan struct{}),
		eventRegistry:  make(map[string][]*chan struct{}),
		Session:        make(<-chan zk.Event),
		zkEventHandler: &DefaultHandler{},
	}

	option := &options{}
	for _, opt := range opts {
		opt(option)
	}

	// connect to zookeeper
	if option.Ts != nil {
		ts = option.Ts
	} else {
		ts, err = zk.StartTestCluster(1, nil, nil, zk.WithRetryTimes(40))
		if err != nil {
			return nil, nil, nil, perrors.WithMessagef(err, "zk.StartTestCluster fail")
		}
	}

	z.Conn, z.Session, err = ts.ConnectWithOptions(timeout)
	if err != nil {
		return nil, nil, nil, perrors.WithMessagef(err, "zk.Connect fail")
	}
	atomic.StoreUint32(&z.valid, 1)
	z.activeNumber++
	return ts, z, z.Session, nil
}

// HandleZkEvent handles zookeeper events
func (d *DefaultHandler) HandleZkEvent(z *ZookeeperClient) {
	var (
		state int
		event zk.Event
	)
	for {
		select {
		case event = <-z.Session:
			switch (int)(event.State) {
			case (int)(zk.StateDisconnected):
				atomic.StoreUint32(&z.valid, 0)
			case (int)(zk.EventNodeDataChanged), (int)(zk.EventNodeChildrenChanged):
				z.eventRegistryLock.RLock()
				for p, a := range z.eventRegistry {
					if strings.HasPrefix(p, event.Path) {
						for _, e := range a {
							*e <- struct{}{}
						}
					}
				}
				z.eventRegistryLock.RUnlock()
			case (int)(zk.StateConnecting), (int)(zk.StateConnected), (int)(zk.StateHasSession):
				if state == (int)(zk.StateHasSession) {
					continue
				}
				if event.State == zk.StateHasSession {
					atomic.StoreUint32(&z.valid, 1)
					close(z.reconnectCh)
					z.reconnectCh = make(chan struct{})
				}
				z.eventRegistryLock.RLock()
				if a, ok := z.eventRegistry[event.Path]; ok && 0 < len(a) {
					for _, e := range a {
						*e <- struct{}{}
					}
				}
				z.eventRegistryLock.RUnlock()
			}
			state = (int)(event.State)
		}
	}
}

// RegisterEvent registers zookeeper events
func (z *ZookeeperClient) RegisterEvent(zkPath string, event *chan struct{}) {
	if zkPath == "" || event == nil {
		return
	}

	z.eventRegistryLock.Lock()
	defer z.eventRegistryLock.Unlock()
	a := z.eventRegistry[zkPath]
	a = append(a, event)
	z.eventRegistry[zkPath] = a
}

// UnregisterEvent unregisters zookeeper events
func (z *ZookeeperClient) UnregisterEvent(zkPath string, event *chan struct{}) {
	if zkPath == "" {
		return
	}

	z.eventRegistryLock.Lock()
	defer z.eventRegistryLock.Unlock()
	infoList, ok := z.eventRegistry[zkPath]
	if !ok {
		return
	}
	for i, e := range infoList {
		if e == event {
			infoList = append(infoList[:i], infoList[i+1:]...)
		}
	}
	if len(infoList) == 0 {
		delete(z.eventRegistry, zkPath)
	} else {
		z.eventRegistry[zkPath] = infoList
	}
}

// ZkConnValid validates zookeeper connection
func (z *ZookeeperClient) ZkConnValid() bool {
	if atomic.LoadUint32(&z.valid) == 1 {
		return true
	}
	return false
}

// Create will create the node recursively, which means that if the parent node is absent,
// it will create parent node first.
// And the value for the basePath is ""
func (z *ZookeeperClient) Create(basePath string) error {
	return z.CreateWithValue(basePath, []byte(""))
}

// CreateWithValue will create the node recursively, which means that if the parent node is absent,
// it will create parent node first.
func (z *ZookeeperClient) CreateWithValue(basePath string, value []byte) error {
	var (
		err     error
		tmpPath string
	)

	conn := z.getConn()
	err = ErrNilZkClientConn
	if conn == nil {
		return perrors.WithMessagef(err, "zk.Create(path:%s)", basePath)
	}
	for _, str := range strings.Split(basePath, "/")[1:] {
		tmpPath = path.Join(tmpPath, "/", str)
		_, err = conn.Create(tmpPath, value, 0, zk.WorldACL(zk.PermAll))

		if err != nil {
			if err != zk.ErrNodeExists {
				return perrors.WithMessagef(err, "zk.Create(path:%s)", basePath)
			}
		}
	}

	return nil
}

// CreateTempWithValue will create the node recursively, which means that if the parent node is absent,
// it will create parent node first，and set value in last child path
// If the path exist, it will update data
func (z *ZookeeperClient) CreateTempWithValue(basePath string, value []byte) error {
	var (
		err     error
		tmpPath string
	)

	conn := z.getConn()
	err = ErrNilZkClientConn
	if conn == nil {
		return perrors.WithMessagef(err, "zk.Create(path:%s)", basePath)
	}

	pathSlice := strings.Split(basePath, "/")[1:]
	length := len(pathSlice)
	for i, str := range pathSlice {
		tmpPath = path.Join(tmpPath, "/", str)
		// last child need be ephemeral
		if i == length-1 {
			_, err = conn.Create(tmpPath, value, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
			if err == zk.ErrNodeExists {
				return err
			}
		} else {
			_, err = conn.Create(tmpPath, []byte{}, 0, zk.WorldACL(zk.PermAll))
		}
		if err != nil {
			if err != zk.ErrNodeExists {
				return perrors.WithMessagef(err, "zk.Create(path:%s)", basePath)
			}
		}
	}

	return nil
}

//Delete will delete basePath
func (z *ZookeeperClient) Delete(basePath string) error {
	err := ErrNilZkClientConn
	conn := z.getConn()
	if conn != nil {
		err = conn.Delete(basePath, -1)
	}
	return perrors.WithMessagef(err, "Delete(basePath:%s)", basePath)
}

// RegisterTemp registers temporary node by @basePath and @node
func (z *ZookeeperClient) RegisterTemp(basePath string, node string) (string, error) {
	var (
		err     error
		zkPath  string
		tmpPath string
	)

	err = ErrNilZkClientConn
	zkPath = path.Join(basePath) + "/" + node
	conn := z.getConn()
	if conn != nil {
		tmpPath, err = conn.Create(zkPath, []byte(""), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	}

	if err != nil {
		return zkPath, perrors.WithStack(err)
	}

	return tmpPath, nil
}

// RegisterTempSeq register temporary sequence node by @basePath and @data
func (z *ZookeeperClient) RegisterTempSeq(basePath string, data []byte) (string, error) {
	var (
		err     error
		tmpPath string
	)

	err = ErrNilZkClientConn
	conn := z.getConn()
	if conn != nil {
		tmpPath, err = conn.Create(
			path.Join(basePath)+"/",
			data,
			zk.FlagEphemeral|zk.FlagSequence,
			zk.WorldACL(zk.PermAll),
		)
	}

	if err != nil && err != zk.ErrNodeExists {
		return "", perrors.WithStack(err)
	}
	return tmpPath, nil
}

// GetChildrenW gets children watch by @path
func (z *ZookeeperClient) GetChildrenW(path string) ([]string, <-chan zk.Event, error) {
	var (
		err      error
		children []string
		stat     *zk.Stat
		watcher  *zk.Watcher
	)

	err = ErrNilZkClientConn
	conn := z.getConn()
	if conn != nil {
		children, stat, watcher, err = conn.ChildrenW(path)
	}

	if err != nil {
		if err == zk.ErrNoChildrenForEphemerals {
			return nil, nil, ErrNilChildren
		}
		if err == zk.ErrNoNode {
			return nil, nil, ErrNilNode
		}
		return nil, nil, perrors.WithMessagef(err, "zk.ChildrenW(path:%s)", path)
	}
	if stat == nil {
		return nil, nil, perrors.Errorf("path{%s} get stat is nil", path)
	}
	if len(children) == 0 {
		return nil, nil, ErrNilChildren
	}

	return children, watcher.EvtCh, nil
}

// GetChildren gets children by @path
func (z *ZookeeperClient) GetChildren(path string) ([]string, error) {
	var (
		err      error
		children []string
		stat     *zk.Stat
	)

	err = ErrNilZkClientConn
	conn := z.getConn()
	if conn != nil {
		children, stat, err = conn.Children(path)
	}

	if err != nil {
		if err == zk.ErrNoNode {
			return nil, perrors.Errorf("path{%s} has none children", path)
		}
		return nil, perrors.WithMessagef(err, "zk.Children(path:%s)", path)
	}
	if stat == nil {
		return nil, perrors.Errorf("path{%s} has none children", path)
	}
	if len(children) == 0 {
		return nil, ErrNilChildren
	}

	return children, nil
}

// ExistW to judge watch whether it exists or not by @zkPath
func (z *ZookeeperClient) ExistW(zkPath string) (<-chan zk.Event, error) {
	var (
		exist   bool
		err     error
		watcher *zk.Watcher
	)

	err = ErrNilZkClientConn
	conn := z.getConn()
	if conn != nil {
		exist, _, watcher, err = conn.ExistsW(zkPath)
	}

	if err != nil {
		return nil, perrors.WithMessagef(err, "zk.ExistsW(path:%s)", zkPath)
	}
	if !exist {
		return nil, perrors.Errorf("zkClient{%s} App zk path{%s} does not exist.", z.name, zkPath)
	}

	return watcher.EvtCh, nil
}

// GetContent gets content by @zkPath
func (z *ZookeeperClient) GetContent(zkPath string) ([]byte, *zk.Stat, error) {
	return z.Conn.Get(zkPath)
}

//SetContent set content of zkPath
func (z *ZookeeperClient) SetContent(zkPath string, content []byte, version int32) (*zk.Stat, error) {
	return z.Conn.Set(zkPath, content, version)
}

// getConn gets zookeeper connection safely
func (z *ZookeeperClient) getConn() *zk.Conn {
	if z == nil {
		return nil
	}
	z.RLock()
	defer z.RUnlock()
	return z.Conn
}

// Reconnect gets zookeeper reconnect event
func (z *ZookeeperClient) Reconnect() <-chan struct{} {
	return z.reconnectCh
}

// GetEventHandler gets zookeeper event handler
func (z *ZookeeperClient) GetEventHandler() ZkEventHandler {
	return z.zkEventHandler
}

func (z *ZookeeperClient) Close() {
	if z.share {
		zkClientPool.Lock()
		defer zkClientPool.Unlock()
		z.activeNumber--
		if z.activeNumber == 0 {
			z.Conn.Close()
			delete(zkClientPool.zkClient, z.name)
		}
	} else {
		z.Lock()
		conn := z.Conn
		z.activeNumber--
		z.Conn = nil
		z.Unlock()
		if conn != nil {
			conn.Close()
		}
	}

}
