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

/*
 * Portions of this file are derived from github.com/dubbogo/go-zookeeper.
 *
 * Copyright (c) 2013, Samuel Stauffer <samuel@descolada.com>
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice,
 *    this list of conditions and the following disclaimer.
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *    this list of conditions and the following disclaimer in the documentation
 *    and/or other materials provided with the distribution.
 * 3. Neither the name of the copyright holder nor the names of its contributors
 *    may be used to endorse or promote products derived from this software
 *    without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 * ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 * LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 * CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 * SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 * INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 * CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 * ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 * POSSIBILITY OF SUCH DAMAGE.
 */

package gxzookeeper

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

import (
	"github.com/dubbogo/go-zookeeper/zk"
)

// TestServer represents one server in a ZooKeeper test cluster.
type TestServer struct {
	Port int
	Path string
	Srv  *Server
}

// TestCluster represents a ZooKeeper test cluster.
type TestCluster struct {
	Path    string
	Servers []TestServer
}

type testClusterOptions struct {
	retryTimes int
}

type testClusterOption func(*testClusterOptions)

// WithRetryTimes sets the number of health-check attempts when starting a test cluster.
func WithRetryTimes(times int) testClusterOption {
	return func(options *testClusterOptions) {
		options.retryTimes = times
	}
}

// StartTestCluster starts a ZooKeeper test cluster with the requested number of servers.
func StartTestCluster(
	size int,
	stdout io.Writer,
	stderr io.Writer,
	opts ...testClusterOption,
) (*TestCluster, error) {
	tmpPath, err := os.MkdirTemp("", "gozk")
	if err != nil {
		return nil, err
	}

	success := false
	startPort := int(rand.Int31n(6000) + 10000)
	cluster := &TestCluster{Path: tmpPath}
	defer func() {
		if !success {
			_ = cluster.Stop()
		}
	}()

	options := &testClusterOptions{retryTimes: 10}
	for _, opt := range opts {
		opt(options)
	}

	for serverN := 0; serverN < size; serverN++ {
		srvPath := filepath.Join(tmpPath, fmt.Sprintf("srv%d", serverN))
		if err := os.Mkdir(srvPath, 0o700); err != nil {
			return nil, err
		}
		port := startPort + serverN*3

		// ZooKeeper treats backslashes as escapes in dataDir, so always write a slash-separated path.
		dataDir := filepath.ToSlash(srvPath)
		cfg := ServerConfig{
			ClientPort: port,
			DataDir:    dataDir,
		}

		for i := 0; i < size; i++ {
			cfg.Servers = append(cfg.Servers, ServerConfigServer{
				ID:                 i + 1,
				Host:               "127.0.0.1",
				PeerPort:           startPort + i*3 + 1,
				LeaderElectionPort: startPort + i*3 + 2,
			})
		}

		cfgPath := filepath.Join(srvPath, "zoo.cfg")
		cfgFile, err := os.Create(cfgPath)
		if err != nil {
			return nil, err
		}
		if err = cfg.Marshall(cfgFile); err != nil {
			_ = cfgFile.Close()
			return nil, err
		}
		if err = cfgFile.Close(); err != nil {
			return nil, err
		}

		myIDFile, err := os.Create(filepath.Join(srvPath, "myid"))
		if err != nil {
			return nil, err
		}
		if _, err = fmt.Fprintf(myIDFile, "%d\n", serverN+1); err != nil {
			_ = myIDFile.Close()
			return nil, err
		}
		if err = myIDFile.Close(); err != nil {
			return nil, err
		}

		srv := &Server{
			ConfigPath: cfgPath,
			Stdout:     stdout,
			Stderr:     stderr,
		}
		if err := srv.Start(); err != nil {
			return nil, err
		}
		cluster.Servers = append(cluster.Servers, TestServer{
			Path: srvPath,
			Port: cfg.ClientPort,
			Srv:  srv,
		})
	}

	if err := cluster.waitForStart(options.retryTimes, time.Second); err != nil {
		return nil, err
	}
	success = true
	return cluster, nil
}

// Connect connects to a server in the test cluster.
func (tc *TestCluster) Connect(index int) (*zk.Conn, error) {
	conn, _, err := zk.Connect([]string{fmt.Sprintf("127.0.0.1:%d", tc.Servers[index].Port)}, 15*time.Second)
	return conn, err
}

// ConnectAll connects to all servers in the test cluster with the default session timeout.
func (tc *TestCluster) ConnectAll() (*zk.Conn, <-chan zk.Event, error) {
	return tc.ConnectAllTimeout(15 * time.Second)
}

// ConnectAllTimeout connects to all servers with the specified session timeout.
func (tc *TestCluster) ConnectAllTimeout(sessionTimeout time.Duration) (*zk.Conn, <-chan zk.Event, error) {
	return tc.ConnectWithOptions(sessionTimeout)
}

// ConnectWithOptions connects to all servers with the specified session timeout.
func (tc *TestCluster) ConnectWithOptions(sessionTimeout time.Duration) (*zk.Conn, <-chan zk.Event, error) {
	hosts := make([]string, len(tc.Servers))
	for i, srv := range tc.Servers {
		hosts[i] = fmt.Sprintf("127.0.0.1:%d", srv.Port)
	}
	return zk.Connect(hosts, sessionTimeout)
}

// Stop stops all servers and removes the test cluster data directory.
func (tc *TestCluster) Stop() error {
	for _, srv := range tc.Servers {
		_ = srv.Srv.Stop()
	}
	defer func() {
		_ = os.RemoveAll(tc.Path)
	}()
	return tc.waitForStop(5, time.Second)
}

func (tc *TestCluster) waitForStart(maxRetry int, interval time.Duration) error {
	serverAddrs := make([]string, len(tc.Servers))
	for i, server := range tc.Servers {
		serverAddrs[i] = fmt.Sprintf("127.0.0.1:%d", server.Port)
	}

	for i := 0; i < maxRetry; i++ {
		_, ok := zk.FLWSrvr(serverAddrs, time.Second)
		if ok {
			return nil
		}
		time.Sleep(interval)
	}
	return fmt.Errorf("unable to verify health of servers")
}

func (tc *TestCluster) waitForStop(maxRetry int, interval time.Duration) error {
	serverAddrs := make([]string, len(tc.Servers))
	for i, server := range tc.Servers {
		serverAddrs[i] = fmt.Sprintf("127.0.0.1:%d", server.Port)
	}

	var success bool
	for i := 0; i < maxRetry && !success; i++ {
		success = true
		for _, ok := range zk.FLWRuok(serverAddrs, time.Second) {
			if ok {
				success = false
			}
		}
		if !success {
			time.Sleep(interval)
		}
	}
	if !success {
		return fmt.Errorf("unable to verify servers are down")
	}
	return nil
}

// StartServer starts one server selected by address.
func (tc *TestCluster) StartServer(server string) {
	for _, testServer := range tc.Servers {
		if strings.HasSuffix(server, fmt.Sprintf(":%d", testServer.Port)) {
			_ = testServer.Srv.Start()
			return
		}
	}
	panic(fmt.Sprintf("unknown server: %s", server))
}

// StopServer stops one server selected by address.
func (tc *TestCluster) StopServer(server string) {
	for _, testServer := range tc.Servers {
		if strings.HasSuffix(server, fmt.Sprintf(":%d", testServer.Port)) {
			_ = testServer.Srv.Stop()
			return
		}
	}
	panic(fmt.Sprintf("unknown server: %s", server))
}

// StartAllServers starts every server in the test cluster.
func (tc *TestCluster) StartAllServers() error {
	for _, server := range tc.Servers {
		if err := server.Srv.Start(); err != nil {
			return fmt.Errorf("failed to start server listening on port %d: %w", server.Port, err)
		}
	}
	return nil
}

// StopAllServers stops every server in the test cluster without removing its data directory.
func (tc *TestCluster) StopAllServers() error {
	for _, server := range tc.Servers {
		if err := server.Srv.Stop(); err != nil {
			return fmt.Errorf("failed to stop server listening on port %d: %w", server.Port, err)
		}
	}
	return nil
}
