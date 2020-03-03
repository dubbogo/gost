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

package gxpage

import (
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestNewDefaultPage(t *testing.T) {
	data := make([]interface{}, 10)
	page := New(121, 10, data, 499)

	assert.Equal(t, 10, page.GetDataSize())
	assert.Equal(t, 121, page.GetOffset())
	assert.Equal(t, 10, page.GetPageSize())
	assert.Equal(t, 50, page.GetTotalPages())
	assert.Equal(t, data, page.GetData())
	assert.True(t, page.HasNext())
	assert.True(t, page.HasData())

	page = New(492, 10, data, 499)
	assert.False(t, page.HasNext())
}
