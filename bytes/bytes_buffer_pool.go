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

package gxbytes

import (
	"bytes"
	"sync"
)

var (
	defaultPool BytesBufferPool
)

// GetBytesBuffer returns bytes.Buffer from pool
func GetBytesBuffer() *bytes.Buffer {
	return defaultPool.Get()
}

// PutIoBuffer returns IoBuffer to pool
func PutBytesBuffer(buf *bytes.Buffer) {
	defaultPool.Put(buf)
}

// BytesBufferPool is bytes.Buffer Pool
type BytesBufferPool struct {
	pool sync.Pool
}

// take returns *bytes.Buffer from Pool
func (p *BytesBufferPool) Get() *bytes.Buffer {
	v := p.pool.Get()
	if v == nil {
		return new(bytes.Buffer)
	}

	return v.(*bytes.Buffer)
}

// give returns *byes.Buffer to Pool
func (p *BytesBufferPool) Put(buf *bytes.Buffer) {
	buf.Reset()
	p.pool.Put(buf)
}
