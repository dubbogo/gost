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

package gxxorlist

import (
	"fmt"
)

// OutputElem outputs a xorlist element.
func OutputElem(e *XorElement) {
	if e != nil {
		// fmt.Printf("addr:%p, value:%v", e, e)
		fmt.Printf("value:%v", e.Value)
	}
}

// OutputList iterates through list and print its contents.
func OutputList(l *XorList) {
	idx := 0
	for e, p := l.Front(); e != nil; p, e = e, e.Next(p) {
		fmt.Printf("idx:%v, ", idx)
		OutputElem(e)
		fmt.Printf("\n")
		idx++
	}
}

// OutputListR iterates through list and print its contents in reverse.
func OutputListR(l *XorList) {
	idx := 0
	for e, n := l.Back(); e != nil; e, n = e.Next(n), e {
		fmt.Printf("idx:%v, ", idx)
		OutputElem(e)
		fmt.Printf("\n")
		idx++
	}
}
