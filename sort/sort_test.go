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

package gxsort

import (
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestSortInt32(t *testing.T) {
	data := []int32{3, 5, 1, 9, 0, 2, 2}
	Int32(data)
	assert.Equal(t, int32(0), data[0])
	assert.Equal(t, int32(1), data[1])
	assert.Equal(t, int32(2), data[2])
	assert.Equal(t, int32(2), data[3])
	assert.Equal(t, int32(3), data[4])
	assert.Equal(t, int32(5), data[5])
	assert.Equal(t, int32(9), data[6])
}

func TestSortInt64(t *testing.T) {
	data := []int64{3, 5, 1, 9, 0, 2, 2}
	Int64(data)
	assert.Equal(t, int64(0), data[0])
	assert.Equal(t, int64(1), data[1])
	assert.Equal(t, int64(2), data[2])
	assert.Equal(t, int64(2), data[3])
	assert.Equal(t, int64(3), data[4])
	assert.Equal(t, int64(5), data[5])
	assert.Equal(t, int64(9), data[6])
}

func TestSortUint32(t *testing.T) {
	data := []uint32{3, 5, 1, 9, 0, 2, 2}
	Uint32(data)
	assert.Equal(t, uint32(0), data[0])
	assert.Equal(t, uint32(1), data[1])
	assert.Equal(t, uint32(2), data[2])
	assert.Equal(t, uint32(2), data[3])
	assert.Equal(t, uint32(3), data[4])
	assert.Equal(t, uint32(5), data[5])
	assert.Equal(t, uint32(9), data[6])
}
