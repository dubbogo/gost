package gxzookeeper

import (
	"sync"
	"time"
)

import (
	"github.com/dubbogo/go-zookeeper/zk"
)

// nolint
type Options struct {
	zkName string
	client *ZookeeperClient
	ts     *zk.TestCluster
}

// Option will define a function of handling Options
type Option func(*Options)

// WithZkName sets zk client name
func WithZkName(name string) Option {
	return func(opt *Options) {
		opt.zkName = name
	}
}

// ZookeeperClient represents zookeeper client Configuration
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

type ZkClientOption func(*ZookeeperClient)

// WithZkEventHandler sets zk client event
func WithZkEventHandler(handler ZkEventHandler) ZkClientOption {
	return func(opt *ZookeeperClient) {
		opt.zkEventHandler = handler
	}
}

// WithZkEventHandler sets zk client timeout
func WithZkTimeOut(t time.Duration) ZkClientOption {
	return func(opt *ZookeeperClient) {
		opt.Timeout = t
	}
}
