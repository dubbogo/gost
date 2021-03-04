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
	"time"
)

import (
	"github.com/dubbogo/go-zookeeper/zk"
)

// nolint
type options struct {
	ZkName string
	Client *ZookeeperClient
	Ts     *zk.TestCluster
}

// Option will define a function of handling Options
type Option func(*options)

// WithZkName sets zk Client name
func WithZkName(name string) Option {
	return func(opt *options) {
		opt.ZkName = name
	}
}

type zkClientOption func(*ZookeeperClient)

// WithZkEventHandler sets zk Client event
func WithZkEventHandler(handler ZkEventHandler) zkClientOption {
	return func(opt *ZookeeperClient) {
		opt.zkEventHandler = handler
	}
}

// WithZkTimeOut sets zk Client timeout
func WithZkTimeOut(t time.Duration) zkClientOption {
	return func(opt *ZookeeperClient) {
		opt.Timeout = t
	}
}
