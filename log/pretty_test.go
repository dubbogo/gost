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

/* log_test.go - test for log.go */
package gxlog

import (
	"fmt"
	"testing"
)

type info struct {
	name string
	age  float32
	m    map[string]string
}

func TestPrettyString(t *testing.T) {
	i := info{name: "hello", age: 23.5, m: map[string]string{"h": "w", "hello": "world"}}
	fmt.Println(PrettyString(i))
}

func TestColorPrint(t *testing.T) {
	i := info{name: "hello", age: 23.5, m: map[string]string{"h": "w", "hello": "world"}}
	ColorPrintln(i)
}

func TestColorPrintf(t *testing.T) {
	i := info{name: "hello", age: 23.5, m: map[string]string{"h": "w", "hello": "world"}}
	ColorPrintf("exapmle format:%s\n", i)
}
