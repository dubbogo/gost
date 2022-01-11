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

package gxmath

import (
	"math"
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestAbs(t *testing.T) {
	mathTable := []struct {
		in, want int64
	}{
		{-0, 0},
		{1, 1},
		{-1, 1},
		{-3, 3},
	}

	for _, val := range mathTable {
		assert.Equal(t, val.want, AbsInt64(val.in))
		assert.Equal(t, int32(val.want), AbsInt32(int32(val.in)))
		assert.Equal(t, int16(val.want), AbsInt16(int16(val.in)))
		assert.Equal(t, int8(val.want), AbsInt8(int8(val.in)))
	}

	assert.Equal(t, int64(math.MaxInt64), AbsInt64(-math.MaxInt64))
	assert.Equal(t, int32(math.MaxInt32), AbsInt32(-math.MaxInt32))
	assert.Equal(t, int16(math.MaxInt16), AbsInt16(-math.MaxInt16))
	assert.Equal(t, int8(math.MaxInt8), AbsInt8(-math.MaxInt8))
}
