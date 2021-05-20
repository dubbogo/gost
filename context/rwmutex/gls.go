/*
 * MIT License
 *
 * Copyright (c) 2020 Mahendra Kanani
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

package rwmutex

import (
	"runtime"
	"sync"
	"unsafe"
)

import (
	"github.com/dubbogo/gost/context/goroutines"
)

const factor = 10

type Map struct {
	m map[unsafe.Pointer]interface{}
	sync.RWMutex
}

var (
	shards = runtime.GOMAXPROCS(-1) * 2 // will be executed at load time.
	// most of the time, number of core are 2^x, but can be different due to virtualisation/containerisation
	// bitwise AND can be used when shards is 2^x.
	// division = shards - 1
	glsArr = make([]*Map, shards)
)

func init() {
	for idx := range glsArr {
		glsArr[idx] = &Map{
			m: make(map[unsafe.Pointer]interface{}),
		}
	}
}

// Set accepts a value. key will be the current go-routine.
func Set(value interface{}) {
	curRtn := goroutines.CurRoutine()
	idx := int(uintptr(curRtn)>>factor) % shards
	glsArr[idx].Lock()
	glsArr[idx].m[curRtn] = value
	glsArr[idx].Unlock()
}

// Get returns a value present in map for calling go-routine
func Get() interface{} {
	curRtn := goroutines.CurRoutine()
	idx := int(uintptr(curRtn)>>factor) % shards
	glsArr[idx].RLock()
	defer glsArr[idx].RUnlock()
	val, ok := glsArr[idx].m[curRtn]
	if !ok {
		return nil
	}
	return val
}

// Del deletes the value and key from map.
func Del() {
	curRtn := goroutines.CurRoutine()
	idx := int(uintptr(curRtn)>>factor) % shards

	glsArr[idx].Lock()
	delete(glsArr[idx].m, curRtn)
	glsArr[idx].Unlock()
}
