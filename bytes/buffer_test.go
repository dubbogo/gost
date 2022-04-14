// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

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
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestBufferWithPeek(t *testing.T) {
	var b Buffer
	b.WriteString("hello")

	b1 := b
	b1.WriteNextBegin(100)
	assert.True(t, b.off == b1.off)
	assert.True(t, b.lastRead == b1.lastRead)
	assert.True(t, len(b.buf) == len(b1.buf))
	assert.True(t, cap(b.buf) < cap(b1.buf))

	// out of range
	// l, err := b1.WriteNextEnd(101)
	// assert.Zero(t, l)
	// assert.NotNil(t, err)

	l, err := b1.WriteNextEnd(99)
	assert.Nil(t, err)
	assert.True(t, l == 99)
	assert.NotNil(t, b1.buf)
}
