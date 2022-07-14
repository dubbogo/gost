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

const (
	SLASH = "/"
)

var (
	zkClientPool   zookeeperClientPool
	clientPoolOnce sync.Once

	// ErrNilZkClientConn no conn error
	ErrNilZkClientConn = perrors.New("Zookeeper Client{conn} is nil")
	ErrStatIsNil       = perrors.New("Stat of the node is nil")
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
	initialized       uint32
	reconnectCh       chan struct{}
	eventRegistry     map[string][]chan zk.Event
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
type DefaultHandler struct{}

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
	default:
		return state.String()
	}
}

func initZookeeperClientPool() {
	zkClientPool.zkClient = make(map[string]*ZookeeperClient)
}

// NewZookeeperClient will create a ZookeeperClient
func NewZookeeperClient(name string, zkAddrs []string, share bool, opts ...zkClientOption) (*ZookeeperClient, error) {
	if !share {
		return newClient(name, zkAddrs, share, opts...)
	}
	clientPoolOnce.Do(initZookeeperClientPool)
	zkClientPool.Lock()
	defer zkClientPool.Unlock()
	if zkClient, ok := zkClientPool.zkClient[name]; ok {
		zkClient.activeNumber++
		return zkClient, nil
	}
	newZkClient, err := newClient(name, zkAddrs, share, opts...)
	if err != nil {
		return nil, err
	}
	zkClientPool.zkClient[name] = newZkClient
	return newZkClient, nil
}

func newClient(name string, zkAddrs []string, share bool, opts ...zkClientOption) (*ZookeeperClient, error) {
	newZkClient := &ZookeeperClient{
		name:           name,
		ZkAddrs:        zkAddrs,
		activeNumber:   0,
		share:          share,
		reconnectCh:    make(chan struct{}),
		eventRegistry:  make(map[string][]chan zk.Event),
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
		eventRegistry:  make(map[string][]chan zk.Event),
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
		ok    bool
		state int
		event zk.Event
	)
	for {
		select {
		case event, ok = <-z.Session:
			if !ok {
				// channel already closed
				return
			}
			switch event.State {
			case zk.StateDisconnected:
				atomic.StoreUint32(&z.valid, 0)
			case zk.StateConnected:
				z.eventRegistryLock.RLock()
				for path, a := range z.eventRegistry {
					if strings.HasPrefix(event.Path, path) {
						for _, e := range a {
							e <- event
						}
					}
				}
				z.eventRegistryLock.RUnlock()
			case zk.StateConnecting, zk.StateHasSession:
				if state == (int)(zk.StateHasSession) {
					continue
				}
				if event.State == zk.StateHasSession {
					atomic.StoreUint32(&z.valid, 1)
					//if this is the first connection, don't trigger reconnect event
					if !atomic.CompareAndSwapUint32(&z.initialized, 0, 1) {
						close(z.reconnectCh)
						z.reconnectCh = make(chan struct{})
					}
				}
				z.eventRegistryLock.RLock()
				if a, ok := z.eventRegistry[event.Path]; ok && 0 < len(a) {
					for _, e := range a {
						e <- event
					}
				}
				z.eventRegistryLock.RUnlock()
			}
			state = (int)(event.State)
		}
	}
}

// RegisterEvent registers zookeeper events
func (z *ZookeeperClient) RegisterEvent(zkPath string, event chan zk.Event) {
	if zkPath == "" {
		return
	}

	z.eventRegistryLock.Lock()
	defer z.eventRegistryLock.Unlock()
	a := z.eventRegistry[zkPath]
	a = append(a, event)
	z.eventRegistry[zkPath] = a
}

// UnregisterEvent unregisters zookeeper events
func (z *ZookeeperClient) UnregisterEvent(zkPath string, event chan zk.Event) {
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
	return z.CreateWithValue(basePath, []byte{})
}

// CreateWithValue will create the node recursively, which means that if the parent node is absent,
// it will create parent node first.
// basePath should start with "/"
func (z *ZookeeperClient) CreateWithValue(basePath string, value []byte) error {
	conn := z.getConn()
	if conn == nil {
		return ErrNilZkClientConn
	}

	if !strings.HasPrefix(basePath, SLASH) {
		basePath = SLASH + basePath
	}
	paths := strings.Split(basePath, SLASH)
	// Check the ancestor's path
	for idx := 2; idx < len(paths); idx++ {
		tmpPath := strings.Join(paths[:idx], SLASH)
		_, err := conn.Create(tmpPath, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil && err != zk.ErrNodeExists {
			return perrors.WithMessagef(err, "Error while invoking zk.Create(path:%s), the reason maybe is: ", tmpPath)
		}
	}

	_, err := conn.Create(basePath, value, 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		return err
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
	if conn == nil {
		return ErrNilZkClientConn
	}

	if !strings.HasPrefix(basePath, SLASH) {
		basePath = SLASH + basePath
	}
	pathSlice := strings.Split(basePath, SLASH)[1:]
	length := len(pathSlice)
	for i, str := range pathSlice {
		tmpPath = path.Join(tmpPath, SLASH, str)
		// last child need be ephemeral
		if i == length-1 {
			_, err = conn.Create(tmpPath, value, zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
			if err != nil {
				return perrors.WithMessagef(err, "Error while invoking zk.Create(path:%s), the reason maybe is: ", tmpPath)
			}
			break
		}
		// we need ignore node exists error for those parent node
		_, err = conn.Create(tmpPath, []byte{}, 0, zk.WorldACL(zk.PermAll))
		if err != nil && err != zk.ErrNodeExists {
			return perrors.WithMessagef(err, "Error while invoking zk.Create(path:%s), the reason maybe is: ", tmpPath)
		}
	}

	return nil
}

// Delete will delete basePath
func (z *ZookeeperClient) Delete(basePath string) error {
	conn := z.getConn()
	if conn == nil {
		return ErrNilZkClientConn
	}
	return perrors.WithMessagef(conn.Delete(basePath, -1), "Delete(basePath:%s)", basePath)
}

// RegisterTemp registers temporary node by @basePath and @node
func (z *ZookeeperClient) RegisterTemp(basePath string, node string) (string, error) {
	zkPath := path.Join(basePath) + SLASH + node
	conn := z.getConn()
	if conn == nil {
		return "", ErrNilZkClientConn
	}
	tmpPath, err := conn.Create(zkPath, []byte(""), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))

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
			path.Join(basePath)+SLASH,
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
	conn := z.getConn()
	if conn == nil {
		return nil, nil, ErrNilZkClientConn
	}
	children, stat, watcher, err := conn.ChildrenW(path)

	if err != nil {
		return nil, nil, perrors.WithMessagef(err, "Error while invoking zk.ChildrenW(path:%s), the reason maybe is: ", path)
	}
	if stat == nil {
		return nil, nil, perrors.WithMessagef(ErrStatIsNil, "Error while invokeing zk.ChildrenW(path:%s), the reason is: ", path)
	}

	return children, watcher.EvtCh, nil
}

// GetChildren gets children by @path
func (z *ZookeeperClient) GetChildren(path string) ([]string, error) {
	conn := z.getConn()
	if conn == nil {
		return nil, ErrNilZkClientConn
	}
	children, stat, err := conn.Children(path)

	if err != nil {
		return nil, perrors.WithMessagef(err, "Error while invoking zk.Children(path:%s), the reason maybe is: ", path)
	}
	if stat == nil {
		return nil, perrors.Errorf("Error while invokeing zk.Children(path:%s), the reason is that the stat is nil", path)
	}

	return children, nil
}

// ExistW to judge watch whether it exists or not by @zkPath
func (z *ZookeeperClient) ExistW(zkPath string) (<-chan zk.Event, error) {
	conn := z.getConn()
	if conn == nil {
		return nil, ErrNilZkClientConn
	}
	_, _, watcher, err := conn.ExistsW(zkPath)

	if err != nil {
		return nil, perrors.WithMessagef(err, "zk.ExistsW(path:%s)", zkPath)
	}

	return watcher.EvtCh, nil
}

// GetContent gets content by @zkPath
func (z *ZookeeperClient) GetContent(zkPath string) ([]byte, *zk.Stat, error) {
	return z.Conn.Get(zkPath)
}

// SetContent set content of zkPath
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
