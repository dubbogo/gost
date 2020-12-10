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

package gxfilepath

import (
	"testing"
)

import (
	"github.com/stretchr/testify/assert"
)

func TestFileExists(t *testing.T) {
	file := "./path_test.go"
	ok, err := Exists(file)
	assert.True(t, ok)
	assert.Nil(t, err)
	ok, err = DirExists(file)
	assert.False(t, ok)
	assert.NotNil(t, err)
	ok, err = FileExists(file)
	assert.True(t, ok)
	assert.Nil(t, err)

	file = "./path_test1.go"
	ok, err = Exists(file)
	assert.False(t, ok)
	assert.Nil(t, err)
	ok, err = DirExists(file)
	assert.False(t, ok)
	assert.NotNil(t, err)
	ok, err = FileExists(file)
	assert.False(t, ok)
	assert.NotNil(t, err)
}

func TestDirExists(t *testing.T) {
	file := "./"
	ok, err := Exists(file)
	assert.True(t, ok)
	assert.Nil(t, err)
	ok, err = DirExists(file)
	assert.True(t, ok)
	assert.Nil(t, err)
	ok, err = FileExists(file)
	assert.False(t, ok)
	assert.NotNil(t, err)

	file = "./go"
	ok, err = Exists(file)
	assert.False(t, ok)
	assert.Nil(t, err)
	ok, err = DirExists(file)
	assert.False(t, ok)
	assert.NotNil(t, err)
	ok, err = FileExists(file)
	assert.False(t, ok)
	assert.NotNil(t, err)
}
