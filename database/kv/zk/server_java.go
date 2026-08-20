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
	"os"
	"os/exec"
	"path/filepath"
)

import (
	"github.com/dubbogo/go-zookeeper/zk"
)

// ErrMissingServerConfigField indicates a required ZooKeeper server configuration field is missing.
type ErrMissingServerConfigField string

func (e ErrMissingServerConfigField) Error() string {
	return fmt.Sprintf("zk: missing server config field '%s'", string(e))
}

const (
	DefaultServerTickTime                 = 2000
	DefaultServerInitLimit                = 10
	DefaultServerSyncLimit                = 5
	DefaultServerAutoPurgeSnapRetainCount = 3
	DefaultPeerPort                       = 2888
	DefaultLeaderElectionPort             = 3888
)

// ServerConfigServer contains the peer configuration for one ZooKeeper server.
type ServerConfigServer struct {
	ID                 int
	Host               string
	PeerPort           int
	LeaderElectionPort int
}

// ServerConfig contains the configuration written to a ZooKeeper server config file.
type ServerConfig struct {
	TickTime                 int
	InitLimit                int
	SyncLimit                int
	DataDir                  string
	ClientPort               int
	AutoPurgeSnapRetainCount int
	AutoPurgePurgeInterval   int
	Servers                  []ServerConfigServer
}

// Marshall writes the ZooKeeper server configuration.
func (sc ServerConfig) Marshall(w io.Writer) error {
	if sc.DataDir == "" {
		return ErrMissingServerConfigField("dataDir")
	}
	if _, err := fmt.Fprintf(w, "dataDir=%s\n", sc.DataDir); err != nil {
		return err
	}
	if sc.TickTime <= 0 {
		sc.TickTime = DefaultServerTickTime
	}
	if _, err := fmt.Fprintf(w, "tickTime=%d\n", sc.TickTime); err != nil {
		return err
	}
	if sc.InitLimit <= 0 {
		sc.InitLimit = DefaultServerInitLimit
	}
	if _, err := fmt.Fprintf(w, "initLimit=%d\n", sc.InitLimit); err != nil {
		return err
	}
	if sc.SyncLimit <= 0 {
		sc.SyncLimit = DefaultServerSyncLimit
	}
	if _, err := fmt.Fprintf(w, "syncLimit=%d\n", sc.SyncLimit); err != nil {
		return err
	}
	if sc.ClientPort <= 0 {
		sc.ClientPort = zk.DefaultPort
	}
	if _, err := fmt.Fprintf(w, "clientPort=%d\n", sc.ClientPort); err != nil {
		return err
	}
	if sc.AutoPurgePurgeInterval > 0 {
		if sc.AutoPurgeSnapRetainCount <= 0 {
			sc.AutoPurgeSnapRetainCount = DefaultServerAutoPurgeSnapRetainCount
		}
		if _, err := fmt.Fprintf(w, "autopurge.snapRetainCount=%d\n", sc.AutoPurgeSnapRetainCount); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "autopurge.purgeInterval=%d\n", sc.AutoPurgePurgeInterval); err != nil {
			return err
		}
	}
	for _, srv := range sc.Servers {
		if srv.PeerPort <= 0 {
			srv.PeerPort = DefaultPeerPort
		}
		if srv.LeaderElectionPort <= 0 {
			srv.LeaderElectionPort = DefaultLeaderElectionPort
		}
		if _, err := fmt.Fprintf(
			w,
			"server.%d=%s:%d:%d\n",
			srv.ID,
			srv.Host,
			srv.PeerPort,
			srv.LeaderElectionPort,
		); err != nil {
			return err
		}
	}
	return nil
}

var jarSearchPaths = []string{
	"zookeeper-*/contrib/fatjar/zookeeper-*-fatjar.jar",
	"../zookeeper-*/contrib/fatjar/zookeeper-*-fatjar.jar",
	"/usr/share/java/zookeeper-*.jar",
	"/usr/local/zookeeper-*/contrib/fatjar/zookeeper-*-fatjar.jar",
	"/usr/local/Cellar/zookeeper/*/libexec/contrib/fatjar/zookeeper-*-fatjar.jar",
}

func findZookeeperFatJar() string {
	var paths []string
	zkPath := os.Getenv("ZOOKEEPER_PATH")
	if zkPath == "" {
		paths = jarSearchPaths
	} else {
		paths = []string{filepath.Join(zkPath, "contrib/fatjar/zookeeper-*-fatjar.jar")}
	}
	for _, path := range paths {
		matches, _ := filepath.Glob(path)
		if len(matches) > 0 {
			return matches[0]
		}
	}
	return ""
}

// Server manages one ZooKeeper Java server process.
type Server struct {
	JarPath        string
	ConfigPath     string
	Stdout, Stderr io.Writer

	cmd *exec.Cmd
}

// Start starts the ZooKeeper Java server process.
func (srv *Server) Start() error {
	if srv.JarPath == "" {
		srv.JarPath = findZookeeperFatJar()
		if srv.JarPath == "" {
			return fmt.Errorf("zk: unable to find server jar")
		}
	}
	srv.cmd = exec.Command("java", "-jar", srv.JarPath, "server", srv.ConfigPath)
	srv.cmd.Stdout = srv.Stdout
	srv.cmd.Stderr = srv.Stderr
	return srv.cmd.Start()
}

// Stop stops the ZooKeeper Java server process.
func (srv *Server) Stop() error {
	if err := srv.cmd.Process.Signal(os.Kill); err != nil {
		return err
	}
	return srv.cmd.Wait()
}
